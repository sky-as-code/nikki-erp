*** Settings ***
Documentation     Organization-level settings. Essential declares three: the locale an
...               organization's shared documents are written in, the timezone its business day
...               runs on, and the currency its figures are quoted in.
...
...               These are organization-level rather than user-level on purpose. A user sitting
...               in a different timezone from their organization needs both: the organization's
...               zone is what a shared report is stamped with, their own is how that report is
...               shown to them. Collapsing the two makes one of them wrong for anyone remote.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    levels    org


*** Test Cases ***
Read Org Settings Returns Every Declared Setting
    ${resp}=    Get Settings At Level    ${SETTINGS_ORG_API}
    Response Status Should Be    ${resp}    200
    Find Setting Item    ${resp}    ${SETTING_SYSTEM_LOCALE}
    Find Setting Item    ${resp}    ${SETTING_SYSTEM_TIMEZONE}
    Find Setting Item    ${resp}    ${SETTING_DEFAULT_CURRENCY}

Every Org Setting Carries A Translated Description
    [Documentation]    The sentence explaining a setting is shown beside its control, so it must
    ...    be a translation reference rather than literal text. A missing description leaves the
    ...    pane showing a bare field name to whoever has to decide what to set it to.
    ${resp}=    Get Settings At Level    ${SETTINGS_ORG_API}
    Response Status Should Be    ${resp}    200
    FOR    ${item}    IN    @{resp.json()}[items]
        ${field}=    Get From Dictionary    ${item}    field
        ${description}=    Get From Dictionary    ${field}    description    ${NONE}
        Should Not Be Equal    ${description}    ${NONE}
        ...    msg=Setting '${item}[name]' carries no description.
        Dictionary Should Contain Key    ${description}    $ref
        ...    msg=Setting '${item}[name]' must describe itself by $ref, not literal text.
        Should Be Equal    ${description}[$ref]    settings_desc.${item}[name]
        ...    msg=The description key convention is settings_desc.<setting name>.
    END

Write An Org Setting Succeeds
    ${resp}=    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'system_timezone', 'value': 'Asia/Ho_Chi_Minh'}] }}
    Response Status Should Be    ${resp}    200

    ${read}=    Get Settings At Level    ${SETTINGS_ORG_API}
    Setting Item Should Be    ${read}    ${SETTING_SYSTEM_TIMEZONE}    Asia/Ho_Chi_Minh

An Org Setting Write Leaves The Others Untouched
    [Documentation]    The partial save is what protects concurrent editors: two admins changing
    ...    different settings of one organization must not overwrite each other.
    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'default_currency', 'value': 'VND'}] }}
    ${before}=    Get Settings At Level    ${SETTINGS_ORG_API}
    ${currency}=    Find Setting Item    ${before}    ${SETTING_DEFAULT_CURRENCY}

    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'system_timezone', 'value': 'Europe/Paris'}] }}

    ${after}=    Get Settings At Level    ${SETTINGS_ORG_API}
    ${currency_after}=    Find Setting Item    ${after}    ${SETTING_DEFAULT_CURRENCY}
    Should Be Equal    ${currency_after}[value]    ${currency}[value]
    ...    msg=A partial save changed an org setting it did not name.

An Org Locale Outside The Supported Set Is Refused
    [Documentation]    system_locale is an enum of exactly the locales the application ships
    ...    translations for. There is no fallback language, so an unsupported locale would render
    ...    every key of every shared document as its raw namespace:key.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'system_locale', 'value': 'xx-XX'}] }}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=An unsupported locale must be refused.

An Org Setting Cannot Be Written At The User Level
    [Documentation]    Each level declares its own names. Reaching an org setting through the user
    ...    route would let a person redefine what their whole organization means by a currency.
    [Tags]    negative
    ${resp}=    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'default_currency', 'value': 'USD'}] }}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
    ...    msg=An org-level setting must not be writable through the user-level route.

A User Timezone Is Separate From The Organization Timezone
    [Documentation]    Both exist deliberately. Writing one must not move the other, or a remote
    ...    user changing their own clock would silently restamp every shared report.
    Set Settings At Level    ${SETTINGS_ORG_API}
    ...    ${{ [{'name': 'system_timezone', 'value': 'Asia/Ho_Chi_Minh'}] }}
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'timezone', 'value': 'Europe/London'}] }}

    ${org}=    Get Settings At Level    ${SETTINGS_ORG_API}
    Setting Item Should Be    ${org}    ${SETTING_SYSTEM_TIMEZONE}    Asia/Ho_Chi_Minh
    ${user}=    Get Settings At Level    ${SETTINGS_USER_API}
    Setting Item Should Be    ${user}    ${SETTING_TIMEZONE}    Europe/London
