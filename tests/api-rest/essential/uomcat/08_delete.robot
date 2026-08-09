*** Settings ***
Documentation     Deleting the UoM Category under test — always the LAST suite, doubling
...               as cleanup. Its member units must go first: the category FK is
...               ON DELETE NO ACTION so that BR-UOM-ESS-002 (a UoM belongs to exactly
...               one category) cannot be violated by orphaning them.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Category Under Test
Test Tags         essential    uomcat    delete


*** Test Cases ***
Delete With Member Uoms Fails
    [Documentation]    BR-UOM-ESS-002: removing a category out from under its units would
    ...    leave them belonging to nothing, so the database refuses.
    [Tags]    negative
    Ensure Reference Uom
    ${resp}=    DELETE On Session    api    ${UOMCAT_API}/${UOMCAT_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A category with member UoMs must not be deletable

Delete Succeeds
    [Documentation]    With the members gone the category deletes normally. Delete Uom
    ...    Seed Data (suite teardown) removes the units; doing it here keeps this test
    ...    independent of teardown ordering.
    Delete Uom Seed Data
    ${resp}=    DELETE On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so
    ...    the category under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOMCAT_API}/${UOMCAT_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${UOMCAT_ID}    ${EMPTY}
    Set Global Variable    ${UOMCAT_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOMCAT_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOMCAT_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
