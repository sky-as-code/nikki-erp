*** Settings ***
Documentation     Essential Sequence helpers: the definition under test, a definition that has
...               issued a number, and the routes of the three Sequence resources.
...
...               NextValue has no REST route (BR §6.1): a number is only ever issued by a
...               consumer. The "issued" fixture therefore goes through Inventory: the org's
...               inventory.transfer definition (registered by the org lifecycle or the seeds) issues
...               a fresh operation type's first number on that type's own counter when one transfer
...               of it is created. All keywords expect an authorized session with alias "api".
Library           Collections
Library           String
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource
Resource          ${CURDIR}/inventory.resource


*** Variables ***
${SEQ_DEFINITION_API}       /v1/essential/essential_sequence_definition
${SEQ_PREVIEW_API}          /v1/essential/essential_sequence_definition/preview
${SEQ_COUNTER_API}          /v1/essential/essential_sequence_counter
${SEQ_ISSUANCE_API}         /v1/essential/essential_sequence_issuance
${SEQ_SCHEMA_DIR}           ${CURDIR}/schemas/essential
# BR §12.1 error keys, verbatim from modules/essential/domain/services/sequence_errors.go.
${SEQ_ERR_INVALID_TEMPLATE}     essential:err_sequence_invalid_template
${SEQ_ERR_INVALID_TIMEZONE}     essential:err_sequence_invalid_timezone
${SEQ_ERR_INVALID_INPUT}        essential:err_sequence_invalid_input
${SEQ_ERR_PADDING}              essential:err_sequence_padding
${SEQ_ERR_READ_ONLY_FIELD}      essential:err_sequence_read_only_field
${SEQ_ERR_CONFIG_LOCKED}        essential:err_sequence_configuration_locked
${SEQ_ERR_USED_BY_MODULES}      common:resource_used_by_modules
# Crockford Base32: digits and uppercase letters without I, L, O and U.
${CROCKFORD_CHAR}               [0-9A-HJKMNP-TV-Z]


*** Keywords ***
Ensure Sequence Org
    [Documentation]    ORG_ID from the environment, else the first organization the API returns.
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${SEQ_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${SEQ_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    Should Not Be Empty    ${resp.json()}[items]
    Set Global Variable    ${SEQ_ORG_ID}    ${resp.json()}[items][0][id]

Unique Sequence Code
    [Documentation]    (code, org_id) is unique, so every code carries a hash. Codes are
    ...    lowercase dotted segments: ^[a-z0-9_]+(\.[a-z0-9_]+)*$.
    [Arguments]    ${words}=robot.seq
    ${suffix}=    Unique Suffix
    RETURN    ${words}.${suffix}

Sequence Definition Body
    [Documentation]    A valid org-scoped definition owned by "robot", merged with the overrides.
    ...    An override set to None removes the key, so a test can leave a field out on purpose.
    [Arguments]    &{overrides}
    Ensure Sequence Org
    ${code}=    Unique Sequence Code
    ${name}=    Unique Display Name    Robot Sequence
    ${body}=    Evaluate
    ...    {k: v for k, v in {'configuration_scope': 'org', 'org_id': $SEQ_ORG_ID, 'owner_module': 'robot', 'code': $code, 'name': {'en-US': $name}, **$overrides}.items() if v is not None}
    RETURN    ${body}

Post Sequence Definition
    [Documentation]    POSTs a definition built by Sequence Definition Body; returns the response.
    [Arguments]    &{overrides}
    ${body}=    Sequence Definition Body    &{overrides}
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}    json=${body}    expected_status=any
    RETURN    ${resp}

Create Sequence Definition
    [Documentation]    Creates a definition and returns (id, etag).
    [Arguments]    &{overrides}
    ${resp}=    Post Sequence Definition    &{overrides}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${etag}

Read Sequence Definition
    [Documentation]    GETs one definition and returns its item.
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_definition.json    200
    RETURN    ${item}

Ensure Sequence Definition Under Test
    [Documentation]    The definition the definition suite exercises: AC05's configuration
    ...    (initial 0, increment 2, padding 3) behind a "RB" prefix. It never issues a number, so
    ...    its configuration stays editable.
    ${id}=    Get Variable Value    ${SEQ_DEFINITION_ID}    ${EMPTY}
    IF    $id    RETURN
    ${id}    ${etag}=    Create Sequence Definition
    ...    prefix_template=RB    padding=${3}    initial_value=${0}    increment=${2}
    Set Global Variable    ${SEQ_DEFINITION_ID}    ${id}
    Set Global Variable    ${SEQ_DEFINITION_ETAG}    ${etag}

Ensure Issued Sequence Definition
    [Documentation]    A definition that has issued numbers: the org's inventory.transfer definition
    ...    (SEQ-FB-T02), after one transfer of a fresh operation type, whose own counter on it then
    ...    stands at 1. Sets SEQ_ISSUED_DEFINITION_ID, the type (SEQ_ISSUED_TYPE_ID, the scope object
    ...    of that counter), the transfer (SEQ_ISSUED_TRANSFER_ID) and its number.
    ${id}=    Get Variable Value    ${SEQ_ISSUED_DEFINITION_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Inventory Location Under Test
    Ensure Stock Destination Location Under Test
    ${name}=    Unique Display Name    Robot Sequence Type
    ${resp}=    Post Stock Operation Type    ${{ {'name': {'en-US': $name}, 'operation_code': 'internal', 'reservation_method': 'manual', 'backorder_policy': 'never', 'org_id': $INV_ORG_ID} }}
    ${type_id}    ${type_etag}=    Response Should Be Create Success    ${resp}
    ${transfer_id}    ${transfer_etag}=    Create Stock Transfer    ${type_id}
    ${transfer}=    GET On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${transfer}    200
    Set Global Variable    ${SEQ_ISSUED_TYPE_ID}    ${type_id}
    Set Global Variable    ${SEQ_ISSUED_TRANSFER_ID}    ${transfer_id}
    Set Global Variable    ${SEQ_ISSUED_TRANSFER_NUMBER}    ${transfer.json()}[item][transfer_number]
    ${definition_id}=    Org Transfer Sequence Definition Id
    Set Global Variable    ${SEQ_ISSUED_DEFINITION_ID}    ${definition_id}

Org Transfer Sequence Definition Id
    [Documentation]    The one definition numbering the transfers of INV_ORG_ID, every operation
    ...    type on its own counter. The org lifecycle (or the seeds) registered it.
    ${graph}=    Evaluate
    ...    __import__('json').dumps({'and': [{'if': ['code', '=', 'inventory.transfer']}, {'if': ['org_id', '=', $INV_ORG_ID]}]})
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}    params=${{ {'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    1    msg=An org has exactly one inventory.transfer definition
    RETURN    ${resp.json()}[items][0][id]

Type Sequence Counter
    [Documentation]    The counter an operation type numbers its transfers on: the org's transfer
    ...    definition's counter whose scope object is the type. Fails unless exactly one exists.
    [Arguments]    ${type_id}
    ${resp}=    Search Sequence Records    ${SEQ_COUNTER_API}    scope_object_id    ${type_id}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    1
    RETURN    ${resp.json()}[items][0]

New Idempotency Id
    [Documentation]    A canonical UUIDv7, the only form the X-Idempotency-ID header accepts.
    ${id}=    Evaluate    (lambda ms, r: '%08x-%04x-7%03x-%04x-%012x' % (ms >> 16, ms & 0xffff, r >> 50 & 0xfff, 0x8000 | (r >> 36 & 0x3fff), r & 0xffffffffffff))(int(__import__('time').time() * 1000), __import__('random').getrandbits(62))
    RETURN    ${id}

Search Sequence Records
    [Documentation]    Searches a Sequence resource with one equality filter; returns the response.
    [Arguments]    ${api}    ${field}    ${value}    ${size}=50
    ${graph}=    Evaluate    __import__('json').dumps({'if': [$field, '=', $value]})
    ${resp}=    GET On Session    api    ${api}    params=${{ {'graph': $graph, 'size': $size} }}    expected_status=any
    RETURN    ${resp}

Response Should Be Sequence Error
    [Documentation]    One client error with the given key on the given field (an empty field
    ...    asserts none). Sequence errors carry their reason in vars, so only key, field and the
    ...    400 status are contract here.
    [Arguments]    ${resp}    ${key}    ${field}=${EMPTY}
    Response Status Should Be    ${resp}    400
    ${errors}=    Set Variable    ${resp.json()}
    ${keys}=    Evaluate    [(item.get('key'), item.get('field') or '') for item in $errors]
    Should Contain    ${keys}    ${{ ($key, $field) }}
    ...    msg=Expected ${key} on '${field}', got ${errors}

Response Should Carry No Secret
    [Documentation]    algorithm_params holds the FF1 key: it must never leave the server, in any
    ...    record of the response (BR §3, SEQ-004).
    [Arguments]    ${resp}
    Should Not Contain    ${resp.text}    "algorithm_params":
    ...    msg=algorithm_params leaked in ${resp.text}

Archive Sequence Definition
    [Documentation]    Definitions cannot be deleted; a throwaway one is archived instead.
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${id}    expected_status=any
    IF    $resp.status_code != 200    RETURN
    POST On Session    api    ${SEQ_DEFINITION_API}/${id}/archived
    ...    json=${{ {'etag': $resp.json()['item']['etag'], 'is_archived': True} }}    expected_status=any

Delete Sequence Fixtures
    [Documentation]    Best-effort cleanup. The issued fixture's transfer goes first (it holds the
    ...    type), then the type; the org's transfer definition stays, as the org still numbers with it.
    ...    Definitions themselves have no delete route (BR §4.4): the free ones are archived instead
    ...    and stay as history.
    ${transfer}=    Get Variable Value    ${SEQ_ISSUED_TRANSFER_ID}    ${EMPTY}
    IF    $transfer
        DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer}
        ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    END
    ${type}=    Get Variable Value    ${SEQ_ISSUED_TYPE_ID}    ${EMPTY}
    IF    $type
        DELETE On Session    api    ${STOCK_OPERATION_TYPE_API}/${type}
        ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    END
    Set Global Variable    ${SEQ_ISSUED_TRANSFER_ID}    ${EMPTY}
    Set Global Variable    ${SEQ_ISSUED_TYPE_ID}    ${EMPTY}
    Set Global Variable    ${SEQ_ISSUED_DEFINITION_ID}    ${EMPTY}
    ${definition}=    Get Variable Value    ${SEQ_DEFINITION_ID}    ${EMPTY}
    IF    $definition
        ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${definition}    expected_status=any
        IF    $resp.status_code == 200 and not $resp.json()['item'].get('is_archived')
            POST On Session    api    ${SEQ_DEFINITION_API}/${definition}/archived
            ...    json=${{ {'etag': $resp.json()['item']['etag'], 'is_archived': True} }}    expected_status=any
        END
    END
    Set Global Variable    ${SEQ_DEFINITION_ID}    ${EMPTY}
    Delete Inventory Seed Data
