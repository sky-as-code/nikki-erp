*** Settings ***
Documentation     Existence checks over Stock Locations.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Location Under Test
Test Tags         inventory    stock_location    exists


*** Test Cases ***
Exists With One Id Succeeds
    ${resp}=    POST On Session    api    ${STOCK_LOCATION_API}/exists
    ...    json=${{ {'ids': [$STOCK_LOCATION_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=0

Exists With Not Found Id Succeeds
    [Documentation]    A missing id is an answer, not an error: exists reports it as
    ...    not-existing rather than failing the request.
    ${resp}=    POST On Session    api    ${STOCK_LOCATION_API}/exists
    ...    json=${{ {'ids': [$NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=0    not_existing=1

Exists With Mixed Ids Succeeds
    ${resp}=    POST On Session    api    ${STOCK_LOCATION_API}/exists
    ...    json=${{ {'ids': [$STOCK_LOCATION_ID, $NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=1

Exists With Empty Ids Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_LOCATION_API}/exists
    ...    json=${{ {'ids': []} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=An empty id list is a malformed request, not a query with no results
