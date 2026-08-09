*** Settings ***
Documentation     Archiving a Unit of Measure (BR-UOM-ESS-019). Archive is the supported
...               way to retire a UoM: deleting one that historical documents reference
...               would change what those documents mean.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Under Test
Test Tags         essential    uom    archive


*** Test Cases ***
Archive Succeeds
    ${resp}=    POST On Session    api    ${UOM_API}/${UOM_ID}/archived
    ...    json=${{ {'etag': $UOM_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOM_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOM_ETAG}    ${etag}

Archived Uom Is Still Readable
    [Documentation]    BR-UOM-ESS-019: archived records stay in the data so historical
    ...    quantities remain interpretable; only new use is barred.
    ${resp}=    GET On Session    api    ${UOM_API}/${UOM_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uom.json    200
    Should Be True    ${item}[is_archived]    msg=The UoM should read back as archived

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${UOM_API}/${UOM_ID}/archived
    ...    json=${{ {'etag': $UOM_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOM_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOM_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $UOM_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_API}/${UOM_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_API}/${UOM_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
