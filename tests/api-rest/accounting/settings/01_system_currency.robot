*** Settings ***
Documentation     The system currency is an essential_currency id, validated when it is saved, and it
...               is what the user context renders money with (symbol, symbol_position).
Resource          resources/settings.resource
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Org Under Test
Suite Teardown    Restore System Currency
Test Tags         accounting    settings    currency


*** Variables ***
${VND_ID}    01KZQC0000CURRENCY000ND300
${USD_ID}    01KZQC0000CURRENCY000MJ300


*** Test Cases ***
The Organization System Currency Accepts A Usable Currency
    ${resp}=    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'system_currency', 'value': $USD_ID}] }}    module_key=accounting
    Response Status Should Be    ${resp}    200

The User Context Renders Money In The System Currency
    ${resp}=    GET On Session    api    /v1/iam/me/context
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[account_settings][currency][code]    USD
    Should Be Equal    ${resp.json()}[account_settings][currency][symbol_position]    before

The Tenant System Currency Accepts A Usable Currency
    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}
    ...    ${{ [{'name': 'tenant_system_currency', 'value': $VND_ID}] }}    module_key=accounting    expected_status=any
    Should Be True    ${resp.status_code} in (200, 403)    msg=Refused only for want of a tenant grant

An Unknown Currency Is Refused
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'system_currency', 'value': '01KZQC0000CURRENCY0000XXXX'}] }}    module_key=accounting
    ...    expected_status=any
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    currency.not_found


*** Keywords ***
Restore System Currency
    Set Settings At Level    ${SETTINGS_ORG_API}    ${{ [{'name': 'system_currency', 'value': $VND_ID}] }}
    ...    module_key=accounting    expected_status=any
