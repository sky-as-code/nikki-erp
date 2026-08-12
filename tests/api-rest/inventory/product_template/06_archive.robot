*** Settings ***
Documentation     Archiving the Product Template under test, and the cascade of
...               BR-PROD-TPL-002 / AC-PROD-019: archiving a template must take its variants
...               with it, or a product line would remain selectable through SKUs whose
...               catalog entry is gone. The cascade is exercised on a dedicated template
...               and variant, since it is destructive to both.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
Test Tags         inventory    product_template    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_TEMPLATE_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Archiving Does Not Change Status
    [Documentation]    BR-PROD-TPL-004 / AC-PROD-018 from the other direction: 03_update
    ...    proved discontinuing does not archive; this proves archiving does not discontinue.
    ...    Both states stay independently visible.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be Equal    ${item}[is_archived]    ${True}
    Should Be Equal    ${item}[status]    active
    ...    msg=Archiving a template must leave its business status untouched
    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${item}[etag]

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_TEMPLATE_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Archive Cascades To Variants
    [Documentation]    BR-PROD-TPL-002 / AC-PROD-019: every variant of an archived template
    ...    becomes non-selectable. The cascade stamps archive_source=template_cascade, which
    ...    is what lets the unarchive below tell a cascaded variant from one a user archived
    ...    deliberately (BR §8.9). Runs on its own template, because both records end up
    ...    archived and the shared fixtures must stay live for the later suites.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Cascade Template
    ${key}=    Unique Code    cascade
    ${variant_id}    ${variant_etag}=    Create Product Variant    ${template_id}    ${key}
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}/archived
    ...    json=${{ {'etag': $template_etag, 'is_archived': True} }}
    ${template_etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${template_etag}

    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${variant_id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[is_archived]    ${True}
    ...    msg=Archiving a template must archive its variants (BR-PROD-TPL-002)
    Should Be Equal    ${item}[archive_source]    template_cascade
    ...    msg=A cascaded archive must be stamped so unarchiving can tell it from a deliberate one
    Set Suite Variable    ${CASCADE_TEMPLATE_ID}    ${template_id}
    Set Suite Variable    ${CASCADE_TEMPLATE_ETAG}    ${template_etag}
    Set Suite Variable    ${CASCADE_VARIANT_ID}    ${variant_id}

Unarchive Cascades Back To Variants
    [Documentation]    BR §8.9: unarchiving restores only the variants the cascade archived.
    ...    The archive_source stamp is cleared with them, so a later user-archive is not
    ...    mistaken for a cascade.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${CASCADE_TEMPLATE_ID}/archived
    ...    json=${{ {'etag': $CASCADE_TEMPLATE_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${CASCADE_TEMPLATE_ETAG}

    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${CASCADE_VARIANT_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[is_archived]    ${False}
    ...    msg=Unarchiving a template must restore the variants it cascaded to
    Should Be True    ${{ not $item.get('archive_source') }}
    ...    msg=Restoring a cascaded variant must clear its archive_source stamp
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${CASCADE_VARIANT_ID}    expected_status=any
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${CASCADE_TEMPLATE_ID}    expected_status=any

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $PRODUCT_TEMPLATE_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
