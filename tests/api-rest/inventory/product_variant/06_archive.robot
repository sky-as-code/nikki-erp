*** Settings ***
Documentation     Archiving Product Variants, including the availability sync of
...               BR-PROD-VAR-006 / BR-PROD-VAR-007 / AC-PROD-020: archiving the last
...               non-archived variant archives its template, because a product line with
...               nothing selectable must not keep appearing as available. The sync cases run
...               on their own template, since both records end up archived.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Variant Under Test
Test Tags         inventory    product_variant    archive


*** Test Cases ***
Archive Succeeds
    [Documentation]    Runs against a template that still has other variants (05_exists
    ...    seeded 50 of them), so this exercises the plain archive without tripping the
    ...    last-variant sync below.
    Ensure Seeded Product Variants    50
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_VARIANT_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Archiving Stamps The User Source
    [Documentation]    BR §8.9: a variant archived directly is stamped `user`, not
    ...    `template_cascade`. That is what lets a later template unarchive restore only the
    ...    variants it cascaded to and leave this one deliberately archived.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
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
    Response Status Should Be    ${resp}    200
    ${data}=    Set Variable    ${resp.json()}
    Validate Json Schema    ${data}    ${INVENTORY_SCHEMA_DIR}/effective_product.json
    Should Be Equal    ${data}[is_variant_archived]    ${True}
    Should Be Equal    ${data}[is_selectable]    ${False}
    ...    msg=An archived variant must not be selectable

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_VARIANT_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Archiving The Last Variant Archives Its Template
    [Documentation]    BR-PROD-VAR-006 / AC-PROD-020: with its only variant archived the
    ...    template has nothing transactable left, so leaving it live would advertise a
    ...    product that cannot be bought. Uses a dedicated one-variant template.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Last Variant Template
    ${key}=    Unique Code    lastcomb
    ${variant_id}    ${variant_etag}=    Create Product Variant    ${template_id}    ${key}
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${variant_id}/archived
    ...    json=${{ {'etag': $variant_etag, 'is_archived': True} }}
    ${variant_etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${variant_etag}

    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be Equal    ${item}[is_archived]    ${True}
    ...    msg=Archiving the last active variant must archive its template (BR-PROD-VAR-006)
    Set Suite Variable    ${SYNC_TEMPLATE_ID}    ${template_id}
    Set Suite Variable    ${SYNC_VARIANT_ID}    ${variant_id}
    Set Suite Variable    ${SYNC_VARIANT_ETAG}    ${variant_etag}

Unarchiving A Variant Restores Its Template
    [Documentation]    BR-PROD-VAR-007: the symmetry of the rule above. Bringing a variant
    ...    back makes the product line transactable again, so the template returns with it.
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${SYNC_VARIANT_ID}/archived
    ...    json=${{ {'etag': $SYNC_VARIANT_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${SYNC_VARIANT_ETAG}

    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${SYNC_TEMPLATE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be Equal    ${item}[is_archived]    ${False}
    ...    msg=Unarchiving a variant must restore its template (BR-PROD-VAR-007)
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${SYNC_VARIANT_ID}    expected_status=any
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${SYNC_TEMPLATE_ID}    expected_status=any

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_VARIANT_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
