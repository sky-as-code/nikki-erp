*** Settings ***
Documentation     Creating Product Attributes. The first test saves the attribute under test
...               (${PRODUCT_ATTRIBUTE_ID}/${PRODUCT_ATTRIBUTE_ETAG}) consumed by the later
...               suites and deleted last by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    product_attribute    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Size Attribute
    ${code}=    Unique Code    size
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'data_type': 'option', 'variant_creation_mode': 'instant', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ID}    ${id}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_ETAG}    ${etag}
    Set Global Variable    ${PRODUCT_ATTRIBUTE_CODE}    ${code}

Create Applies The Declared Enum Defaults
    [Documentation]    BR §6.5.3 / §14.3 step 2: data_type defaults to `option` and
    ...    variant_creation_mode defaults to `instant`. An attribute that silently defaulted
    ...    to `never` would be dropped from every combination key without anyone asking for
    ...    that, so the defaults must be exactly these two values rather than merely "some
    ...    value".
    ${name}=    Unique Display Name    Robot Default Attribute
    ${code}=    Unique Code    default
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_attribute.json    200
    Should Be Equal    ${item}[data_type]    option
    Should Be Equal    ${item}[variant_creation_mode]    instant
    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${id}    expected_status=any

Create With Each Variant Creation Mode Succeeds
    [Documentation]    BR §14.3 step 2: instant, dynamic and never are all legal at create
    ...    time — dynamic and never are not error states, they are deliberate choices that
    ...    change how (or whether) the attribute takes part in the combination key.
    FOR    ${mode}    IN    instant    dynamic    never
        ${name}=    Unique Display Name    Robot ${mode} Attribute
        ${code}=    Unique Code    ${mode}
        ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
        ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'data_type': 'option', 'variant_creation_mode': $mode, 'org_id': $INV_ORG_ID} }}
        ${id}    ${etag}=    Response Should Be Create Success    ${resp}
        DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${id}    expected_status=any
    END

Create With Invalid Data Type Fails
    [Documentation]    BR §6.5.3: data_type is a closed enum (option/text/number/date/
    ...    boolean); an out-of-range value must not be silently accepted. No dedicated
    ...    assertion keyword pins the enum contract, so this only asserts the create did
    ...    not succeed rather than guessing the exact error key.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Data Type
    ${code}=    Unique Code    baddatatype
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'data_type': 'not_a_type', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=data_type must reject a value outside option/text/number/date/boolean

Create With Invalid Variant Creation Mode Fails
    [Documentation]    BR §14.3 step 2: variant_creation_mode is a closed enum (instant/
    ...    dynamic/never); an out-of-range value must not be silently accepted. No dedicated
    ...    assertion keyword pins the enum contract, so this only asserts the create did not
    ...    succeed rather than guessing the exact error key.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Variant Mode
    ${code}=    Unique Code    badvariantmode
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'variant_creation_mode': 'sometimes', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=variant_creation_mode must reject a value outside instant/dynamic/never

Create With Duplicate Code And Org Fails
    [Documentation]    composite_uniques is (code, org_id): an attribute code is unique per
    ...    org, not globally, since two orgs may each define their own COLOR attribute.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Attribute
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $PRODUCT_ATTRIBUTE_CODE, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    code    org_id

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    name    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${code}=    Unique Code    nonexist
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
