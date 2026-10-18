*** Settings ***
Documentation     The avatar file field of the user under test: files/avatar is the only writer of
...               avatar_file_key, and every read presigns it into avatar_file_url. Runs after GET
...               and before EXISTS; every mutation rotates ${USER_ETAG} for the suites after it.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure User Under Test
Test Tags         iam    user    file


*** Test Cases ***
Upload Avatar Succeeds
    ${part}=    Avatar File Part    png
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${part}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}

Get Presigns The Avatar
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user.json    200
    Should Not Be Empty    ${item}[avatar_file_key]
    Should Start With    ${item}[avatar_file_url]    http
    Set Global Variable    ${USER_ETAG}    ${item}[etag]

Upload Again Without Replace Fails
    [Tags]    negative
    ${part}=    Avatar File Part    png
    ${resp}=    POST On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${part}    expected_status=any
    Response Should Be File Field Error    ${resp}    avatar    common:file.err_file_exists

Replace Avatar Succeeds
    ${part}=    Avatar File Part    gif
    ${resp}=    PUT On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${part}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}

Upload To Unknown File Field Fails
    [Tags]    negative
    ${part}=    Avatar File Part    png
    ${resp}=    PUT On Session    api    ${USER_API}/${USER_ID}/files/no_such_field
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${part}    expected_status=any
    Response Should Be File Field Error    ${resp}    field    common:file.err_field_unknown

Upload Of A Non Image Fails
    [Tags]    negative
    ${part}=    Evaluate    {'file': ('avatar.png', b'plain text, not an image', 'image/png')}
    ${resp}=    PUT On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${part}    expected_status=any
    Response Should Be File Field Error    ${resp}    avatar    common:err_file_type_not_allowed

Upload Of An Oversize Image Fails
    [Tags]    negative
    ${png}=    Avatar File Part    png
    ${padded}=    Evaluate    {'file': ('avatar.png', $png['file'][1] + b'\\0' * (10 * 1024 * 1024 + 1), 'image/png')}
    ${resp}=    PUT On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    data=${{ {'etag': $USER_ETAG} }}    files=${padded}    expected_status=any
    Response Should Be File Field Error    ${resp}    avatar    common:err_file_too_large

Update Setting The File Key Directly Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${USER_API}/${USER_ID}
    ...    json=${{ {'avatar_file_key': 'iam/users/forged.png', 'etag': $USER_ETAG} }}    expected_status=any
    Response Should Be File Field Error    ${resp}    avatar_file_key    common:file.err_field_managed

Delete Avatar Succeeds
    ${resp}=    DELETE On Session    api    ${USER_API}/${USER_ID}/files/avatar
    ...    params=${{ {'etag': $USER_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    previous_etag=${USER_ETAG}
    Set Global Variable    ${USER_ETAG}    ${etag}

Get After Delete Has No Avatar
    ${resp}=    GET On Session    api    ${USER_API}/${USER_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${IAM_SCHEMA_DIR}/user.json    200
    Dictionary Should Not Contain Key    ${item}    avatar_file_url
    Set Global Variable    ${USER_ETAG}    ${item}[etag]
