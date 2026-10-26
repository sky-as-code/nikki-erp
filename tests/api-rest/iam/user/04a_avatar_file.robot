*** Settings ***
Documentation     The avatar of the user under test is an IT Storage file id. The file is uploaded to
...               /v1/itstorage/files/upload first (it starts pending), then avatar_file_id is
...               written through the plain PATCH, which marks the file in use. Replacing or
...               clearing the avatar releases the old file back to pending, so its object expires.
...               No presigned URL is returned and no per-field files route exists.
...               Runs after GET and before EXISTS; every mutation rotates ${USER_ETAG} for the
...               suites after it.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure User Under Test
Test Tags         iam    user    file


*** Variables ***
${STORAGE_UPLOAD_API}    /v1/itstorage/files/upload
${STORAGE_FILE_API}      /v1/itstorage/files


*** Test Cases ***
Set Avatar From An Uploaded File Succeeds
    ${file_id}=    Upload Avatar To Storage    png
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_id': $file_id, 'etag': $USER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}
    Set Suite Variable    ${FIRST_AVATAR_ID}    ${file_id}
    Stored File Status Should Be    ${file_id}    in_use

Get Returns The File Id And No Url
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user.json    200
    Should Be Equal    ${item}[avatar_file_id]    ${FIRST_AVATAR_ID}
    Dictionary Should Not Contain Key    ${item}    avatar_file_url
    Set Global Variable    ${USER_ETAG}    ${item}[etag]

Replace Avatar Succeeds And Releases The Old File
    ${file_id}=    Upload Avatar To Storage    gif
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_id': $file_id, 'etag': $USER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}
    Set Suite Variable    ${SECOND_AVATAR_ID}    ${file_id}
    Stored File Status Should Be    ${file_id}    in_use
    Stored File Status Should Be    ${FIRST_AVATAR_ID}    pending

Update With An Unknown File Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_id': '01NOSUCHFILE00000000000000', 'etag': $USER_ETAG} }}
    ...    expected_status=any
    Response Should Be File Field Error    ${resp}    avatar_file_id    ${EMPTY}

The Old Files Route Is Gone
    [Tags]    negative
    ${part}=    Avatar File Part    png
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${part}    expected_status=any
    Should Be True    ${resp.status_code} in (404, 405)

Clear Avatar Succeeds And Releases The File
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_id': None, 'etag': $USER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}
    Stored File Status Should Be    ${SECOND_AVATAR_ID}    pending

Get After Clear Has No Avatar
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user.json    200
    ${file_id}=    Evaluate    $item.get('avatar_file_id')
    Should Be Equal    ${file_id}    ${None}
    Set Global Variable    ${USER_ETAG}    ${item}[etag]


*** Keywords ***
Upload Avatar To Storage
    [Documentation]    Uploads a real image to the tenant space (a user has no org) and returns
    ...    its file id. The text parts precede the file part, as the upload API requires.
    [Arguments]    ${kind}
    ${part}=    Avatar File Part    ${kind}
    ${size}=    Evaluate    len($part['file'][1])
    ${resp}=    POST On Session    api    ${STORAGE_UPLOAD_API}
    ...    data=${{ {'name': 'avatar.' + $kind, 'mime_type': 'image/' + $kind, 'size_bytes': str($size)} }}
    ...    files=${part}
    Should Be True    ${resp.status_code} in (200, 201)    msg=upload failed: ${resp.text}
    Should Be Equal    ${resp.json()}[status]    pending
    ${file_id}=    Set Variable    ${resp.json()}[id]
    Should Not Be Empty    ${file_id}
    RETURN    ${file_id}

Stored File Status Should Be
    [Documentation]    The file's lifecycle state, read by id. An avatar is tenant-wide (no org), so
    ...    the org-scoped standard itstorage_file GET cannot read it.
    [Arguments]    ${file_id}    ${expected}
    ${resp}=    GET On Session    api    ${STORAGE_FILE_API}/${file_id}/meta    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200    msg=file ${file_id}: ${resp.text}
    Should Be Equal    ${resp.json()}[status]    ${expected}
