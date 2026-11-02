*** Settings ***
Documentation     GET /v1/essential/timezones searches the IANA catalogue by "contains" on the id and
...               on its label in the requested language, "_" read as a space.
Resource          resources/essential.resource
Suite Setup       Create Authorized API Session
Test Tags         essential    timezone


*** Test Cases ***
A Keyword Matches The Zone Id
    ${resp}=    GET On Session    api    ${TIMEZONES_API}    params=${{ {'keyword': 'ho chi'} }}
    Response Status Should Be    ${resp}    200
    ${codes}=    Evaluate    [item['code'] for item in $resp.json()['items']]
    Should Contain    ${codes}    Asia/Ho_Chi_Minh

A Keyword Matches The Translated Label
    ${resp}=    GET On Session    api    ${TIMEZONES_API}    params=${{ {'keyword': 'Châu Á/Hồ', 'language_code': 'vi-VN'} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[items][0][label]    Châu Á/Hồ_Chí_Minh

The Catalogue Pages
    ${resp}=    GET On Session    api    ${TIMEZONES_API}    params=${{ {'page': 1, 'size': 10} }}
    Response Status Should Be    ${resp}    200
    Length Should Be    ${resp.json()}[items]    10
    Should Be True    ${resp.json()}[total] > 400

A Caller Without A Token Is Refused
    [Tags]    negative
    Create Anonymous API Session    anonymous
    ${resp}=    GET On Session    anonymous    ${TIMEZONES_API}    expected_status=any
    Should Be True    400 <= ${resp.status_code} < 500
