*** Settings ***
Documentation     Reading the value rows directly. The rows a user owns are created by the
...               per-level write path, so this suite reads what the levels suite wrote rather
...               than creating rows behind the application service's back.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    settings_record    search


*** Test Cases ***
Search Settings Records Succeeds
    ${resp}=    GET On Session    api    ${SETTINGS_RECORD_API}    params=${{ {'size': 20} }}
    Response Status Should Be    ${resp}    200

A Written Preference Becomes A Stored Row
    [Documentation]    The value the per-level route writes is the row the resource endpoint
    ...    reads back. If these two diverged, the settings pane and the stored data would
    ...    disagree with nothing reporting an error.
    ${original}=    Read Own Theme Mode
    Set Global Variable    ${SETTINGS_ORIGINAL_THEME}    ${original}

    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': 'theme_mode', 'value': 'dark'}] }}

    ${resp}=    GET On Session    api    ${SETTINGS_RECORD_API}    params=${{ {'size': 200} }}
    Response Status Should Be    ${resp}    200
    ${matches}=    Evaluate
    ...    [i for i in $resp.json()['items'] if i['name'] == 'theme_mode' and i['level'] == 'user']
    Should Not Be Empty    ${matches}
    ...    msg=Writing a user preference stored no settings_record row.
