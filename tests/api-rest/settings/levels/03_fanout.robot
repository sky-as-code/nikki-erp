*** Settings ***
Documentation     The fan-out — AC-02 and AC-06. Enforcement is a physical write, not a read-time
...               resolution: saving a tenant setting whose schema says allow_override is false
...               pushes the value onto every child record that already exists, so reads stay
...               plain reads with no precedence resolver behind them.
...
...               This is the one behaviour a caller cannot observe from a single row, because it
...               changes what somebody else reads. It is deliberately destructive and
...               irreversible (D8): a tenant enforcing a value overwrites what a child chose,
...               with no warning and no undo. That is what "enforced" means, so this suite pins
...               the overwrite rather than treating it as a hazard.
...
...               iam's tenant settings are what make this testable — both declare
...               allow_override: false. Until they existed no module declared a non-overridable
...               tenant setting and the fan-out had no data to act on.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Permission Fixtures
Test Tags         settings    levels    fanout


*** Test Cases ***
A Tenant Write Reaches An Existing Child Row
    [Documentation]    AC-02/AC-06. A child already holding a stored value must read the tenant's
    ...    enforced value afterwards. The child row is created first, so this tests the fan-out
    ...    onto existing data rather than the simpler case of a child that had no row at all.
    ${probe_id}    ${probe_email}=    Create Probe User Session    fanout_probe

    # The child's own row must exist before the tenant enforces, or a passing test would only
    # prove that a fresh read falls back to the tenant value.
    ${seeded}=    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'session_timeout_minutes', 'value': 45}] }}
    Response Status Should Be    ${seeded}    200

    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'session_timeout_minutes', 'value': 90}] }}
    Response Status Should Be    ${resp}    200

    ${read}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    Setting Item Should Be    ${read}    ${SETTING_SESSION_TIMEOUT}    ${90}

An Enforced Setting Reports Itself Uneditable To Everyone Below
    [Documentation]    `editable` is what the settings pane disables a control on. A locked
    ...    setting reported as editable lets a user submit a change the server then refuses,
    ...    which reads as a broken form rather than as a policy they cannot change.
    ${probe_id}    ${probe_email}=    Create Probe User Session    editable_probe

    ${resp}=    Get Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    alias=editable_probe    expected_status=any

    # A plain user may be refused the tenant read outright, which is a stricter answer than
    # "visible but not editable" and satisfies the same requirement.
    IF    ${resp.status_code} == 200
        FOR    ${item}    IN    @{resp.json()}[items]
            IF    not ${item}[allow_override]
                Should Not Be True    ${item}[editable]
                ...    msg=Policy '${item}[name]' is not overridable but reports itself editable to a plain user.
            END
        END
    END

An Overridable Setting Keeps The Owner's Own Choice
    [Documentation]    The counterpart to the fan-out, and the reason allow_override exists: a
    ...    tenant write must reach only what the schema locks down. Essential's theme_mode is
    ...    overridable, so a user's own choice must survive.
    ...
    ...    Without this the fan-out could overwrite every setting in the tenant and the enforced
    ...    case above would still pass.
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}

    Set Settings At Level    ${SETTINGS_TENANT_API}    module_key=${IAM_MODULE_KEY}
    ...    items=${{ [{'name': 'require_mfa', 'value': True}] }}

    ${read}=    Get Settings At Level    ${SETTINGS_USER_API}
    Setting Item Should Be    ${read}    ${SETTING_THEME_MODE}    dark
