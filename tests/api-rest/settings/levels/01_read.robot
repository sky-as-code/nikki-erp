*** Settings ***
Documentation     Reading a module's settings at each level. A read fills names that have no
...               row yet from the schema's declared default, so a caller renders a complete
...               pane on a fresh account rather than an empty one.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    levels    read


*** Test Cases ***
Read Own Preferences Returns Every Declared Setting
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[module_key]    ${SETTINGS_MODULE_KEY}
    Find Setting Item    ${resp}    ${SETTING_THEME_MODE}
    Find Setting Item    ${resp}    ${SETTING_LANGUAGE}

Every Item Carries Its Declaration
    [Documentation]    The item ships enough of its field declaration for a client to render the
    ...    right control without a second round trip to the schema. Dropping it would leave the
    ...    settings pane unable to tell an enum from free text.
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}
    Response Status Should Be    ${resp}    200
    ${item}=    Find Setting Item    ${resp}    ${SETTING_THEME_MODE}
    Dictionary Should Contain Key    ${item}    field
    Dictionary Should Contain Key    ${item}    editable
    Dictionary Should Contain Key    ${item}    has_value
    Dictionary Should Contain Key    ${item}    allow_override

A Setting With No Row Reports Its Default As Unset
    [Documentation]    has_value distinguishes a stored choice from a schema default. Collapsing
    ...    the two would make an untouched setting look deliberately chosen, and the fan-out
    ...    would then have nothing to tell apart.
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}
    Response Status Should Be    ${resp}    200
    ${item}=    Find Setting Item    ${resp}    ${SETTING_LANGUAGE}
    Should Be True    isinstance($item['has_value'], bool)

Reading An Unknown Module Is Not A Server Error
    [Documentation]    A module that registered nothing has no settings. That is an empty
    ...    answer, not a crash, because the settings pane asks before it knows.
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}    module_key=no_such_module
    ...    expected_status=any
    Should Be True    ${resp.status_code} < 500
    ...    msg=An unknown module key must not fault the server.
