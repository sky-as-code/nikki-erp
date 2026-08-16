*** Settings ***
Documentation     The stock figures a product page shows.
...
...               Every one of these is a read. Product displays what Stock reports and stores
...               none of it, so a product with no stock answers zero rather than 404, and no
...               call here can change a balance.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Product Variant Under Test
Test Tags         inventory    product_stock    summary


*** Test Cases ***
Variant Summary Answers For A Product With No Stock
    [Documentation]    A variant nothing has been done to still has an answer: zero. Reporting
    ...    "not found" would make a caller distinguish "no stock" from "no product", which are
    ...    different things and only one of them is an error.
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/variant_stock_summary
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Should Be Equal As Numbers    ${body}[onHand]    0
    Should Be Equal As Numbers    ${body}[reserved]    0
    Should Be Equal As Numbers    ${body}[available]    0
    Should Be Equal As Integers    ${body}[locationCount]    0

Available Is Derived, Never Stored
    [Documentation]    AC-PROD-INT-006. Available is on-hand minus reserved, computed on read,
    ...    so nothing can persist a value that disagrees with the two it comes from.
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/variant_stock_summary
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    ${expected}=    Evaluate    float($body['onHand']) - float($body['reserved'])
    Should Be Equal As Numbers    ${body}[available]    ${expected}

Batch Summary Answers A Whole Page In One Request
    [Documentation]    TS-PROD-15 and AC-PROD-INT-035. A product listing summarises its rows in
    ...    one call. The point of this endpoint is that a page of a hundred products costs one
    ...    request rather than a hundred, so it is asserted to return every id it was given.
    Ensure Second Variant For Summary
    ${ids}=    Create List    ${PRODUCT_VARIANT_ID}    ${SUMMARY_SECOND_VARIANT_ID}
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/variant_stock_summaries
    ...    json=${{ {'product_variant_ids': $ids} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Dictionary Should Contain Key    ${body}    ${PRODUCT_VARIANT_ID}
    Dictionary Should Contain Key    ${body}    ${SUMMARY_SECOND_VARIANT_ID}

Template Summary Aggregates Its Variants
    [Documentation]    TS-PROD-02 and AC-PROD-INT-011. A template holds no stock of its own: the
    ...    total is the sum of its variants, and the rows it was summed from come back with it so
    ...    a reader can see where the number came from.
    Ensure Second Variant For Summary
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/template_stock_summary
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Dictionary Should Contain Key    ${body}    summary
    Dictionary Should Contain Key    ${body}    variants
    ${count}=    Get Length    ${body}[variants]
    Should Be True    ${count} >= 2
    ...    msg=The breakdown must list every variant the total was summed from

Stock By Warehouse And By Location Are Readable
    [Documentation]    AC-PROD-INT-014 and AC-PROD-INT-015. Both answer for a product with no
    ...    stock too: an empty list, not an error.
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/stock_by_warehouse
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be True    isinstance($resp.json(), list)

    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/stock_by_location
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be True    isinstance($resp.json(), list)

Product Usage Reports Whether A Product Can Be Archived
    [Documentation]    The reader's own verdict travels with its numbers, so a UI explains a
    ...    refusal without restating the rule and risking restating it differently.
    ${resp}=    POST On Session    api    ${STOCK_QUANT_API}/product_usage
    ...    json=${{ {'product_variant_id': $PRODUCT_VARIANT_ID} }}
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    Dictionary Should Contain Key    ${body}    canArchive
    Should Be True    ${body}[canArchive]
    ...    msg=A variant with no stock and no open work can be archived


*** Keywords ***
Ensure Second Variant For Summary
    [Documentation]    A second variant of the same template, so the template aggregate has more
    ...    than one row to sum and the batch read has more than one id to resolve.
    ${id}=    Get Variable Value    ${SUMMARY_SECOND_VARIANT_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Product Variant Under Test
    ${key}=    Unique Code    summary
    ${id}    ${etag}=    Create Product Variant    ${PRODUCT_TEMPLATE_ID}    ${key}
    Set Global Variable    ${SUMMARY_SECOND_VARIANT_ID}    ${id}
