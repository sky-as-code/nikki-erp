*** Settings ***
Documentation     The model schema of the value rows.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    settings_record    meta


*** Test Cases ***
Settings Record Model Schema Is Served
    ${resp}=    GET On Session    api    ${SETTINGS_RECORD_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Settings Records Carry No Version Column
    [Documentation]    D17: writes are last-write-wins per row. An etag reintroduced here would
    ...    make every partial save fail until the frontend started sending a version it has no
    ...    reason to hold, so the absence is pinned rather than assumed.
    ${resp}=    GET On Session    api    ${SETTINGS_RECORD_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${names}=    Evaluate    list($resp.json().get('fields', {}).keys())
    Should Not Contain    ${names}    etag
    ...    msg=settings_records must stay unversioned; see D17.
