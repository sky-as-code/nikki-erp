*** Settings ***
Documentation     The avatar of the user under test is an IT Storage path. The file is uploaded to
...               /v1/itstorage/files/upload first, then avatar_file_path is written through the
...               plain PATCH. No presigned URL is returned and no per-field files route exists.
...               Runs after GET and before EXISTS; every mutation rotates ${USER_ETAG} for the
...               suites after it.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure User Under Test
Test Tags         iam    user    file


*** Variables ***
${STORAGE_UPLOAD_API}    /v1/itstorage/files/upload


*** Test Cases ***
Set Avatar From An Uploaded File Succeeds
    ${path}=    Upload Avatar To Storage    png
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_path': $path, 'etag': $USER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}
    Set Suite Variable    ${FIRST_AVATAR_PATH}    ${path}

Get Returns The Path And No Url
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user.json    200
    Should Be Equal    ${item}[avatar_file_path]    ${FIRST_AVATAR_PATH}
    Dictionary Should Not Contain Key    ${item}    avatar_file_url
    Set Global Variable    ${USER_ETAG}    ${item}[etag]

Replace Avatar Succeeds And Removes The Old File
    ${path}=    Upload Avatar To Storage    gif
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_path': $path, 'etag': $USER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}
    ${meta}=    GET On Session    api    /v1/itstorage/files/meta
    ...    params=${{ {'path': $FIRST_AVATAR_PATH} }}    expected_status=any
    # IT Storage answers a missing file with err_not_found, a business violation (400), as its own suite asserts.
    Should Be Equal As Integers    ${meta.status_code}    400

Update With An Unknown Path Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_path': '/no/such/01NOSUCHFILE00000000000000.png', 'etag': $USER_ETAG} }}
    ...    expected_status=any
    Response Should Be File Field Error    ${resp}    avatar_file_path    ${EMPTY}

The Old Files Route Is Gone
    [Tags]    negative
    ${part}=    Avatar File Part    png
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${part}    expected_status=any
    Should Be True    ${resp.status_code} in (404, 405)

Clear Avatar Succeeds
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_path': None, 'etag': $USER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}

Get After Clear Has No Avatar
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user.json    200
    ${path}=    Evaluate    $item.get('avatar_file_path')
    Should Be Equal    ${path}    ${None}
    Set Global Variable    ${USER_ETAG}    ${item}[etag]


*** Keywords ***
Upload Avatar To Storage
    [Documentation]    Uploads a real image to the tenant space (a user has no org) and returns
    ...    its file_path. The text parts precede the file part, as the upload API requires.
    [Arguments]    ${kind}
    ${part}=    Avatar File Part    ${kind}
    ${size}=    Evaluate    len($part['file'][1])
    ${resp}=    POST On Session    api    ${STORAGE_UPLOAD_API}
    ...    data=${{ {'name': 'avatar.' + $kind, 'mime_type': 'image/' + $kind, 'size_bytes': str($size)} }}
    ...    files=${part}
    Should Be True    ${resp.status_code} in (200, 201)    msg=upload failed: ${resp.text}
    ${path}=    Set Variable    ${resp.json()}[file_path]
    Should Not Be Empty    ${path}
    RETURN    ${path}
