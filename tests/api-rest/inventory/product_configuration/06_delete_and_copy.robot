*** Settings ***
Documentation     Template delete and copy (D7, CR01 §11, PROD-021): a delete removes the unused
...               variants in the same transaction and is refused while any variant is used; a copy
...               gets new ids and links over the same catalogue, and never stock, history or
...               SKU/barcode.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Copy Fixture
Test Tags         inventory    product_configuration    copy


*** Test Cases ***
Copy Duplicates The Active Configuration
    [Documentation]    AC29: new template id, new lines over the same catalogue attributes and
    ...    values, exclusions carried over, new variant ids, no SKU. The name gets a copy suffix
    ...    when none is given.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${SOURCE_ID}/copy
    ...    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${copy_id}=    Set Variable    ${resp.json()}[id]
    Should Not Be Equal    ${copy_id}    ${SOURCE_ID}
    ${source_config}=    Read Template Configuration    ${SOURCE_ID}
    ${copy_config}=    Read Template Configuration    ${copy_id}
    ${source_lines}=    Evaluate    sorted((l['attribute_id'], tuple(sorted(l['value_ids']))) for l in $source_config['attribute_lines'])
    ${copy_lines}=    Evaluate    sorted((l['attribute_id'], tuple(sorted(l['value_ids']))) for l in $copy_config['attribute_lines'])
    Should Be Equal    ${copy_lines}    ${source_lines}
    ${line_ids}=    Evaluate    {l['id'] for l in $source_config['attribute_lines']} & {l['id'] for l in $copy_config['attribute_lines']}
    Should Be Empty    ${line_ids}    msg=A copy owns new line records
    Length Should Be    ${copy_config}[exclusions]    1
    ${source_variants}=    Search Template Variants    ${SOURCE_ID}
    ${copy_variants}=    Search Template Variants    ${copy_id}
    Length Should Be    ${copy_variants}    ${{ len($source_variants) }}
    ${shared}=    Evaluate    {v['id'] for v in $source_variants} & {v['id'] for v in $copy_variants}
    Should Be Empty    ${shared}
    FOR    ${variant}    IN    @{copy_variants}
        Should Be True    ${{ not $variant.get('sku') }}    msg=SKU must not be copied
    END
    ${copy}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${copy_id}    id,name
    Should Contain    ${copy}[name][en-US]    (copy)
    Set Suite Variable    ${COPY_ID}    ${copy_id}

Copy With A Name Uses It
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${SOURCE_ID}/copy
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': {'en-US': 'Robot Named Copy'}} }}
    Response Status Should Be    ${resp}    200
    ${copy}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${resp.json()}[id]    id,name
    Should Be Equal    ${copy}[name][en-US]    Robot Named Copy

Copy Of A Missing Template Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}/copy
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete Is Refused While A Variant Is Used
    [Tags]    negative
    ${variants}=    Search Template Variants    ${COPY_ID}
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${variants}[0][id]
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${COPY_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Template Has Variants Error    ${resp}
    [Teardown]    Release Stock Move    ${transfer_id}    ${move_id}

Delete Removes A Template With Only Unused Variants
    [Documentation]    D7: lines, exclusions and the unused variants go with the template.
    ${variants}=    Search Template Variants    ${COPY_ID}
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${COPY_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${variants}[0][id]
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}


*** Keywords ***
Ensure Copy Fixture
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Copy Color
    ${size}    ${sizes}=    Create Attribute With Values    instant    s    m    words=Robot Copy Size
    ${lines}=    Create List
    ...    ${{ {'attribute_id': $color, 'value_ids': list($colors.values())} }}
    ...    ${{ {'attribute_id': $size, 'value_ids': list($sizes.values())} }}
    ${excl}=    Exclusion    ${colors}[red]    ${sizes}[s]
    ${template_id}=    Create Configured Template    ${lines}    ${{ [$excl] }}    words=Robot Copy Source
    ${variants}=    Search Template Variants    ${template_id}
    ${sku}=    Unique Code    COPYSKU
    ${resp}=    Patch Record    ${PRODUCT_VARIANT_API}    ${variants}[0][id]    ${{ {'sku': $sku} }}
    Response Status Should Be    ${resp}    200
    Set Suite Variable    ${SOURCE_ID}    ${template_id}
