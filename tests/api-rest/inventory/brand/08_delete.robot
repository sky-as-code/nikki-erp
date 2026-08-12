*** Settings ***
Documentation     Deleting the Brand under test — always the LAST suite, doubling as
...               cleanup. Unlike a product type, brand_id is NULLABLE on a template, so a
...               brand has no mandatory dependant and there is no "delete with referencing
...               records fails" case here: the brand deletes cleanly precisely because
...               nothing is required to reference it.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Brand Under Test
Test Tags         inventory    brand    delete


*** Test Cases ***
Delete Succeeds
    ${resp}=    DELETE On Session    api    ${BRAND_API}/${BRAND_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    brand under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${BRAND_API}/${BRAND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${BRAND_ID}    ${EMPTY}
    Set Global Variable    ${BRAND_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${BRAND_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${BRAND_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
