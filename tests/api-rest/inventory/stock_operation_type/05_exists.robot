*** Settings ***
Documentation     Existence checks over Stock Operation Types.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Operation Type Under Test
Test Tags         inventory    stock_operation_type    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/exists
    ...    json=${{ {'ids': [$STOCK_OPERATION_TYPE_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Not Found Id Succeeds
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/exists
    ...    json=${{ {'ids': [$NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=0    not_existing=1

Exists With Mixed Ids Succeeds
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/exists
    ...    json=${{ {'ids': [$STOCK_OPERATION_TYPE_ID, $NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=1
