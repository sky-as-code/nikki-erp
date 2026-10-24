*** Settings ***
Documentation     Variant generation (BR §8.2, AC-PROD-011, CR01 AC08). The 2x2 fixture is the
...               canonical case: two INSTANT attributes of two values each must yield exactly four
...               variants. Saving the configuration reconciles at once, so the four exist before
...               anyone presses generate; `generate_variants` re-runs the same reconcile against
...               the current configuration and must then create nothing.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Two By Two Template
Test Tags         inventory    products    generate


*** Test Cases ***
Saving The Configuration Creates The Full Combination Set
    [Documentation]    BR §8.2 / AC-PROD-011 / AC08 / ADD11: 2 values x 2 values = 4 complete
    ...    variants and no base. A count of 2 would mean the attributes were treated as
    ...    alternatives; 1 would mean the key collapsed. The template's own empty default was
    ...    unused, so it is gone rather than archived.
    ${variants}=    Search Template Variants    ${MATRIX_TEMPLATE_ID}    include_archived=${True}
    Length Should Be    ${variants}    4
    ${bases}=    Variants Of Kind    ${variants}    base
    Should Be Empty    ${bases}    msg=An instant-only template pre-creates no base (ADD11)
    ${counts}=    Count Template Variants    ${MATRIX_TEMPLATE_ID}
    Should Be Equal As Integers    ${counts}[complete]    4
    Should Be Equal As Integers    ${counts}[base]    0
    Should Be Equal As Integers    ${counts}[possible]    4

Generated Variants Have Distinct Combinations
    [Documentation]    BR-PROD-VAR-002: four different products means four different keys.
    ${variants}=    Search Template Variants    ${MATRIX_TEMPLATE_ID}
    ${keys}=    Evaluate    sorted({item['combination_key'] for item in $variants})
    Length Should Be    ${keys}    4
    ...    msg=Each generated variant must carry its own combination key
    Set Suite Variable    ${MATRIX_VARIANT_IDS}    ${{ sorted(v['id'] for v in $variants) }}

Generate Again Creates Nothing
    [Documentation]    BR §8.5: generation reconciles rather than appends. Re-running it reuses
    ...    the four existing variants and reports them unchanged.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${MATRIX_TEMPLATE_ID}/generate_variants
    ...    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[created_variant_ids]    0
    ...    msg=Re-generating must create no duplicates (BR 8.5)
    Should Be Equal As Integers    ${resp.json()}[unchanged_count]    4

Reordering The Lines Keeps Ids And Keys
    [Documentation]    AC08 / ADD08: the combination key sorts by attribute id, so sending the
    ...    same configuration in another order and sequence changes no identity.
    ${config}=    Read Template Configuration    ${MATRIX_TEMPLATE_ID}
    ${lines}=    Evaluate    [{'attribute_id': l['attribute_id'], 'value_ids': list(reversed(l['value_ids'])), 'sequence': 10 - i} for i, l in enumerate(reversed($config['attribute_lines']))]
    Configure Template Successfully    ${MATRIX_TEMPLATE_ID}    ${lines}
    ${variants}=    Search Template Variants    ${MATRIX_TEMPLATE_ID}
    ${ids}=    Evaluate    sorted(v['id'] for v in $variants)
    Lists Should Be Equal    ${ids}    ${MATRIX_VARIANT_IDS}

Generate On A Template Without Attributes Keeps Its One Default
    [Documentation]    AC04 / BR §4.5: a template with no generating attribute gets exactly one
    ...    variant with the empty combination when it is created; generate finds it and
    ...    creates nothing more.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Attributeless Template
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}/generate_variants
    ...    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[created_variant_ids]    0
    ${variants}=    Search Template Variants    ${template_id}
    Length Should Be    ${variants}    1
    Should Be Equal    ${variants}[0][combination_key]    ${EMPTY}
    Should Be Equal    ${variants}[0][variant_kind]    complete
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Generate With Not Found Template Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}/generate_variants
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Generate With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/not-existing-1234567890123/generate_variants
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
