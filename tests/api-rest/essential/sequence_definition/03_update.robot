*** Settings ***
Documentation     Updating the definition under test while it has issued nothing: its whole
...               configuration is still editable, under optimistic concurrency (AC27).
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Sequence Definition Under Test
Test Tags         essential    sequence_definition    update


*** Test Cases ***
Update Name Succeeds
    ${name}=    Unique Display Name    Robot Sequence Renamed
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}
    ...    json=${{ {'name': $name, 'etag': $SEQ_DEFINITION_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${SEQ_DEFINITION_ETAG}
    Set Global Variable    ${SEQ_DEFINITION_ETAG}    ${etag}

Update Configuration Before The First Issuance Succeeds
    [Documentation]    SEQ-CFG-01 locks the configuration only once a number was issued.
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}
    ...    json=${{ {'suffix_template': '-X', 'etag': $SEQ_DEFINITION_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${SEQ_DEFINITION_ETAG}
    Set Global Variable    ${SEQ_DEFINITION_ETAG}    ${etag}
    ${item}=    Read Sequence Definition    ${SEQ_DEFINITION_ID}
    Should Be Equal    ${item}[suffix_template]    -X

Update With An Invalid Template Fails
    [Documentation]    An update is validated like a create, on the record it would produce.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}
    ...    json=${{ {'prefix_template': 'RB9', 'etag': $SEQ_DEFINITION_ETAG} }}    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_INVALID_TEMPLATE}    prefix_template

Update With A Stale Etag Fails
    [Documentation]    AC27: an update carrying an old version must not overwrite the newer one.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}
    ...    json=${{ {'description': 'stale', 'etag': '1000000000000000000'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}
    ${item}=    Read Sequence Definition    ${SEQ_DEFINITION_ID}
    Should Not Be Equal    ${item.get('description')}    stale

Update With Algorithm Params Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}
    ...    json=${{ {'algorithm_params': {'key': '00'}, 'etag': $SEQ_DEFINITION_ETAG} }}    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_READ_ONLY_FIELD}    algorithm_params

Update Cannot Change The Algorithm
    [Documentation]    algorithm is no_update: switching an issued sequence to FF1 would re-key
    ...    every future code. The engine drops or refuses it; either way it must not change.
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}
    ...    json=${{ {'algorithm': 'pseudorandom', 'etag': $SEQ_DEFINITION_ETAG} }}    expected_status=any
    ${item}=    Read Sequence Definition    ${SEQ_DEFINITION_ID}
    Should Be Equal    ${item}[algorithm]    increment
    Set Global Variable    ${SEQ_DEFINITION_ETAG}    ${item}[etag]

Update Of An Unknown Definition Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${NOT_FOUND_ID}
    ...    json=${{ {'description': 'ghost', 'etag': '1000000000000000000'} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
