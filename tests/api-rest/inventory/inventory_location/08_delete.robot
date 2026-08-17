*** Settings ***
Documentation     Deleting the Inventory Location under test — always the LAST suite, doubling
...               as cleanup. The child created by 02_create.robot points at it, so it goes
...               first: every FK here is ON DELETE NO ACTION, so a location with a
...               dependant cannot be removed.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Location Under Test
Test Tags         inventory    inventory_location    delete


*** Test Cases ***
Delete With Referencing Child Fails
    [Documentation]    The parent_location_id FK is ON DELETE NO ACTION, so removing a
    ...    location that still has children must be refused rather than orphaning them.
    [Tags]    negative
    ${child}=    Get Variable Value    ${CHILD_INVENTORY_LOCATION_ID}    ${EMPTY}
    IF    not $child    Skip    No child location was created by 02_create.robot
    ${resp}=    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=A location with a child must not be deletable

Delete Child Succeeds
    ${child}=    Get Variable Value    ${CHILD_INVENTORY_LOCATION_ID}    ${EMPTY}
    IF    not $child    Skip    No child location was created by 02_create.robot
    ${resp}=    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${child}
    Response Should Be Delete Success    ${resp}    count=1
    Set Global Variable    ${CHILD_INVENTORY_LOCATION_ID}    ${EMPTY}

Delete Succeeds
    ${resp}=    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    location under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${INVENTORY_LOCATION_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${INVENTORY_LOCATION_ID}    ${EMPTY}
    Set Global Variable    ${INVENTORY_LOCATION_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${INVENTORY_LOCATION_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
