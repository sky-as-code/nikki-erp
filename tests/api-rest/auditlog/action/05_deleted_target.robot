*** Settings ***
Documentation     Deleting the audited category is recorded with the record as it was, and its
...               history outlives it: a deleted record cannot be read, so the history then asks
...               for read on the resource in the org instead. Last, doubling as cleanup.
Resource          resources/auditlog.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Audited Category
Test Tags         auditlog    action    delete


*** Test Cases ***
Delete Is Recorded With The Old Values
    ${resp}=    DELETE On Session    api    ${UOMCAT_API}/${AUDITED_CAT_ID}
    ...    params=${{ {'org_id': $UOM_ORG_ID} }}
    Response Should Be Delete Success    ${resp}
    ${entry}=    Latest Entry    ${AUDITED_CAT_ID}
    Should Be Equal    ${entry}[action]    delete
    ${name}=    Change Of    ${entry}    name
    Should Be Equal    ${name}[old][en-US]    ${AUDITED_CAT_NAME}
    Should Be Equal    ${name}[new]    ${None}

History Of The Deleted Record Is Still Readable
    ${body}=    Get History    ${AUDITED_CAT_ID}
    ${actions}=    Evaluate    [item['action'] for item in $body['items']]
    Should Be Equal    ${actions}[0]    delete
    Should Be Equal    ${actions}[-1]    create
    Set Global Variable    ${AUDITED_CAT_ID}    ${EMPTY}

History Of A Deleted Record Still Requires The Org
    [Tags]    negative
    ${resp}=    Get History Response
    ...    ${{ {'resource_code': $AUDIT_RESOURCE_CODE, 'resource_id': $NOT_FOUND_ID} }}
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    common:err_org_id_required
