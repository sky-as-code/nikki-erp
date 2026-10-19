*** Settings ***
Documentation     An organization is set up and removed by every installed module, after the
...               request that asked for it has answered. A create answers with the org
...               provisioning and it turns ready once each module has written its defaults;
...               a delete archives it as deleted and removes it once each module has cleaned up.
...               The suite polls for both outcomes, then checks the defaults the org got.
Resource          ../../resources/iam.resource
Suite Setup       Create Authorized API Session


*** Variables ***
${ORG_API}              /v1/iam/iam_org
# Long enough for every module to answer, short enough that a hung module fails the run.
${LIFECYCLE_TIMEOUT}    60s
${POLL_INTERVAL}        2s


*** Test Cases ***
A New Org Starts Provisioning Whatever The Client Sends
    [Documentation]    status is the org lifecycle's to write: a client asking for ready is ignored.
    [Tags]    org    lifecycle
    ${org_id}=    Create Lifecycle Org    status=ready
    ${org}=    Get Org    ${org_id}
    Should Be True    $org['status'] in ('provisioning', 'ready')
    Set Suite Variable    ${LIFECYCLE_ORG_ID}    ${org_id}

The Org Becomes Ready Once Every Module Has Set It Up
    [Tags]    org    lifecycle
    Wait Until Org Is Ready    ${LIFECYCLE_ORG_ID}    ${LIFECYCLE_TIMEOUT}
    ${org}=    Get Org    ${LIFECYCLE_ORG_ID}
    Should Be Equal    ${org.get('status_message')}    ${None}

A Ready Org Is In Its Creator's Context
    [Tags]    org    lifecycle
    ${me}=    GET On Session    api    /v1/iam/me/context
    Response Status Should Be    ${me}    200
    ${org_ids}=    Evaluate    [org['id'] for org in body['orgs']]    namespace=${{ {'body': $me.json()} }}
    Should Contain    ${org_ids}    ${LIFECYCLE_ORG_ID}

A Ready Org Has Its Units Of Measure
    [Documentation]    Essential's defaults: the unit catalogue, written for this org alone.
    [Tags]    org    lifecycle    defaults
    ${resp}=    GET On Session    api    /v1/essential/essential_uom
    ...    params=${{ {'org_id': $LIFECYCLE_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${symbols}=    Evaluate    [item['symbol'] for item in body['items']]    namespace=${{ {'body': $resp.json()} }}
    Should Contain    ${symbols}    pc
    Should Contain    ${symbols}    kg

A Ready Org Has A Main Warehouse
    [Documentation]    Inventory's defaults: a warehouse created through its own provisioning.
    [Tags]    org    lifecycle    defaults
    ${resp}=    GET On Session    api    /v1/inventory/inventory_warehouse
    ...    params=${{ {'org_id': $LIFECYCLE_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${codes}=    Evaluate    [item['code'] for item in body['items']]    namespace=${{ {'body': $resp.json()} }}
    Should Contain    ${codes}    MAIN

Deleting An Org Archives It Then Removes It
    [Tags]    org    lifecycle
    ${resp}=    DELETE On Session    api    ${ORG_API}/${LIFECYCLE_ORG_ID}
    Response Should Be Delete Success    ${resp}
    Wait Until Keyword Succeeds    ${LIFECYCLE_TIMEOUT}    ${POLL_INTERVAL}    Org Should Be Gone    ${LIFECYCLE_ORG_ID}


*** Keywords ***
Create Lifecycle Org
    [Arguments]    ${status}=${None}
    ${name}=    Unique Display Name    Lifecycle Org
    ${slug}=    Evaluate    'lifecycle-org-' + str(__import__('uuid').uuid4())[:8]
    ${body}=    Create Dictionary    display_name=${{ {'en-US': $name} }}    slug=${slug}
    IF    $status is not None    Set To Dictionary    ${body}    status=${status}
    ${resp}=    POST On Session    api    ${ORG_API}    json=${body}
    ${org_id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${org_id}

Get Org
    [Arguments]    ${org_id}
    ${resp}=    GET On Session    api    ${ORG_API}/${org_id}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item]

Org Should Be Gone
    [Arguments]    ${org_id}
    ${resp}=    GET On Session    api    ${ORG_API}/${org_id}    expected_status=any
    Response Should Be Not Found Error    ${resp}
