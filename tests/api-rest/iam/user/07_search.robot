*** Settings ***
Documentation     Bruno: IAM/User/User - Search (+ Create test data + CORS). Graph
...               filters rely on the seeded users (Lead/Admin/.com variants), so they
...               pass on any database. The "graph" query values are JSON strings that
...               requests URL-encodes via the params dict.
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Seeded Users    50
...               AND    Ensure Archived User
Test Tags         iam    user    search


*** Variables ***
${USER_SCHEMA}    ${IAM_SCHEMA_DIR}/user.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${USER_API}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${USER_API}    params=${{ {'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${USER_API}    params=${{ {'page': 99} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=99    item_count=0

Search By Root Field Succeeds
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"if":["display_name", "*", "lead"]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search By Edge Field Succeeds
    [Tags]    seed
    Ensure Seed Group
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"if": ["groups.name", "!^", "brand"]}', 'fields': ['display_name', 'groups.name']} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search With And Condition Succeeds
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"and":[{"if":["display_name", "*", "admin"]},{"if": ["email", "$", ".com"]}]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search With Or Condition Succeeds
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"or":[{"if":["display_name", "*", "Lead"]},{"if": ["email", "^", "ed"]}]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search With Nested And Or Succeeds
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"or":[{"if":["display_name", "*", "Lead"]},{"and":[{"if": ["status", "*", "o"]},{"if": ["status", "*", "w"]}]}]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search Order By Root Field Succeeds
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"order": [["display_name", "desc"]]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search Order By Edge Root Field Succeeds
    [Tags]    seed
    Ensure Seed Group
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"order": [["groups.name", "desc"]]}', 'fields': ['display_name', 'groups.name']} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0

Search Order By Edge Json Subfield Succeeds
    [Documentation]    Orders by a JSON (langText) column reached through an edge. Bruno used
    ...    "user_status.label.vi_VN", which names no field on iam_user (the column is
    ...    "status") and spends 2 dots against the 1-dot limit of a graph field path, so it
    ...    could only ever 400. "groups.name" is the same shape the server does support.
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"order": [["groups.name", "desc"]]}'} }}    expected_status=any
    Response Status Should Be    ${resp}    200

Search Excludes Archived By Default
    [Documentation]    Omitting include_archived makes the repository prepend
    ...    "is_archived = false", so the archived fixture must not come back.
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"if":["display_name", "*", "Seed Archived User"]}'} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Not Contain Id    ${resp}    ${ARCHIVED_USER_ID}

Search With Include Archived False Excludes Archived
    [Documentation]    An explicit false must behave exactly like omitting the parameter.
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'include_archived': False, 'graph': '{"if":["display_name", "*", "Seed Archived User"]}'} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Not Contain Id    ${resp}    ${ARCHIVED_USER_ID}

Search With Include Archived True Includes Archived
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'include_archived': True, 'graph': '{"if":["display_name", "*", "Seed Archived User"]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0
    Search Results Should Contain Id    ${resp}    ${ARCHIVED_USER_ID}

Search With Include Archived And Graph Succeeds
    [Documentation]    Pins that the injected condition is ANDed with the caller's graph
    ...    rather than replacing it: the exact-name filter must still apply.
    ${graph}=    Evaluate    json.dumps({'if': ['display_name', '=', $ARCHIVED_USER_NAME]})    modules=json
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'include_archived': True, 'graph': $graph} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0    item_count=1
    Search Results Should Contain Id    ${resp}    ${ARCHIVED_USER_ID}

Search With Include Archived And Order Succeeds
    [Documentation]    Regression guard for the order-only graph: folding it into the
    ...    injected AND must not emit an empty predicate, and the order must survive.
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'include_archived': True, 'graph': '{"order": [["display_name", "desc"]]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0
    ${names}=    Evaluate    [i['display_name'] for i in $resp.json()['items']]
    Should Be Equal    ${names}    ${{ sorted($names, reverse=True) }}
    ...    msg=Order was dropped when the is_archived condition was prepended

Search Excluding Archived Keeps Order
    [Documentation]    The same order-only graph on the default (archived-excluded) path.
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'graph': '{"order": [["display_name", "desc"]]}'} }}
    Response Should Be Search Success    ${resp}    ${USER_SCHEMA}    size=50    page=0
    ${names}=    Evaluate    [i['display_name'] for i in $resp.json()['items']]
    Should Be Equal    ${names}    ${{ sorted($names, reverse=True) }}
    ...    msg=Order was dropped when the is_archived condition was prepended
    Search Results Should Not Contain Id    ${resp}    ${ARCHIVED_USER_ID}

Search With Invalid Include Archived Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'include_archived': 'not-a-bool'} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    include_archived

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'page': 1, 'size': 100, 'graph': '{"if":["fake-field", "=", "Owner"]}'} }}    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    fake-field

Search With Invalid Paging Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${USER_API}
    ...    params=${{ {'page': -1, 'size': 999999999999999999, 'graph': '{"if":["fake-field", "=", "Owner"]}'} }}    expected_status=any
    Response Should Be Invalid Number Range Error    ${resp}    page    size

Cors Preflight Accepted
    [Tags]    cors
    ${headers}=    Create Dictionary
    ...    Access-Control-Request-Method=GET
    ...    Access-Control-Request-Headers=Authorization, Content-Type
    ...    Origin=${API_HOST}
    ${resp}=    OPTIONS On Session    api    ${USER_API}    headers=${headers}
    ${origin}=    Evaluate    $resp.headers.get('access-control-allow-origin')
    Should Be True    ${{ $origin in ($API_HOST, '*') }}
    ...    msg=Missing/unexpected Access-Control-Allow-Origin: ${origin}
    Should Be Equal    ${resp.headers}[access-control-allow-methods]    ${CORS_ALLOW_METHODS}
    Should Be Equal    ${resp.headers}[access-control-allow-headers]    ${CORS_ALLOW_HEADERS}

Cors Preflight Rejected
    [Documentation]    No Origin header -> not a CORS request -> no CORS response headers.
    [Tags]    cors    negative
    ${headers}=    Create Dictionary
    ...    Access-Control-Request-Method=PATCH
    ...    Access-Control-Request-Headers=X-Custom-Header
    ${resp}=    OPTIONS On Session    api    ${USER_API}    headers=${headers}
    FOR    ${header}    IN
    ...    access-control-allow-origin
    ...    access-control-allow-methods
    ...    access-control-allow-headers
        Should Be True    ${{ $header not in $resp.headers }}
        ...    msg=Unexpected CORS header '${header}' on a non-CORS request
    END
