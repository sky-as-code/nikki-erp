*** Settings ***
Documentation     Creating Brands. The first test saves the brand under test
...               (${BRAND_ID}/${BRAND_ETAG}) consumed by the later suites and deleted last
...               by 08_delete.robot.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    brand    create


*** Test Cases ***
Create With Required Fields Succeeds
    ${name}=    Unique Display Name    Robot Brand
    ${code}=    Unique Code    brand
    ${resp}=    POST On Session    api    ${BRAND_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${BRAND_ID}    ${id}
    Set Global Variable    ${BRAND_ETAG}    ${etag}
    Set Global Variable    ${BRAND_CODE}    ${code}

Create With All Optional Fields Succeeds
    [Documentation]    website is data_type `url`, not a plain string; this pins that a
    ...    well-formed URL is accepted and stored on the record.
    ${name}=    Unique Display Name    Robot Full Brand
    ${code}=    Unique Code    fullbrand
    ${resp}=    POST On Session    api    ${BRAND_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID, 'website': 'https://example.com', 'description': {'en-US': 'A robot brand'}} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${BRAND_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/brand.json    200
    Should Be Equal    ${item}[website]    https://example.com
    DELETE On Session    api    ${BRAND_API}/${id}    expected_status=any

Create With Malformed Website Fails
    [Documentation]    website is data_type `url`; no dedicated assertion keyword pins the
    ...    exact error key for this contract, so this only asserts the create is rejected.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Website Brand
    ${code}=    Unique Code    badwebsite
    ${resp}=    POST On Session    api    ${BRAND_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID, 'website': 'not a url'} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    201
    ...    msg=A malformed website URL must not be accepted on create

Create With Duplicate Code And Org Fails
    [Documentation]    composite_uniques on brand.json is ["code", "org_id"]: the same code
    ...    is only ambiguous within one org.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Brand
    ${resp}=    POST On Session    api    ${BRAND_API}
    ...    json=${{ {'code': $BRAND_CODE, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    code    org_id

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${BRAND_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    code    name    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${BRAND_API}
    ...    data={ "name": {"en-US": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Nonexist Field Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Nonexist Field
    ${code}=    Unique Code    nonexist
    ${resp}=    POST On Session    api    ${BRAND_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $name}, 'org_id': $INV_ORG_ID, 'bla_bla_field': 'x'} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
