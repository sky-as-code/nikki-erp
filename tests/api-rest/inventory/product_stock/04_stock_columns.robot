*** Settings ***
Documentation     Stock figures as columns on the product list.
...
...               A product listing can show what each variant holds without Product storing any
...               of it: the figures are computed fields on the variant, each one a correlated
...               subquery over its stock balances, so the number is Stock's and is read fresh.
...
...               They are opt-in. A list that does not name them must not pay for the
...               subqueries, which is what makes them optional columns rather than a cost on
...               every product read.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Product Variant Under Test
Test Tags         inventory    product_stock    stock_columns


*** Test Cases ***
The Stock Columns Are Offered By The Schema
    [Documentation]    The column picker builds itself from meta/schema, so the fields have to be
    ...    advertised there — and advertised as computed, so nothing offers them as inputs.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]

    FOR    ${name}    IN    on_hand_quantity    reserved_quantity    available_quantity
        Dictionary Should Contain Key    ${fields}    ${name}
        Should Be True    ${fields}[${name}][is_computed]
        ...    msg=${name} must be advertised as computed
        Should Not Be True    ${fields}[${name}][is_persisted]
        ...    msg=${name} must not claim a column
        Should Not Be True    ${fields}[${name}][is_system_field]
        ...    msg=${name} carries business meaning, so it belongs in the column picker
    END

A Variant With No Stock Reads Zero, Not Null
    [Documentation]    AC-PROD-INT-004. SUM over no rows is NULL in SQL; the declared default is
    ...    what turns that into the zero a product page should show.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=fields=id,on_hand_quantity,reserved_quantity,available_quantity
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}

    Should Be Equal As Numbers    ${item}[on_hand_quantity]    0
    Should Be Equal As Numbers    ${item}[reserved_quantity]    0
    Should Be Equal As Numbers    ${item}[available_quantity]    0

Available Agrees With The Two Figures It Derives From
    [Documentation]    available is an expression over the other two rather than a third
    ...    subquery, so it cannot report a total that disagrees with them.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=fields=id,on_hand_quantity,reserved_quantity,available_quantity
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}
    ${expected}=    Evaluate    float($item['on_hand_quantity']) - float($item['reserved_quantity'])
    Should Be Equal As Numbers    ${item}[available_quantity]    ${expected}

Asking Only For Available Still Answers Correctly
    [Documentation]    available is filled in Go from two operands the client never named. If the
    ...    engine did not pull them into the projection, it would evaluate against absent operands
    ...    and read as zero for a product that does hold stock.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=fields=id,available_quantity
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    available_quantity

The Columns Are Read-Only
    [Documentation]    AC-PROD-INT-034. Product must not be able to set a balance. A computed
    ...    field is rejected on write rather than silently dropped, so a client that tries is told.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    Response Status Should Be    ${resp}    200
    ${item}=    Set Variable    ${resp.json()}
    ${resp}=    PUT On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'on_hand_quantity': '999', 'etag': $item['etag']} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=Writing a stock figure through the product must be refused

    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    params=fields=id,on_hand_quantity
    Response Status Should Be    ${resp}    200
    Should Be Equal As Numbers    ${resp.json()}[on_hand_quantity]    0
    ...    msg=The refused write must not have reached anything

Stock Cannot Be Filtered Or Sorted On Yet
    [Documentation]    The figures have no column behind them, so SQL cannot ORDER BY or WHERE
    ...    them. This is the documented boundary of INV-PI-016, asserted rather than assumed: it
    ...    changes when stored computed fields ship, and this test is what will notice.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}
    ...    params=order_by=on_hand_quantity desc    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    ...    msg=Sorting by a field with no column must be refused, not silently ignored
