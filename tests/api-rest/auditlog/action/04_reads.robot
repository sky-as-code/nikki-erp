*** Settings ***
Documentation     Read logging is opt-in per resource (AuditReads, nil means off). No resource
...               opts in yet, so a read of the UoM category, or of its history, must leave no
...               entry; the opted-in path (get and search entries) is covered by the composable
...               unit test TestGetByIdIsAuditedOnlyWhenOptedIn until a resource enables it.
Resource          resources/auditlog.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Audited Category
Test Tags         auditlog    action    reads


*** Test Cases ***
Reads Are Not Recorded By Default
    ${before}=    History Count    ${AUDITED_CAT_ID}
    ${resp}=    GET On Session    api    ${UOMCAT_API}/${AUDITED_CAT_ID}
    ...    params=${{ {'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${resp}=    GET On Session    api    ${UOMCAT_API}    params=${{ {'org_id': $UOM_ORG_ID, 'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${after}=    History Count    ${AUDITED_CAT_ID}
    Should Be Equal As Integers    ${after}    ${before}    msg=A read of a resource that did not opt in is not audited

Reading The History Is Not Recorded
    [Documentation]    The audit log is built with IsAudited false, so it never audits itself.
    ${first}=    History Count    ${AUDITED_CAT_ID}
    ${second}=    History Count    ${AUDITED_CAT_ID}
    Should Be Equal As Integers    ${second}    ${first}
