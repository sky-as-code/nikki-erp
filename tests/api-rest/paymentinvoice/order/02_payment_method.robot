*** Settings ***
Documentation     The payment method an order names. It is the one resource in this suite with
...               an ordinary CRUD lifecycle, and it is covered here because create_payment
...               cannot be reached without one.
Resource          resources/paymentinvoice.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Payment Currency
Test Tags         paymentinvoice    order    payment_method


*** Test Cases ***
Create Payment Method Succeeds
    [Documentation]    Saves the method under test, consumed by 03_create_payment.robot and
    ...    removed by the suite teardown.
    ${code}=    Unique Payment Code    method
    ${resp}=    POST On Session    api    ${PAYMENT_METHOD_API}
    ...    json=${{ {'code': $code, 'adapter_code': 'momo', 'name': {'en-US': 'Robot Momo'}, 'currency_id': $PAYINV_CURRENCY_ID, 'is_active': True, 'min_amount': '1000', 'max_amount': '50000000'} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PAYMENT_METHOD_ID}    ${id}
    Set Global Variable    ${PAYMENT_METHOD_ETAG}    ${etag}
    Set Global Variable    ${PAYMENT_METHOD_CODE}    ${code}

Get Payment Method Succeeds
    ${resp}=    GET On Session    api    ${PAYMENT_METHOD_API}/${PAYMENT_METHOD_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/payment_method.json    200
    Should Be Equal    ${item}[adapter_code]    momo
    Should Be Equal    ${item}[is_active]    ${True}

Update Payment Method Succeeds
    ${resp}=    PATCH On Session    api    ${PAYMENT_METHOD_API}/${PAYMENT_METHOD_ID}
    ...    json=${{ {'name': {'en-US': 'Robot Momo Updated'}, 'etag': $PAYMENT_METHOD_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PAYMENT_METHOD_ETAG}
    IF    $etag is not None    Set Global Variable    ${PAYMENT_METHOD_ETAG}    ${etag}

Withdrawing A Method Is A Flag Not A Deletion
    [Documentation]    is_active false is how a method stops taking payments. Deleting it
    ...    would orphan every order that named it, and those are the financial record.
    ${resp}=    PATCH On Session    api    ${PAYMENT_METHOD_API}/${PAYMENT_METHOD_ID}
    ...    json=${{ {'is_active': False, 'etag': $PAYMENT_METHOD_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PAYMENT_METHOD_ETAG}
    IF    $etag is not None    Set Global Variable    ${PAYMENT_METHOD_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${PAYMENT_METHOD_API}/${PAYMENT_METHOD_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${PAYINV_SCHEMA_DIR}/payment_method.json    200
    Should Be Equal    ${item}[is_active]    ${False}
    #    Restored, because 03_create_payment.robot needs an active method to reach the
    #    gateway-disabled refusal rather than being turned away earlier.
    ${resp}=    PATCH On Session    api    ${PAYMENT_METHOD_API}/${PAYMENT_METHOD_ID}
    ...    json=${{ {'is_active': True, 'etag': $item['etag']} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${item}[etag]
    IF    $etag is not None    Set Global Variable    ${PAYMENT_METHOD_ETAG}    ${etag}

Create With Missing Required Fields Fails
    [Documentation]    is_active is absent from this list although it is required-for-create:
    ...    it carries a default_value, so the engine fills it rather than reporting it missing.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PAYMENT_METHOD_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    adapter_code    name

Create With Duplicate Code Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PAYMENT_METHOD_API}
    ...    json=${{ {'code': $PAYMENT_METHOD_CODE, 'adapter_code': 'momo', 'name': {'en-US': 'Robot Duplicate'}, 'is_active': True} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=Two payment methods must not share a code

Search Payment Methods Succeeds
    ${resp}=    GET On Session    api    ${PAYMENT_METHOD_API}
    Response Should Be Search Success    ${resp}    ${PAYINV_SCHEMA_DIR}/payment_method.json
    ...    size=50    page=0
    Search Results Should Contain Id    ${resp}    ${PAYMENT_METHOD_ID}
