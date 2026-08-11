*** Settings ***
Documentation     Creating Product Attribute Values. The first test saves the value under test
...               (${ATTRIBUTE_VALUE_ID}/${ATTRIBUTE_VALUE_ETAG}) consumed by the later suites
...               and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...    AND    Ensure Product Attribute Under Test
Test Tags         inventory    product_attribute_value    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Black
    ${code}=    Unique Code    val
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $PRODUCT_ATTRIBUTE_ID, 'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${ATTRIBUTE_VALUE_ID}    ${id}
    Set Global Variable    ${ATTRIBUTE_VALUE_ETAG}    ${etag}
    Set Global Variable    ${ATTRIBUTE_VALUE_CODE}    ${code}

Create With Positive Price Extra Succeeds
    [Documentation]    price_extra is a surcharge: the common case is a positive addition to
    ...    the base price. Decimals travel as strings.
    ${name}=    Unique Display Name    Robot Premium
    ${code}=    Unique Code    val
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $PRODUCT_ATTRIBUTE_ID, 'code': $code, 'name': {'en-US': $name}, 'price_extra': '12.5', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${id}    expected_status=any

Create With Negative Price Extra Succeeds
    [Documentation]    price_extra is signed (BR: "a value may also discount"), so a negative
    ...    surcharge is a legitimate discount, not an error.
    ${name}=    Unique Display Name    Robot Discount
    ${code}=    Unique Code    val
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $PRODUCT_ATTRIBUTE_ID, 'code': $code, 'name': {'en-US': $name}, 'price_extra': '-5.5', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_attribute_value.json    200
    Should Be Equal As Numbers    ${item}[price_extra]    -5.5
    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${id}    expected_status=any

Create With Out Of Range Price Extra Fails
    [Documentation]    price_extra is bounded to +/-1000000000000; a caller sending beyond
    ...    that must be rejected rather than silently truncated.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Out Of Range
    ${code}=    Unique Code    val
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $PRODUCT_ATTRIBUTE_ID, 'code': $code, 'name': {'en-US': $name}, 'price_extra': '2000000000000', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=price_extra beyond the declared max must be rejected. Response body: ${resp.text}

Create With Duplicate Code Under Same Attribute Fails
    [Documentation]    composite_uniques is ["attribute_id", "code"] — code is unique per
    ...    ATTRIBUTE, not per org, so the same code under the same attribute must collide.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Value
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $PRODUCT_ATTRIBUTE_ID, 'code': $ATTRIBUTE_VALUE_CODE, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    attribute_id    code

Create With Same Code Under Different Attribute Succeeds
    [Documentation]    The composite unique is scoped by attribute, not by org: "Red" belonging
    ...    to Color and to Trim are genuinely different values, so reusing the value-under-test
    ...    code on a fresh attribute must be accepted.
    Ensure Inventory Org
    ${attr_name}=    Unique Display Name    Robot Trim Attribute
    ${attr_code}=    Unique Code    trim
    ${attr_resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $attr_code, 'name': {'en-US': $attr_name}, 'data_type': 'option', 'variant_creation_mode': 'instant', 'org_id': $INV_ORG_ID} }}
    ${attr_id}    ${attr_etag}=    Response Should Be Create Success    ${attr_resp}
    ${name}=    Unique Display Name    Robot Reused Code
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $attr_id, 'code': $ATTRIBUTE_VALUE_CODE, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${id}    expected_status=any
    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${attr_id}    expected_status=any

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    attribute_id    code    name    org_id

Create With Not Found Attribute Id Fails
    [Documentation]    A value cannot exist outside an attribute; a caller pointing at an
    ...    attribute that does not exist must be rejected rather than orphaning the value.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Orphan Value
    ${code}=    Unique Code    val
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $NOT_FOUND_ID, 'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A not-found attribute_id must not be creatable against. Response body: ${resp.text}

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${code}=    Unique Code    val
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $PRODUCT_ATTRIBUTE_ID, 'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
