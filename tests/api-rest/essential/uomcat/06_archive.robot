*** Settings ***
Documentation     Archiving the UoM Category under test, rotating the saved etag. The
...               category is unarchived again so the later suites see it live.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Category Under Test
Test Tags         essential    uomcat    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${UOMCAT_API}/${UOMCAT_ID}/archived
    ...    json=${{ {'etag': $UOMCAT_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOMCAT_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOMCAT_ETAG}    ${etag}

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${UOMCAT_API}/${UOMCAT_ID}/archived
    ...    json=${{ {'etag': $UOMCAT_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOMCAT_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOMCAT_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOMCAT_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $UOMCAT_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOMCAT_API}/${UOMCAT_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOMCAT_API}/${UOMCAT_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
