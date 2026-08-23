*** Settings ***
Documentation     Writing settings. The payload carries only the items the caller changed, which
...               is the mechanism that makes last-write-wins safe (D17): an untouched setting is
...               never in the payload, so it can never be clobbered by someone who did not edit
...               it.
Resource          resources/settings.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Remember Original Theme
Test Tags         settings    levels    write


*** Keywords ***
Remember Original Theme
    ${original}=    Read Own Theme Mode
    Set Global Variable    ${SETTINGS_ORIGINAL_THEME}    ${original}


*** Test Cases ***
Write Own Preference Succeeds
    ${resp}=    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}
    Response Status Should Be    ${resp}    200
    Should Be True    ${resp.json()}[updated] >= 1

    ${read}=    Get Settings At Level    ${SETTINGS_USER_API}
    Setting Item Should Be    ${read}    ${SETTING_THEME_MODE}    dark

A Second Save Overwrites The First
    [Documentation]    D17. There is no version check, so the later write wins outright. This is
    ...    the documented behaviour, not an oversight: a genuine collision needs two actors
    ...    editing the same setting at the same level within one page session, and the loser
    ...    sees the winner's value on the next load.
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'light'}] }}

    ${read}=    Get Settings At Level    ${SETTINGS_USER_API}
    Setting Item Should Be    ${read}    ${SETTING_THEME_MODE}    light

An Absent Item Is Left Untouched
    [Documentation]    The partial save is what protects concurrent editors. A write naming only
    ...    theme_mode must not reset language, or every save would clobber the whole pane.
    ${before}=    Get Settings At Level    ${SETTINGS_USER_API}
    ${language}=    Find Setting Item    ${before}    ${SETTING_LANGUAGE}

    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'auto'}] }}

    ${after}=    Get Settings At Level    ${SETTINGS_USER_API}
    ${language_after}=    Find Setting Item    ${after}    ${SETTING_LANGUAGE}
    Should Be Equal    ${language_after}[value]    ${language}[value]
    ...    msg=A partial save changed a setting it did not name.

Writing An Unknown Setting Name Is Refused
    [Documentation]    AC-10. A name the schema does not declare has no type to validate against
    ...    and no meaning to any reader. Accepting it would store a row nothing ever reads back.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'no_such_setting', 'value': 'x'}] }}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=An undeclared setting name must be refused, not stored.

Writing A Value Outside The Enum Is Refused
    [Documentation]    theme_mode declares exactly light, dark and auto. A fourth value would
    ...    render as no theme at all, and the frontend has no branch for it.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'ultraviolet'}] }}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=A value outside the declared enum must be refused.
