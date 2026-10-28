*** Settings ***
Documentation     One request writes all its entries under one request_id, minted per request
...               by the REST middleware and echoed in the X-Request-Id response header. The
...               bulk create writes two records in one request, so their entries must share it;
...               two separate requests must not.
Resource          resources/auditlog.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Audited Category
Test Tags         auditlog    action    request_id


*** Test Cases ***
Entry Carries The Request Id Of Its Request
    ${body}=    Get History    ${AUDITED_CAT_ID}
    ${create}=    Set Variable    ${body}[items][-1]
    Should Be Equal    ${create}[action]    create
    Should Be Equal    ${create}[request_id]    ${AUDITED_CAT_REQUEST_ID}
    ...    msg=The create entry must carry the X-Request-Id of the create request

Records Written By One Request Share Its Request Id
    ${name_a}=    Unique Display Name    Robot Bulk Category A
    ${name_b}=    Unique Display Name    Robot Bulk Category B
    ${items}=    Evaluate    [{'name': {'en-US': $name_a}, 'org_id': $UOM_ORG_ID}, {'name': {'en-US': $name_b}, 'org_id': $UOM_ORG_ID}]
    ${resp}=    POST On Session    api    ${UOMCAT_API}/bulk
    ...    json=${{ {'org_id': $UOM_ORG_ID, 'items': $items} }}
    Should Be True    ${resp.status_code} in (200, 201)    msg=Bulk create failed: ${resp.text}
    ${request_id}=    Set Variable    ${resp.headers}[X-Request-Id]
    ${id_a}=    Find Category Id    ${name_a}
    Set Global Variable    ${BULK_CAT_A_ID}    ${id_a}
    ${id_b}=    Find Category Id    ${name_b}
    Set Global Variable    ${BULK_CAT_B_ID}    ${id_b}
    ${entry_a}=    Latest Entry    ${id_a}
    ${entry_b}=    Latest Entry    ${id_b}
    Should Be Equal    ${entry_a}[request_id]    ${request_id}
    Should Be Equal    ${entry_b}[request_id]    ${request_id}

Separate Requests Get Separate Request Ids
    ${body}=    Get History    ${AUDITED_CAT_ID}
    ${ids}=    Evaluate    {item['request_id'] for item in $body['items']}
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
