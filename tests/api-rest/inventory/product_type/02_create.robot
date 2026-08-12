*** Settings ***
Documentation     Creating Product Types. The first test saves the type under test
...               (${PRODUCT_TYPE_ID}/${PRODUCT_TYPE_ETAG}) consumed by the later suites
...               and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    product_type    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Goods Type
    ${code}=    Unique Code    goods
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'supports_stock': True} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_TYPE_ID}    ${id}
    Set Global Variable    ${PRODUCT_TYPE_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_TYPE_CODE}    ${code}

Create Applies The Declared Capability Defaults
    [Documentation]    BR §6.3.2: sale and purchase default on, stock and manufacturing
    ...    off. A type created without flags must therefore be sellable and purchasable but
    ...    stockless — the SERVICE shape — rather than inert.
    ${name}=    Unique Display Name    Robot Service Type
    ${code}=    Unique Code    service
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_TYPE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_type.json    200
    Should Be Equal    ${item}[supports_sale]    ${True}
    Should Be Equal    ${item}[supports_purchase]    ${True}
    Should Be Equal    ${item}[supports_stock]    ${False}
    Should Be Equal    ${item}[supports_manufacturing]    ${False}
    DELETE On Session    api    ${PRODUCT_TYPE_API}/${id}    expected_status=any

Create With Duplicate Code Fails
    [Documentation]    BR §6.3.2: `code` is globally unique on a product type — processing
    ...    logic keys off it, so two types answering to GOODS would be ambiguous.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Type
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}
    ...    json=${{ {'code': $PRODUCT_TYPE_CODE, 'name': {'en-US': $name}} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    code

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    name

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${code}=    Unique Code    nonexist
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field

Create With Archived Flag Fails
    [Documentation]    is_archived arrives from core.basemodel.archivable_model and is set
    ...    through the /archived action, never as a create field. Accepting it here would
    ...    let a caller create an already-invisible record.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Prearchived Type
    ${code}=    Unique Code    prearch
    ${resp}=    POST On Session    api    ${PRODUCT_TYPE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'is_archived': True} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=is_archived must not be settable at create time
