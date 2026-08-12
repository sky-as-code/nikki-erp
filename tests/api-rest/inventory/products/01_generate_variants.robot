*** Settings ***
Documentation     Variant generation (BR §8.2, AC-PROD-011). The 2x2 fixture is the canonical
...               case: two INSTANT attributes of two values each must produce exactly four
...               variants — the cartesian product, no more and no fewer. Generation is
...               idempotent, so re-running it creates nothing and reports the existing four
...               as unchanged.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Two By Two Template
Test Tags         inventory    products    generate


*** Test Cases ***
Generate Creates The Full Combination Set
    [Documentation]    BR §8.2 / AC-PROD-011: 2 values x 2 values = 4 variants. A count of 2
    ...    would mean the attributes were treated as alternatives rather than combined; a
    ...    count of 1 would mean the combination key collapsed.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${MATRIX_TEMPLATE_ID}/generate_variants
    ...    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${created}=    Set Variable    ${resp.json()}[created_variant_ids]
    Length Should Be    ${created}    4
    ...    msg=A 2x2 INSTANT template must generate exactly four variants (AC-PROD-011)

Generated Variants Have Distinct Combinations
    [Documentation]    BR-PROD-VAR-002 is what makes the four variants four *different*
    ...    products. Equal combination keys would mean the generator built the same identity
    ...    four times and only the composite unique stopped it.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 50, 'graph': '{"if":["product_template_id", "=", "' + $MATRIX_TEMPLATE_ID + '"]}'} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Length Should Be    ${items}    4
    ${keys}=    Evaluate    sorted({item['combination_key'] for item in $items})
    Length Should Be    ${keys}    4
    ...    msg=Each generated variant must carry its own combination key

Generate Again Creates Nothing
    [Documentation]    BR §8.5: generation reconciles rather than appends. Re-running it must
    ...    reuse the existing variants — creating four more would double every SKU each time
    ...    an admin pressed the button.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${MATRIX_TEMPLATE_ID}/generate_variants
    ...    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${created}=    Set Variable    ${resp.json()}[created_variant_ids]
    Length Should Be    ${created}    0
    ...    msg=Re-generating must create no duplicates (BR 8.5)
    Should Be Equal As Integers    ${resp.json()}[unchanged_count]    4
    ...    msg=The four existing variants must be reported as unchanged, not obsolete

Generate On A Template Without Attributes Creates One Variant
    [Documentation]    BR §4.5 / AC-PROD-008: a template with no variant-generating attributes
    ...    still gets exactly one variant, with an empty combination. Without it the product
    ...    would have nothing transactable, and there is deliberately no is_default_variant
    ...    flag marking it (BR-PROD-VAR-005).
    ${template_id}    ${template_etag}=    Create Product Template    Robot Attributeless Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}/generate_variants
    ...    json=${{ {} }}
    Response Status Should Be    ${resp}    200
    ${created}=    Set Variable    ${resp.json()}[created_variant_ids]
    Length Should Be    ${created}    1
    ...    msg=An attributeless template must still get exactly one variant (AC-PROD-008)
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${created}[0]
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[combination_key]    ${EMPTY}
    ...    msg=The single variant of an attributeless template carries the empty combination
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${created}[0]    expected_status=any
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}    expected_status=any

Generate With Not Found Template Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}/generate_variants
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Generate With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/not-existing-1234567890123/generate_variants
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
