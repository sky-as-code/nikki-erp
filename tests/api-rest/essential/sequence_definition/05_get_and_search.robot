*** Settings ***
Documentation     Reading definitions: GET by id, search filters, paging and EXISTS. The FF1 key
...               (algorithm_params) never appears in any read, and cannot be filtered or sorted.
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Sequence Definition Under Test
Test Tags         essential    sequence_definition    search


*** Test Cases ***
Get Succeeds
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${SEQ_DEFINITION_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_definition.json    200
    Should Be Equal    ${item}[id]    ${SEQ_DEFINITION_ID}
    Response Should Carry No Secret    ${resp}

Get Of A Pseudorandom Definition Carries No Key
    ${id}    ${etag}=    Create Sequence Definition    algorithm=pseudorandom    prefix_template=CORE
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${id}
    Response Status Should Be    ${resp}    200
    Response Should Carry No Secret    ${resp}
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${id}
    ...    params=${{ {'fields': 'id,algorithm,algorithm_params'} }}    expected_status=any
    Response Should Carry No Secret    ${resp}
    [Teardown]    Archive Sequence Definition    ${id}

Get With Not Found Id Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/${NOT_FOUND_ID}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Get With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}/not-existing-1234567890123    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Search By Code Finds The Definition
    ${item}=    Read Sequence Definition    ${SEQ_DEFINITION_ID}
    ${resp}=    Search Sequence Records    ${SEQ_DEFINITION_API}    code    ${item}[code]
    Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_definition.json    50    item_count=1
    Should Be Equal    ${resp.json()}[items][0][id]    ${SEQ_DEFINITION_ID}
    Response Should Carry No Secret    ${resp}

Search By Owner Module Returns Only That Module
    ${resp}=    Search Sequence Records    ${SEQ_DEFINITION_API}    owner_module    robot
    Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_definition.json    50
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Be Equal    ${item}[owner_module]    robot
    END

Search Lists Every Consumer's Definitions
    [Documentation]    The consumers register theirs on org init (and in the seeds): Sales
    ...    (SO), Purchase (PO/PA), Payment invoice and Inventory scrap.
    FOR    ${code}    IN    sales.order    sales.order.public_reference    purchase.order
    ...    purchase.agreement    paymentinvoice.invoice    inventory.scrap
        ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}
        ...    params=${{ {'graph': __import__('json').dumps({'and': [{'if': ['code', '=', $code]}, {'if': ['org_id', '=', $SEQ_ORG_ID]}]}), 'size': 5} }}
        Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_definition.json    5    item_count=1
        Response Should Carry No Secret    ${resp}
    END

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}    params=${{ {'page': 0, 'size': 2} }}
    Response Should Be Search Success    ${resp}    ${SEQ_SCHEMA_DIR}/sequence_definition.json    2    item_count=2

Search Projecting The Secret Field Returns No Secret
    ${resp}=    GET On Session    api    ${SEQ_DEFINITION_API}
    ...    params=${{ {'fields': 'id,code,algorithm_params', 'size': 200} }}    expected_status=any
    Response Should Carry No Secret    ${resp}

Search Filtering On The Secret Field Fails
    [Tags]    negative
    ${resp}=    Search Sequence Records    ${SEQ_DEFINITION_API}    algorithm_params    x
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_READ_ONLY_FIELD}    graph

Exists With Mixed Ids Succeeds
    ${resp}=    POST On Session    api    ${SEQ_DEFINITION_API}/exists
    ...    json=${{ {'ids': [$SEQ_DEFINITION_ID, $NOT_FOUND_ID]} }}
    Response Should Be Exists Success    ${resp}    existing=1    not_existing=1
