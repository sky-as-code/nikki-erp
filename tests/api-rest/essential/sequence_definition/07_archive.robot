*** Settings ***
Documentation     Archive and unarchive (BR §4.4, AC28-AC30). Archiving asks every dependant
...               module whether the definition is still configured for use; a definition a
...               module setting points at cannot be archived.
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Sequence Definition Under Test
...               AND    Ensure Issued Sequence Definition
Test Tags         essential    sequence_definition    archive


*** Test Cases ***
Archiving The Org Transfer Definition Is Refused
    [Documentation]    AC28: Inventory keeps its id in the org settings (transfer_sequence_definition_id);
    ...    every operation type of the org would be left numbering from an archived definition.
    [Tags]    negative
    ${item}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/${SEQ_ISSUED_DEFINITION_ID}/archived
    ...    json=${{ {'etag': $item['etag'], 'is_archived': True} }}    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_USED_BY_MODULES}    id
    ${after}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    Should Not Be True    ${after}[is_archived]

Archiving A Definition A Module Setting Points At Is Refused
    [Documentation]    AC28: Sales keeps sales.order's id in its org settings; archiving it would
    ...    leave every new order unnumbered.
    [Tags]    negative
    Ensure Sequence Org
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}
    ...    params=${{ {'graph': __import__('json').dumps({'and': [{'if': ['code', '=', 'sales.order']}, {'if': ['org_id', '=', $SEQ_ORG_ID]}]})} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    IF    not $items    Skip    Org ${SEQ_ORG_ID} has no sales.order definition
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/${items}[0][id]/archived
    ...    json=${{ {'etag': $items[0]['etag'], 'is_archived': True} }}    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_USED_BY_MODULES}    id

Archive An Unused Definition Succeeds
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}/archived
    ...    json=${{ {'etag': $SEQ_DEFINITION_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${SEQ_DEFINITION_ETAG}
    Set Global Variable    ${SEQ_DEFINITION_ETAG}    ${etag}
    ${item}=    Read Sequence Definition    ${SEQ_DEFINITION_ID}
    Should Be True    ${item}[is_archived]

Unarchive Succeeds
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}/archived
    ...    json=${{ {'etag': $SEQ_DEFINITION_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${SEQ_DEFINITION_ETAG}
    Set Global Variable    ${SEQ_DEFINITION_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': '1000000000000000000', 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}/archived
    ...    json=${{ {'etag': '1000000000000000000', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}
