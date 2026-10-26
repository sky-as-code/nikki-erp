*** Settings ***
Documentation     Template delete and duplicate (D7, CR01 §11, PROD-021): a delete removes the unused
...               variants in the same transaction and is refused while any variant is used; a
...               duplicate gets new ids and links over the same catalogue, its own clone of the
...               preview image, and never stock, history or SKU/barcode.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Duplicate Fixture
Test Tags         inventory    product_configuration    duplicate


*** Variables ***
${STORAGE_UPLOAD_API}    /v1/itstorage/files/upload


*** Test Cases ***
Duplicate Carries The Active Configuration
    [Documentation]    AC29: new template id, new lines over the same catalogue attributes and
    ...    values, exclusions carried over, new variant ids, no SKU. The name gets a copy suffix
    ...    when none is given.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${SOURCE_ID}/duplicate
    ...    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${duplicate_id}=    Set Variable    ${resp.json()}[id]
    Should Not Be Equal    ${duplicate_id}    ${SOURCE_ID}
    ${source_config}=    Read Template Configuration    ${SOURCE_ID}
    ${duplicate_config}=    Read Template Configuration    ${duplicate_id}
    ${source_lines}=    Evaluate    sorted((l['attribute_id'], tuple(sorted(l['value_ids']))) for l in $source_config['attribute_lines'])
    ${duplicate_lines}=    Evaluate    sorted((l['attribute_id'], tuple(sorted(l['value_ids']))) for l in $duplicate_config['attribute_lines'])
    Should Be Equal    ${duplicate_lines}    ${source_lines}
    ${line_ids}=    Evaluate    {l['id'] for l in $source_config['attribute_lines']} & {l['id'] for l in $duplicate_config['attribute_lines']}
    Should Be Empty    ${line_ids}    msg=A duplicate owns new line records
    Length Should Be    ${duplicate_config}[exclusions]    1
    ${source_variants}=    Search Template Variants    ${SOURCE_ID}
    ${duplicate_variants}=    Search Template Variants    ${duplicate_id}
    Length Should Be    ${duplicate_variants}    ${{ len($source_variants) }}
    ${shared}=    Evaluate    {v['id'] for v in $source_variants} & {v['id'] for v in $duplicate_variants}
    Should Be Empty    ${shared}
    FOR    ${variant}    IN    @{duplicate_variants}
        Should Be True    ${{ not $variant.get('sku') }}    msg=SKU must not be duplicated
    END
    ${duplicate}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${duplicate_id}    id,name
    Should Contain    ${duplicate}[name][en-US]    (copy)
    Set Suite Variable    ${DUPLICATE_ID}    ${duplicate_id}

Duplicate Clones The Preview Image
    [Documentation]    A file is never shared: the duplicate references its own clone of the
    ...    source's image, a new file id, while the source keeps its own.
    ${duplicate}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${DUPLICATE_ID}    id,preview_image_file_id
    ${source}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${SOURCE_ID}    id,preview_image_file_id
    Should Be Equal    ${source}[preview_image_file_id]    ${SOURCE_IMAGE_ID}
    Should Not Be Empty    ${duplicate}[preview_image_file_id]
    Should Not Be Equal    ${duplicate}[preview_image_file_id]    ${SOURCE_IMAGE_ID}

Duplicate With A Name Uses It
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${SOURCE_ID}/duplicate
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': {'en-US': 'Robot Named Duplicate'}} }}
    Response Status Should Be    ${resp}    200
    ${duplicate}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${resp.json()}[id]    id,name
    Should Be Equal    ${duplicate}[name][en-US]    Robot Named Duplicate

Duplicate Of A Missing Template Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}/duplicate
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

The Old Copy Route Is Gone
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${SOURCE_ID}/copy
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} in (404, 405)

Delete Is Refused While A Variant Is Used
    [Tags]    negative
    ${variants}=    Search Template Variants    ${DUPLICATE_ID}
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${variants}[0][id]
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${DUPLICATE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Template Has Variants Error    ${resp}
    [Teardown]    Release Stock Move    ${transfer_id}    ${move_id}

Delete Removes A Template With Only Unused Variants
    [Documentation]    D7: lines, exclusions and the unused variants go with the template.
    ${variants}=    Search Template Variants    ${DUPLICATE_ID}
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${DUPLICATE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${variants}[0][id]
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}


*** Keywords ***
Ensure Duplicate Fixture
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Duplicate Color
    ${size}    ${sizes}=    Create Attribute With Values    instant    s    m    words=Robot Duplicate Size
    ${lines}=    Create List
    ...    ${{ {'attribute_id': $color, 'value_ids': list($colors.values())} }}
    ...    ${{ {'attribute_id': $size, 'value_ids': list($sizes.values())} }}
    ${excl}=    Exclusion    ${colors}[red]    ${sizes}[s]
    ${template_id}=    Create Configured Template    ${lines}    ${{ [$excl] }}    words=Robot Duplicate Source
    ${variants}=    Search Template Variants    ${template_id}
    ${sku}=    Unique Code    DUPSKU
    ${resp}=    Patch Record    ${PRODUCT_VARIANT_API}    ${variants}[0][id]    ${{ {'sku': $sku} }}
    Response Status Should Be    ${resp}    200
    ${image_id}=    Upload Org Image
    ${resp}=    Patch Record    ${PRODUCT_TEMPLATE_API}    ${template_id}    ${{ {'preview_image_file_id': $image_id} }}
    Response Status Should Be    ${resp}    200
    Set Suite Variable    ${SOURCE_ID}    ${template_id}
    Set Suite Variable    ${SOURCE_IMAGE_ID}    ${image_id}

Upload Org Image
    [Documentation]    Uploads a PNG into the inventory org's space and returns its pending file
    ...    id. The text parts precede the file part, as the upload API requires.
    ${content}=    Evaluate    b'\\x89PNG\\r\\n\\x1a\\n' + b'\\x00' * 248
    ${resp}=    POST On Session    api    ${STORAGE_UPLOAD_API}
    ...    data=${{ {'org_id': $INV_ORG_ID, 'name': 'preview.png', 'mime_type': 'image/png', 'size_bytes': str(len($content))} }}
    ...    files=${{ {'file': ('preview.png', $content, 'image/png')} }}
    Should Be True    ${resp.status_code} in (200, 201)    msg=upload failed: ${resp.text}
    RETURN    ${resp.json()}[id]
