*** Settings ***
Documentation     Whether a product may be withdrawn from the working set.
...
...               The line drawn here is between stock that is still live and stock that is
...               merely remembered. A product with goods on the shelf cannot be archived;
...               one whose only trace is completed movement can, and the records keep
...               resolving it afterwards.
...
...               Archiving never tidies up on the way through: it does not unreserve, cancel,
...               zero or scrap anything. A refused archive leaves the stock exactly as it was.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Guarded Variant With Stock
Test Tags         inventory    product_stock    archive_guard


*** Test Cases ***
A Variant With Stock Cannot Be Archived
    [Documentation]    TS-PROD-10 and AC-PROD-INT-028. The product still has goods on hand, so
    ...    the archive is refused rather than quietly stranding them.
    [Tags]    negative
    ${item}=    Get Guarded Variant
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_VARIANT_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=A product still holding stock must not be archivable

The Refused Archive Changed Nothing
    [Documentation]    AC-PROD-INT-033. Archiving must never generate a movement, an adjustment
    ...    or a scrap to make the stock go away, so the balance is unchanged by the refusal.
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/product_usage
    ...    json=${{ {'product_variant_id': $GUARDED_VARIANT_ID} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Be True    float($body['onHandQuantity']) > 0
    ...    msg=The stock is still there; the refusal did not dispose of it
    Should Not Be True    ${body}[canArchive]

    ${item}=    Get Guarded Variant
    Should Not Be True    ${item}[is_archived]
    ...    msg=The variant is still in the working set

A Template Is Refused While Any Variant Holds Stock
    [Documentation]    TS-PROD-12 and AC-PROD-INT-032. The product line is archived as a unit or
    ...    not at all, so one variant with stock blocks the whole template.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${GUARDED_TEMPLATE_ID}
    Response Status Should Be    ${resp}    200
    ${template}=    Set Variable    ${resp.json()}
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${GUARDED_TEMPLATE_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $template['etag']} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=One variant with stock blocks archiving the product line

No Variant Was Archived By The Refused Template Archive
    [Documentation]    The guard runs over every variant before anything is written. Were it to
    ...    check inside the cascade instead, the variants it had already reached would be left
    ...    archived by an operation that was then refused — half a product line withdrawn.
    ${item}=    Get Guarded Variant
    Should Not Be True    ${item}[is_archived]
    ...    msg=A refused template archive must leave every variant untouched

    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_CLEAN_VARIANT_ID}
    Response Status Should Be    ${resp}    200
    Should Not Be True    ${resp.json()}[is_archived]
    ...    msg=The clean variant must not be archived by an operation that was rejected

A Variant With No Stock Can Be Archived
    [Documentation]    TS-PROD-11 and AC-PROD-INT-031. History does not block: a variant holding
    ...    nothing archives, and completed records keep resolving it.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_CLEAN_VARIANT_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_CLEAN_VARIANT_ID}/archived
    ...    json=${{ {'is_archived': True, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200

    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_CLEAN_VARIANT_ID}
    Response Status Should Be    ${resp}    200
    Should Be True    ${resp.json()}[is_archived]

Unarchiving Is Never Blocked By Stock
    [Documentation]    Restoring a product to the working set strands nothing. Guarding it would
    ...    make a variant archived by mistake impossible to recover.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_CLEAN_VARIANT_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}
    ${resp}=    POST On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_CLEAN_VARIANT_ID}/archived
    ...    json=${{ {'is_archived': False, 'etag': $item['etag']} }}
    Response Status Should Be    ${resp}    200


*** Keywords ***
Get Guarded Variant
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${GUARDED_VARIANT_ID}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}

Ensure Guarded Variant With Stock
    [Documentation]    A template with two variants, one holding stock and one clean, which is
    ...    what makes the template guard's "all before any" behaviour observable.
    ...
    ...    The stock is put there by receiving it, never by writing a quant directly: the engine
    ...    refuses client writes to a balance, and a balance with no movement behind it is
    ...    exactly what that rule exists to prevent.
    ${existing}=    Get Variable Value    ${GUARDED_VARIANT_ID}    ${EMPTY}
    IF    $existing    RETURN

    ${template_id}    ${template_etag}=    Create Product Template    Guarded Template
    Set Global Variable    ${GUARDED_TEMPLATE_ID}    ${template_id}

    ${stocked_key}=    Unique Code    guarded
    ${stocked_id}    ${stocked_etag}=    Create Product Variant    ${template_id}    ${stocked_key}
    Set Global Variable    ${GUARDED_VARIANT_ID}    ${stocked_id}

    ${clean_key}=    Unique Code    clean
    ${clean_id}    ${clean_etag}=    Create Product Variant    ${template_id}    ${clean_key}
    Set Global Variable    ${GUARDED_CLEAN_VARIANT_ID}    ${clean_id}

    Receive Stock For Guarded Variant    ${stocked_id}

Receive Stock For Guarded Variant
    [Documentation]    Brings goods in through a receipt and validates it, so the balance exists
    ...    because a movement completed against it.
    [Arguments]    ${variant_id}
    Ensure Supplier Location Under Test
    Ensure Inventory Location Under Test
    Ensure Receipt Operation Type Under Test

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}
    ...    json=${{ {'operation_type_id': $RECEIPT_OPERATION_TYPE_ID, 'source_location_id': $SUPPLIER_LOCATION_ID, 'destination_location_id': $INVENTORY_LOCATION_ID, 'scheduled_at': '2026-01-01T00:00:00Z', 'org_id': $INV_ORG_ID} }}
    ${transfer_id}    ${transfer_etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${GUARDED_TRANSFER_ID}    ${transfer_id}

    ${resp}=    POST On Session    api    ${STOCK_MOVE_API}
    ...    json=${{ {'transfer_id': $transfer_id, 'sequence': 1, 'product_variant_id': $variant_id, 'uom_id': $UOM_ID, 'demand_quantity': '10', 'base_demand_quantity': '10', 'source_location_id': $SUPPLIER_LOCATION_ID, 'destination_location_id': $INVENTORY_LOCATION_ID, 'org_id': $INV_ORG_ID} }}
    ${move_id}    ${move_etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${GUARDED_MOVE_ID}    ${move_id}

    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/confirm
    ...    json=${{ {} }}    expected_status=any
    ${resp}=    POST On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}/validate
    ...    json=${{ {} }}    expected_status=any
    Should Be True    ${resp.status_code} < 400
    ...    msg=The fixture needs the receipt to complete, or there is no stock to guard against
