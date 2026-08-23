*** Settings ***
Documentation     Tenant-level settings. iam declares two: how long a session may stay idle, and
...               whether everyone must sign in with a second factor.
...
...               Both are declared allow_override: false, which is the point of them. A session
...               timeout or an MFA requirement an individual could opt out of is not a policy,
...               it is a suggestion — so these are the settings that exercise the enforcement
...               fan-out, which no setting could before iam declared them.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    levels    tenant


*** Test Cases ***
Read Tenant Settings Returns Every Declared Setting
    [Documentation]    The tenant level belongs to iam, not essential: essential declares nothing
    ...    a tenant configures, so reading essential here correctly returns an empty section.
    ${resp}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    Response Status Should Be    ${resp}    200
    Find Setting Item    ${resp}    ${SETTING_SESSION_TIMEOUT}
    Find Setting Item    ${resp}    ${SETTING_REQUIRE_MFA}

Essential Declares Nothing At The Tenant Level
    [Documentation]    A module that registered no schema for a level contributes no section.
    ...    That is an ordinary outcome, not a missing resource, so the read is a 200 with no items.
    ${resp}=    Get Settings At Level    ${SETTINGS_TENANT_API}
    Response Status Should Be    ${resp}    200
    Should Be Empty    ${resp.json()}[items]

Every Tenant Setting Carries A Translated Description
    ${resp}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        ${field}=    Get From Dictionary    ${item}    field
        ${description}=    Get From Dictionary    ${field}    description    ${NONE}
        Should Not Be Equal    ${description}    ${NONE}
        ...    msg=Setting '${item}[name]' carries no description.
        Should Be Equal    ${description}[$ref]    settings_desc.${item}[name]
        ...    msg=The description key convention is settings_desc.<setting name>.
    END

Tenant Settings Are Not Overridable
    [Documentation]    allow_override false is what makes these a policy rather than a default,
    ...    and it is what the fan-out keys off. If either flipped to true, a tenant could no
    ...    longer enforce it and the test below would stop meaning anything.
    ${resp}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        Should Not Be True    ${item}[allow_override]
        ...    msg=Tenant policy '${item}[name]' must not be overridable.
    END

Write A Tenant Setting Succeeds
    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'session_timeout_minutes', 'value': 30}] }}
    Response Status Should Be    ${resp}    200

    ${read}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    Setting Item Should Be    ${read}    ${SETTING_SESSION_TIMEOUT}    ${30}

A Boolean Tenant Setting Round Trips
    [Documentation]    require_mfa is the only boolean setting the product ships. A boolean that
    ...    came back as the string "true" would be truthy in the frontend either way, so the type
    ...    is asserted rather than the truthiness.
    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'require_mfa', 'value': True}] }}

    ${read}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ${item}=    Find Setting Item    ${read}    ${SETTING_REQUIRE_MFA}
    Should Be True    isinstance($item['value'], bool)
    ...    msg=require_mfa must read back as a boolean, not as a string.
    Should Be Equal    ${item}[value]    ${True}

A Session Timeout Above Its Ceiling Is Refused
    [Documentation]    The declared ceiling is a week. A session outliving that is indistinguishable
    ...    from no timeout at all, which is not something a tenant should be able to configure by
    ...    typing a large number.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'session_timeout_minutes', 'value': 999999}] }}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=A session timeout past the declared maximum must be refused.

A Session Timeout Of Zero Is Refused
    [Documentation]    Zero would mean "expire immediately", locking every account in the tenant
    ...    out through the very interface that set it.
    ...
    ...    This needs its own assertion because the platform does NOT catch it: ModelField.Validate
    ...    treats a numeric zero as an ABSENT value (isNilOrEmpty returns true for it) and returns
    ...    before the declared range is checked, so a submitted 0 was stored as a real timeout. The
    ...    settings module therefore enforces the declared bounds itself — a settings value arrives
    ...    from an API caller and is read back as policy, so "close enough" validation is not.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'session_timeout_minutes', 'value': 0}] }}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=A session timeout of zero must be refused; it would expire every session immediately.

    ${read}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ${item}=    Find Setting Item    ${read}    ${SETTING_SESSION_TIMEOUT}
    Should Not Be Equal    ${item}[value]    ${0}
    ...    msg=A refused write must not have stored its value.

A Non Numeric Session Timeout Is Refused
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'session_timeout_minutes', 'value': 'thirty'}] }}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=A non-numeric session timeout must be refused.
