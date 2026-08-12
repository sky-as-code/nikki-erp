*** Settings ***
Documentation     Updating UoM Categories, including the Reference UoM rules of
...               BR-UOM-ESS-004. The success case runs first (it consumes and rotates
...               the saved etag); negatives follow.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Category Under Test
Test Tags         essential    uomcat    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Category
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $UOMCAT_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOMCAT_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOMCAT_ETAG}    ${etag}
    Set Global Variable    ${UOMCAT_NAME}    ${name}

Set Reference Uom Succeeds
    [Documentation]    BR-UOM-ESS-003: pointing the category at a UoM of its own is the
    ...    supported way to give it a Reference UoM.
    Ensure Reference Uom
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    json=${{ {'reference_uom_id': $REFERENCE_UOM_ID, 'etag': $UOMCAT_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOMCAT_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOMCAT_ETAG}    ${etag}

Update With Foreign Reference Uom Fails
    [Documentation]    BR-UOM-ESS-004 / UOM-ESS-INV-03.
    [Tags]    negative
    Ensure Foreign Uom Category
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    json=${{ {'reference_uom_id': $FOREIGN_UOM_ID, 'etag': $UOMCAT_ETAG} }}
    ...    expected_status=any
    Response Should Be Uomcat Foreign Reference Error    ${resp}

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Category
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/${UOMCAT_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $UOMCAT_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOMCAT_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $UOMCAT_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
