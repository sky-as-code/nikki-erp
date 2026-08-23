*** Settings ***
Documentation     Reading the registered declarations. Registration is idempotent and runs on
...               every boot, so Essential's rows must be present in any running system.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    settings_schema    search


*** Test Cases ***
Search Returns The Registered Schemas
    ${resp}=    GET On Session    api    ${SETTINGS_SCHEMA_API}    params=${{ {'size': 200} }}
    Response Status Should Be    ${resp}    200
    Should Not Be Empty    ${resp.json()}[items]
    ...    msg=No settings schema is registered; boot-time registration did not run.

Essential Registered Its User Settings Schema
    [Documentation]    theme_mode and language are the first settings the product shipped. A
    ...    missing row here means every account-settings read falls back to nothing.
    Ensure Settings Schema Under Test
    ${resp}=    GET On Session    api    ${SETTINGS_SCHEMA_API}/${SETTINGS_SCHEMA_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${SETTINGS_SCHEMA_DIR}/settings_schema.json    200
    Should Be Equal    ${item}[module_key]    ${SETTINGS_MODULE_KEY}
    Should Be Equal    ${item}[level]         user

Registration Is Idempotent Per Module And Level
    [Documentation]    The unique key is [module_key, level]. Two rows for one pair would mean
    ...    a re-registration created a duplicate instead of matching the existing declaration,
    ...    and reads would then depend on which row came back first.
    ${resp}=    GET On Session    api    ${SETTINGS_SCHEMA_API}    params=${{ {'size': 200} }}
    Response Status Should Be    ${resp}    200
    ${pairs}=    Evaluate    [(i['module_key'], i['level']) for i in $resp.json()['items']]
    ${unique}=    Evaluate    sorted(set($pairs))
    ${all}=       Evaluate    sorted($pairs)
    Should Be Equal    ${all}    ${unique}
    ...    msg=A [module_key, level] pair is registered twice; registration is not idempotent.
