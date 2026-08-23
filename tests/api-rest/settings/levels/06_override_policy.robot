*** Settings ***
Documentation     Who may change whether a setting can be overridden, and what that change does.
...
...               `allow_override` moved from schema metadata onto the record, so a Tenant Admin
...               decides it for their own tenant rather than a module fixing it for everyone at
...               build time. That makes it writable, and therefore something that has to be
...               guarded: the flag is the mechanism that locks the levels below, so an owner able
...               to set it on themselves could unlock what the tenant locked.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    levels    override


*** Test Cases ***
An Org May Not Change Its Own Override Policy
    [Documentation]    The central guard. An org sending allow_override is refused outright rather
    ...    than having the field quietly dropped — a silent drop would let it believe it had
    ...    unlocked a setting the tenant administrator had locked.
    ${resp}=    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'system_timezone', 'value': 'Asia/Ho_Chi_Minh', 'allow_override': True}] }}
    Response Status Should Be    ${resp}    400

A User May Not Change Their Own Override Policy
    [Documentation]    Same rule one level down. A user is the owner the flag exists to constrain,
    ...    so they are the last actor who may relax it.
    ${resp}=    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'timezone', 'value': 'Europe/London', 'allow_override': True}] }}
    Response Status Should Be    ${resp}    400

Omitting The Flag Leaves The Stored Policy Alone
    [Documentation]    The common write. Every ordinary save omits allow_override, and must not be
    ...    read as a request to reset it — otherwise saving a value would silently unlock a setting
    ...    the tenant administrator had enforced.
    ${before}=    Get Settings At Level    ${SETTINGS_USER_API}
    ${item_before}=    Find Setting Item    ${before}    ${SETTING_THEME_MODE}

    ${resp}=    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}
    Response Status Should Be    ${resp}    200

    ${after}=    Get Settings At Level    ${SETTINGS_USER_API}
    ${item_after}=    Find Setting Item    ${after}    ${SETTING_THEME_MODE}
    Should Be Equal    ${item_before}[allow_override]    ${item_after}[allow_override]

Every Item Still Reports An Override Policy
    [Documentation]    The column is nullable, and a row nobody has ruled on falls back to the
    ...    module's declared metadata. A null reaching the client as a missing key would break
    ...    every pane that renders the lock state.
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}
    ${items}=    Set Variable    ${resp.json()}[items]
    FOR    ${item}    IN    @{items}
        Dictionary Should Contain Key    ${item}    allow_override
        Should Be True    isinstance($item['allow_override'], bool)
    END
