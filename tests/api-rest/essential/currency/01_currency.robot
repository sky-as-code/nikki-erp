*** Settings ***
Documentation     Currencies are a tenant catalogue answered per organization: every request names
...               the org, is_active is that org's activation, and decimal_places follows rounding.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Org Under Test
Test Tags         essential    currency


*** Test Cases ***
Create Derives Decimal Places From Rounding
    [Documentation]    Odoo's rule: decimal_places = ceil(log10(1 / rounding)); 0.05 settles in
    ...    nickels and is quoted to two digits.
    ${code}=    Generate Random String    3    [UPPER]
    ${resp}=    POST On Session    api    ${CURRENCY_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': 'Robot Currency'}, 'symbol': 'R', 'rounding': '0.05', 'symbol_position': 'before', 'is_active': False, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${CURRENCY_ID}    ${id}
    ${resp}=    GET On Session    api    ${CURRENCY_API}/${id}    params=${{ {'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[item][decimal_places]    2
    Should Be Equal    ${resp.json()}[item][is_active]    ${False}    msg=Created inactive for this org

Writing Decimal Places Is Refused
    [Tags]    negative
    ${code}=    Generate Random String    3    [UPPER]
    ${resp}=    POST On Session    api    ${CURRENCY_API}    expected_status=any
    ...    json=${{ {'code': $code, 'name': {'en-US': 'X'}, 'rounding': '0.01', 'decimal_places': 2, 'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    currency.decimal_places_server_managed

A Zero Rounding Is Refused
    [Tags]    negative
    ${code}=    Generate Random String    3    [UPPER]
    ${resp}=    POST On Session    api    ${CURRENCY_API}    expected_status=any
    ...    json=${{ {'code': $code, 'name': {'en-US': 'X'}, 'rounding': '0', 'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    currency.rounding_not_positive

A Request Without An Organization Is Refused
    [Documentation]    The activation is per organization, so a currency request without org_id
    ...    has no answer to give.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${CURRENCY_API}    params=${{ {'size': 1} }}    expected_status=any
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    err_org_id_required

Activating For The Organization Is Written Through Update
    ${resp}=    GET On Session    api    ${CURRENCY_API}/${CURRENCY_ID}    params=${{ {'org_id': $UOM_ORG_ID} }}
    ${etag}=    Set Variable    ${resp.json()}[item][etag]
    ${resp}=    PATCH On Session    api    ${CURRENCY_API}/${CURRENCY_ID}
    ...    json=${{ {'is_active': True, 'etag': $etag, 'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    ${resp}=    GET On Session    api    ${CURRENCY_API}/${CURRENCY_ID}    params=${{ {'org_id': $UOM_ORG_ID} }}
    Should Be Equal    ${resp.json()}[item][is_active]    ${True}

Search Sorts And Filters On The Activation
    [Documentation]    is_active is a SQL-computed field, so it sorts and filters like a column.
    ${resp}=    GET On Session    api    ${CURRENCY_API}
    ...    params=${{ {'size': 50, 'org_id': $UOM_ORG_ID, 'graph': '{"order":[["is_active","desc"],["code","asc"]]}'} }}
    Response Status Should Be    ${resp}    200
    ${flags}=    Evaluate    [item['is_active'] for item in $resp.json()['items']]
    Should Be Equal    ${flags}    ${{ sorted($flags, reverse=True) }}
    ${resp}=    GET On Session    api    ${CURRENCY_API}
    ...    params=${{ {'size': 50, 'org_id': $UOM_ORG_ID, 'graph': '{"if":["is_active","=",true]}'} }}
    Response Status Should Be    ${resp}    200
    ${codes}=    Evaluate    [item['code'] for item in $resp.json()['items']]
    Should Contain    ${codes}    VND
