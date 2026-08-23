*** Settings ***
Documentation     Settings-module helpers. The module stores two things: the schema a feature
...               module registers to declare what it can be configured with, and one row per
...               setting value per owner. Both are reachable as plain resources, but the
...               behaviour worth testing lives on the per-level routes.
...
...               Fixtures lean on what Essential registers at boot (theme_mode and language)
...               rather than creating a schema of their own, because a hand-made schema would
...               not exercise the registration path the product actually uses.
...               All keywords expect an authorized session with alias "api".
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource
Resource          ${CURDIR}/iam.resource


*** Variables ***
${SETTINGS_SCHEMA_API}      /v1/settings/settings_schema
${SETTINGS_RECORD_API}      /v1/settings/settings_record
${SETTINGS_TENANT_API}      /v1/settings/tenant
${SETTINGS_ORG_API}         /v1/settings/org
${SETTINGS_USER_API}        /v1/settings/user
${SETTINGS_SCHEMA_DIR}      ${CURDIR}/schemas/settings

# The module every settings fixture reads. Essential registers theme_mode and language at boot,
# so this key exists in any running system without the suite creating it.
${SETTINGS_MODULE_KEY}      essential
${SETTING_THEME_MODE}       theme_mode
${SETTING_LANGUAGE}         language
${SETTING_TIMEZONE}         timezone

# Essential's organization-level settings: an organization's own conventions.
${SETTING_SYSTEM_LOCALE}       system_locale
${SETTING_SYSTEM_TIMEZONE}     system_timezone
${SETTING_DEFAULT_CURRENCY}    default_currency

# iam's tenant-level settings. They are a different module, so a tenant-level read must name iam
# rather than essential — essential declares nothing at the tenant level.
${IAM_MODULE_KEY}              iam
${SETTING_SESSION_TIMEOUT}     session_timeout_minutes
${SETTING_REQUIRE_MFA}         require_mfa


*** Keywords ***
Ensure Settings Schema Under Test
    [Documentation]    Resolves the schema Essential registered, rather than creating one.
    ...    Registration is idempotent and runs on every boot, so a missing row here means the
    ...    registration path itself is broken — which is worth failing on.
    ${id}=    Get Variable Value    ${SETTINGS_SCHEMA_ID}    ${EMPTY}
    IF    $id    RETURN
    ${resp}=    GET On Session    api    ${SETTINGS_SCHEMA_API}    params=${{ {'size': 200} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    FOR    ${item}    IN    @{items}
        ${module_key}=    Get From Dictionary    ${item}    module_key    ${EMPTY}
        ${level}=    Get From Dictionary    ${item}    level    ${EMPTY}
        IF    "${module_key}" == "${SETTINGS_MODULE_KEY}" and "${level}" == "user"
            Set Global Variable    ${SETTINGS_SCHEMA_ID}    ${item}[id]
            RETURN
        END
    END
    Fail    msg=Essential registered no user-level settings schema; the boot-time registration did not run.

Get Settings At Level
    [Documentation]    Reads one module's settings at one level for the acting caller.
    [Arguments]    ${level_api}    ${module_key}=${SETTINGS_MODULE_KEY}    ${alias}=api    ${expected_status}=${NONE}
    ${resp}=    GET On Session    ${alias}    ${level_api}/${module_key}    expected_status=${expected_status}
    RETURN    ${resp}

Set Settings At Level
    [Documentation]    Writes a partial set of changed items. Only the items named here are
    ...    touched; everything else keeps its stored value.
    [Arguments]    ${level_api}    ${items}    ${module_key}=${SETTINGS_MODULE_KEY}    ${alias}=api    ${expected_status}=${NONE}
    ${resp}=    PATCH On Session    ${alias}    ${level_api}/${module_key}
    ...    json=${{ {'items': $items} }}    expected_status=${expected_status}
    RETURN    ${resp}

Setting Item Should Be
    [Documentation]    Asserts one named item of a settings read carries the expected value.
    [Arguments]    ${response}    ${name}    ${expected_value}
    ${item}=    Find Setting Item    ${response}    ${name}
    Should Be Equal    ${item}[value]    ${expected_value}
    ...    msg=Setting '${name}' should read as '${expected_value}'

Find Setting Item
    [Documentation]    Returns one named item of a settings read, failing clearly when the
    ...    schema declared it but the read dropped it.
    [Arguments]    ${response}    ${name}
    FOR    ${item}    IN    @{response.json()}[items]
        IF    "${item}[name]" == "${name}"    RETURN    ${item}
    END
    Fail    msg=Setting '${name}' is missing from the response; the schema declares it.

Read Own Theme Mode
    [Documentation]    The acting user's stored theme, used to restore it after a test that
    ...    changes it — a preference is per-user state that outlives the suite.
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}
    Response Status Should Be    ${resp}    200
    ${item}=    Find Setting Item    ${resp}    ${SETTING_THEME_MODE}
    RETURN    ${item}[value]

Restore Own Theme Mode
    [Documentation]    Puts back a preference a test changed, so re-running the suite starts
    ...    from the same place it did the first time.
    [Arguments]    ${value}
    IF    not $value    RETURN
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': $SETTING_THEME_MODE, 'value': $value}] }}    expected_status=any

Delete Settings Fixtures
    [Documentation]    Removes what the suites created. Records go before schemas: a record
    ...    names the schema it was validated against.
    ${theme}=    Get Variable Value    ${SETTINGS_ORIGINAL_THEME}    ${EMPTY}
    IF    $theme    Restore Own Theme Mode    ${theme}
    FOR    ${name}    ${api}    IN
    ...    SETTINGS_RECORD_ID    ${SETTINGS_RECORD_API}
        ${id}=    Get Variable Value    ${${name}}    ${EMPTY}
        IF    $id
            DELETE On Session    api    ${api}/${id}    expected_status=any
            Set Global Variable    ${${name}}    ${EMPTY}
        END
    END
    Delete Permission Fixtures
