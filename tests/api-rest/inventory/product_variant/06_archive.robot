*** Settings ***
Documentation     Archiving Product Variants, including the availability sync of
...               BR-PROD-VAR-006 / BR-PROD-VAR-007 / AC-PROD-020: archiving the last
...               non-archived variant archives its template, because a product line with
...               nothing selectable must not keep appearing as available. The sync cases run
...               on their own template, since both records end up archived.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Variant Under Test
Test Tags         inventory    product_variant    archive


*** Test Cases ***
Archive Succeeds
    [Documentation]    Runs against a template that still has other variants (05_exists
    ...    seeded 50 of them), so this exercises the plain archive without tripping the
    ...    last-variant sync below.
    Ensure Seeded Product Variants    50
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $PRODUCT_VARIANT_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Archiving Stamps The User Source
    [Documentation]    BR §8.9: a variant archived directly is stamped `user`, not
    ...    `template_cascade`. That is what lets a later template unarchive restore only the
    ...    variants it cascaded to and leave this one deliberately archived.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[is_archived]    ${True}
    Should Be Equal    ${item}[archive_source]    user
    ...    msg=A directly archived variant must be stamped user (BR 8.9)
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${item}[etag]

Archived Variant Is Not Selectable
    [Documentation]    AC-PROD-019/020: is_selectable on the effective product is the single
    ...    answer to "may a transaction line reference this?", so no consumer has to re-apply
    ...    the archive and status rules itself.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/effective
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${data}=    Set Variable    ${resp.json()}
    Validate Json Schema    ${data}    ${INVENTORY_SCHEMA_DIR}/effective_product.json
    Should Be Equal    ${data}[is_variant_archived]    ${True}
    Should Be Equal    ${data}[is_selectable]    ${False}
    ...    msg=An archived variant must not be selectable

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $PRODUCT_VARIANT_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Archiving The Last Variant Archives Its Template
    [Documentation]    BR-PROD-VAR-006 / PROD-018: with its only variant archived, a template that
    ...    has no dynamic and no never attribute has nothing transactable left, so it is archived
    ...    too. Uses a fresh template, whose only variant is the empty default. PFU-007: the
    ...    variant that caused it is stamped template_cascade, so the template's unarchive brings
    ...    it back.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Last Variant Template
    ${variants}=    Search Template Variants    ${template_id}
    Length Should Be    ${variants}    1
    ${variant_id}=    Set Variable    ${variants}[0][id]
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${variant_id}
    Response Status Should Be    ${resp}    200
    ${item}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${template_id}    id,is_archived
    Should Be Equal    ${item}[is_archived]    ${True}
    ...    msg=Archiving the last active variant must archive its template (BR-PROD-VAR-006)
    ${variant}=    Read Record    ${PRODUCT_VARIANT_API}    ${variant_id}    id,is_archived,archive_source
    Should Be Equal    ${variant}[archive_source]    template_cascade
    ...    msg=The archive that took the template down is part of its cascade (PFU-007)
    Set Suite Variable    ${SYNC_TEMPLATE_ID}    ${template_id}
    Set Suite Variable    ${SYNC_VARIANT_ID}    ${variant_id}

Unarchiving A Variant Of An Archived Template Fails
    [Documentation]    D3 / CR01 §7.4: a variant is not brought back on its own while its
    ...    template is archived; the template is unarchived first, which restores what the
    ...    cascade archived. PFU-007: that includes the last variant whose archive archived the
    ...    template, so the restored template has something to sell again.
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${SYNC_VARIANT_ID}    ${False}
    Response Should Contain Business Error    ${resp}    inventory:product_variant.template_archived    is_archived
    ${resp}=    Set Record Archived    ${PRODUCT_TEMPLATE_API}    ${SYNC_TEMPLATE_ID}    ${False}
    Response Status Should Be    ${resp}    200
    ${variant}=    Read Record    ${PRODUCT_VARIANT_API}    ${SYNC_VARIANT_ID}    id,is_archived,archive_source
    Should Be Equal    ${variant}[is_archived]    ${False}
    Should Be True    ${{ not $variant.get('archive_source') }}
    ...    msg=A restored variant must carry no archive_source stamp
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${SYNC_TEMPLATE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Archiving The Last Complete Of A Dynamic Template Keeps The Template
    [Documentation]    PROD-018: a template with a dynamic attribute can still resolve new
    ...    completes, so archiving its only complete does not archive it.
    ${attr}    ${values}=    Create Attribute With Values    dynamic    s    words=Robot Last Dynamic Size
    ${line}=    Template Line    ${attr}    ${values}[s]
    ${template_id}=    Create Configured Template    ${{ [$line] }}    words=Robot Last Dynamic Template
    ${resp}=    Resolve Selection    ${template_id}    ${{ [{'attribute_id': $attr, 'value_id': $values['s']}] }}
    Response Status Should Be    ${resp}    200
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${resp.json()}[variant_id]
    Response Status Should Be    ${resp}    200
    ${item}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${template_id}    id,is_archived
    Should Be Equal    ${item}[is_archived]    ${False}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $PRODUCT_VARIANT_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
