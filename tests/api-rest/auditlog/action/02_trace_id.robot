*** Settings ***
Documentation     One request writes all its entries under one trace_id: the OpenTelemetry trace
...               id the ingress middleware echoes in the X-Trace-ID response header. The bulk
...               create writes two records in one request, so their entries must share it; two
...               separate requests (no traceparent sent) start separate traces and must not.
Resource          resources/auditlog.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Audited Category
Test Tags         auditlog    action    trace_id


*** Test Cases ***
Entry Carries The Trace Id Of Its Request
    ${body}=    Get History    ${AUDITED_CAT_ID}
    ${create}=    Set Variable    ${body}[items][-1]
    Should Be Equal    ${create}[action]    create
    Should Be Equal    ${create}[trace_id]    ${AUDITED_CAT_TRACE_ID}
    ...    msg=The create entry must carry the X-Trace-ID of the create request

Records Written By One Request Share Its Trace Id
    ${name_a}=    Unique Display Name    Robot Bulk Category A
    ${name_b}=    Unique Display Name    Robot Bulk Category B
    ${items}=    Evaluate    [{'name': {'en-US': $name_a}, 'org_id': $UOM_ORG_ID}, {'name': {'en-US': $name_b}, 'org_id': $UOM_ORG_ID}]
    ${resp}=    POST On Session    api    ${UOMCAT_API}/bulk
    ...    json=${{ {'org_id': $UOM_ORG_ID, 'items': $items} }}
    Should Be True    ${resp.status_code} in (200, 201)    msg=Bulk create failed: ${resp.text}
    ${trace_id}=    Set Variable    ${resp.headers}[X-Trace-ID]
    ${id_a}=    Find Category Id    ${name_a}
    Set Global Variable    ${BULK_CAT_A_ID}    ${id_a}
    ${id_b}=    Find Category Id    ${name_b}
    Set Global Variable    ${BULK_CAT_B_ID}    ${id_b}
    ${entry_a}=    Latest Entry    ${id_a}
    ${entry_b}=    Latest Entry    ${id_b}
    Should Be Equal    ${entry_a}[trace_id]    ${trace_id}
    Should Be Equal    ${entry_b}[trace_id]    ${trace_id}

Separate Requests Get Separate Trace Ids
    ${body}=    Get History    ${AUDITED_CAT_ID}
    ${ids}=    Evaluate    {item['trace_id'] for item in $body['items']}
    ${count}=    Get Length    ${body}[items]
    Length Should Be    ${ids}    ${count}    msg=Each mutation of the category was its own request


*** Keywords ***
Find Category Id
    [Arguments]    ${name}
    ${graph}=    Evaluate    json.dumps({'if': ['name', '=', $name]})    modules=json
    ${resp}=    GET On Session    api    ${UOMCAT_API}
    ...    params=${{ {'org_id': $UOM_ORG_ID, 'graph': $graph, 'language': 'en-US'} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    1
    RETURN    ${resp.json()}[items][0][id]
