*** Settings ***
Documentation     Each built-in mutation writes one entry naming its action, holding only the
...               fields that changed: never id, etag, updated_at or another system field.
Resource          resources/auditlog.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Audited Category
Test Tags         auditlog    action    mutation


*** Test Cases ***
Create Is Recorded With The New Values
    ${body}=    Get History    ${AUDITED_CAT_ID}
    Should Be Equal As Integers    ${body}[total]    1    msg=A fresh record has exactly its create entry
    ${entry}=    Set Variable    ${body}[items][0]
    Should Be Equal    ${entry}[action]    create
    ${name}=    Change Of    ${entry}    name
    Should Be Equal    ${name}[old]    ${None}
    Should Be Equal    ${name}[new][en-US]    ${AUDITED_CAT_NAME}
    ${fields}=    Evaluate    [change['field'] for change in $entry['changes']]
    FOR    ${system}    IN    id    etag    created_at    updated_at    created_by    updated_by    org_id    tenant_id
        Should Not Contain    ${fields}    ${system}    msg=${system} is a system field and never in a diff
    END
    Should Be Equal    ${entry}[actor][type]    user
    Should Not Be Equal    ${entry}[actor][id]    ${None}

Update Records Only The Changed Field
    ${old_name}=    Set Variable    ${AUDITED_CAT_NAME}
    ${new_name}=    Unique Display Name    Robot Audited Renamed
    Update Audited Category Name    ${new_name}
    Set Global Variable    ${AUDITED_CAT_NAME}    ${new_name}
    ${entry}=    Latest Entry    ${AUDITED_CAT_ID}
    Entry Should Be    ${entry}    update    name
    ${name}=    Change Of    ${entry}    name
    Should Be Equal    ${name}[old][en-US]    ${old_name}
    Should Be Equal    ${name}[new][en-US]    ${new_name}

Update Without A Change Writes No Entry
    ${before}=    History Count    ${AUDITED_CAT_ID}
    Update Audited Category Name    ${AUDITED_CAT_NAME}
    ${after}=    History Count    ${AUDITED_CAT_ID}
    Should Be Equal As Integers    ${after}    ${before}    msg=An update that changes nothing must not be recorded

Archive Is Recorded
    Set Audited Category Archived    ${True}
    ${entry}=    Latest Entry    ${AUDITED_CAT_ID}
    Entry Should Be    ${entry}    archive    is_archived
    ${flag}=    Change Of    ${entry}    is_archived
    Should Be Equal    ${flag}[old]    ${False}
    Should Be Equal    ${flag}[new]    ${True}

Unarchive Is Recorded
    Set Audited Category Archived    ${False}
    ${entry}=    Latest Entry    ${AUDITED_CAT_ID}
    Entry Should Be    ${entry}    unarchive    is_archived

History Is Newest First And Pages From Zero
    ${body}=    Get History    ${AUDITED_CAT_ID}    page=0    size=2
    Should Be Equal As Integers    ${body}[page]    0
    Should Be Equal As Integers    ${body}[size]    2
    Length Should Be    ${body}[items]    2
    Should Be Equal    ${body}[items][0][action]    unarchive
    Should Be Equal    ${body}[items][1][action]    archive
    ${last}=    Get History    ${AUDITED_CAT_ID}    page=1    size=2
    Should Be Equal    ${last}[items][-1][action]    create    msg=The oldest entry is the create
