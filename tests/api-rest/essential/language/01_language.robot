*** Settings ***
Documentation     Languages are a tenant catalogue answered per organization. A search without an
...               org returns only the public_use languages; with one, is_active is that org's.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Org Under Test
Test Tags         essential    language


*** Test Cases ***
Create A Private Inactive Language
    ${code}=    Unique Locale Code
    ${resp}=    POST On Session    api    ${LANGUAGE_API}
    ...    json=${{ {'name': 'Robot ' + $code, 'locale_code': $code, 'direction': 'ltr', 'public_use': False, 'is_active': False, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${LANGUAGE_ID}    ${id}
    ${resp}=    GET On Session    api    ${LANGUAGE_API}/${id}    params=${{ {'org_id': $UOM_ORG_ID} }}
    Should Be Equal    ${resp.json()}[item][is_active]    ${False}

A Search Without An Organization Returns Public Languages Only
    ${resp}=    GET On Session    api    ${LANGUAGE_API}    params=${{ {'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${publics}=    Evaluate    [item.get('public_use') for item in $resp.json()['items']]
    Should Not Contain    ${publics}    ${False}
    ${ids}=    Evaluate    [item['id'] for item in $resp.json()['items']]
    Should Not Contain    ${ids}    ${LANGUAGE_ID}

The Public Enabled List Lists Public Languages
    ${resp}=    GET On Session    api    ${LANGUAGE_API}/enabled
    Response Status Should Be    ${resp}    200
    ${codes}=    Evaluate    [item['locale_code'] for item in $resp.json()['items']]
    Should Contain    ${codes}    en-US

English Can Be Reactivated For The Organization
    [Documentation]    Activation is written through the language's own update, for the named org.
    ${english}=    Language By Locale Code    en-US
    ${resp}=    PATCH On Session    api    ${LANGUAGE_API}/${english}[id]
    ...    json=${{ {'is_active': True, 'etag': $english['etag'], 'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    200

The System Language Cannot Be Deactivated
    [Documentation]    The org's system_language (the application default en-US when none is
    ...    stored) is what its members fall back to, so it stays active.
    [Tags]    negative
    ${settings}=    GET On Session    api    /v1/settings/org/essential
    ${stored}=    Evaluate    next((item.get('value') for item in $settings.json()['items'] if item['name'] == 'system_language'), None)
    ${code}=    Set Variable If    $stored    ${stored}    en-US
    ${language}=    Language By Locale Code    ${code}
    ${resp}=    PATCH On Session    api    ${LANGUAGE_API}/${language}[id]    expected_status=any
    ...    json=${{ {'is_active': False, 'etag': $language['etag'], 'org_id': $UOM_ORG_ID} }}
    Response Status Should Be    ${resp}    400
    Should Contain    ${resp.text}    err_lang_default_disabled


*** Keywords ***
Language By Locale Code
    [Arguments]    ${code}
    ${resp}=    GET On Session    api    ${LANGUAGE_API}
    ...    params=${{ {'size': 1, 'org_id': $UOM_ORG_ID, 'fields': 'id,etag,locale_code', 'graph': '{"if":["locale_code","=","%s"]}' % $code} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[items][0]
