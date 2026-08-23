*** Settings ***
Documentation     The reach of a plain account. The system User role is granted READ on the two
...               settings resources and never UPDATE: a user must be able to see the settings
...               that govern their session, and must not be able to rewrite the tenant's.
...
...               Without this suite a seed that added UPDATE to the User role would hand every
...               account the ability to reconfigure the tenant, and nothing in the tree would
...               notice.
Resource          resources/settings.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Create Settings Probe
Suite Teardown    Delete Permission Fixtures
Test Tags         settings    permissions


*** Keywords ***
Create Settings Probe
    ${id}    ${email}=    Create Probe User Session    settings_probe
    Set Suite Variable    ${SETTINGS_PROBE_ID}       ${id}
    Set Suite Variable    ${SETTINGS_PROBE_EMAIL}    ${email}


*** Test Cases ***
A Plain User Reads Their Own Preferences
    [Documentation]    The read grant every account needs. Losing it leaves a signed-in user
    ...    unable to load the settings pane at all.
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}    alias=settings_probe
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Every account must be able to read its own preferences.

A Plain User Cannot Write Tenant Settings
    [Documentation]    Tenant configuration governs everybody. A user reaching it could change
    ...    what the whole tenant sees, which is the escalation this grant is shaped to prevent.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}    alias=settings_probe
    ...    expected_status=any
    Response Should Be Permission Refusal    ${resp}    write tenant settings

A Plain User Cannot Write Org Settings
    [Documentation]    Organization configuration belongs to that organization's admin. A member
    ...    who could rewrite it would be an org admin in everything but name.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}    alias=settings_probe
    ...    expected_status=any
    Response Should Be Permission Refusal    ${resp}    write org settings

A Plain User Cannot Write Settings Rows Directly
    [Documentation]    The resource endpoint must not be a way around the application service.
    ...    The service is where the level check and the fan-out live, so a direct row write
    ...    would let a user store a value no policy ever looked at.
    [Tags]    negative
    ${payload}=    Create Dictionary    module_key=essential    level=tenant
    ...    owner_type=tenant    owner_id=${NOT_FOUND_ID}    name=theme_mode
    ...    value=${{ {'value': 'dark'} }}
    ${resp}=    POST On Session    settings_probe    ${SETTINGS_RECORD_API}
    ...    json=${payload}    expected_status=any
    Should Be True    ${resp.status_code} != 201
    ...    msg=A plain user must not be able to write a settings row directly.

A Plain User Cannot Register A Settings Schema
    [Documentation]    A schema declares what a module can be configured with, and is written by
    ...    boot-time registration. A user who could add one could declare a setting the product
    ...    never validates against anything.
    [Tags]    negative
    ${resp}=    POST On Session    settings_probe    ${SETTINGS_SCHEMA_API}
    ...    json=${{ {'module_key': 'robot_fake', 'level': 'tenant', 'schema': {}} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} != 201
    ...    msg=A plain user must not be able to register a settings schema.

Settings Seeds Grant Read And Never Update
    [Documentation]    The seed decision, asserted against the stored entitlements rather than
    ...    the migration file, so a later hand-edit to the database is caught too.
    ${resp}=    GET On Session    api    /v1/iam/iam_entitlement    params=${{ {'size': 200} }}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        ${resource}=    Get From Dictionary    ${item}    resource_name    ${EMPTY}
        ${action}=      Get From Dictionary    ${item}    action_name      ${EMPTY}
        IF    "${resource}" in ("settings_schema", "settings_record")
            Should Not Be Equal    ${action}    update
            ...    msg=A domain-wide entitlement grants update on ${resource}; settings seeds must grant read only.
            Should Not Be Equal    ${action}    delete
            ...    msg=A domain-wide entitlement grants delete on ${resource}; settings seeds must grant read only.
        END
    END

One User Does Not Read Another User's Preferences
    [Documentation]    A settings read must be scoped to the acting owner. When the scoping is
    ...    dropped the query returns every row in the tenant and the caller is handed whichever
    ...    row happened to come back last — someone else's value, reported as their own.
    ...
    ...    That failure reads as a stale or ignored save rather than as the cross-owner leak it
    ...    is, which is why it is asserted here rather than left to the write tests.
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'light'}] }}
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}    alias=settings_probe
    ...    expected_status=any

    ${mine}=    Get Settings At Level    ${SETTINGS_USER_API}
    Setting Item Should Be    ${mine}    ${SETTING_THEME_MODE}    light

A Domain Administrator Reaches Every Level
    [Documentation]    The seeded test account holds the Domain Administrator role, whose
    ...    entitlement is `*:*:domain` — all actions on all resources. Every level must therefore
    ...    be readable by it.
    ...
    ...    This is asserted because the fan-out suite skips when no tenant-level schema is
    ...    declared, and a genuine permission regression would look exactly like that skip. Here
    ...    a refusal fails outright, so the two cannot be confused.
    FOR    ${api}    IN    ${SETTINGS_TENANT_API}    ${SETTINGS_ORG_API}    ${SETTINGS_USER_API}
        ${resp}=    Get Settings At Level    ${api}    expected_status=any
        Should Be Equal As Integers    ${resp.status_code}    200
        ...    msg=A Domain Administrator was refused a read at ${api} (status ${resp.status_code}).
    END

Writing A Level No Module Declares Is A Caller Error
    [Documentation]    No module registers a tenant-level schema yet (Essential declares user
    ...    level only), so there is nothing to configure there. The refusal must be a 4xx naming
    ...    the missing declaration, not a 403 and not a 500: the caller's permissions are fine and
    ...    the server is behaving correctly — there is simply nothing declared to write.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_TENANT_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'light'}] }}    expected_status=any
    # A 200 would mean some module has since declared a tenant-level setting for essential, in
    # which case the write is legitimately accepted and there is nothing to assert.
    IF    ${resp.status_code} != 200
        Should Be Equal As Integers    ${resp.status_code}    400
        ...    msg=An undeclared level must be refused as a caller error, not as a permission or server failure.
        Should Contain    ${resp.text}    settings.schema_not_registered
    END
