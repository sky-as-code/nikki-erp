*** Settings ***
Documentation     Locale-aware search over a LangJson column.
...
...               iam_group.name holds every translation in one jsonb document, so the text
...               a reader sees is never the column itself -- it is one key out of it. These
...               tests prove the server picks that key from the ACTING USER'S OWN stored
...               language setting, without the client naming a language.
...
...               The fixtures are built so that English and Vietnamese sort in opposite
...               orders. That inversion is what makes the assertions load-bearing: a result
...               that merely came back in some stable order would satisfy a single-locale
...               check by luck, but only a server sorting on the right key can produce an
...               order that FLIPS when the reader's language flips.
...
...               The acting user's language is per-user state that outlives this suite, so
...               it is read once and put back in Suite Teardown.
Resource          resources/iam.resource
Resource          ${CURDIR}/../../resources/settings.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Remember Own Language
...               AND    Ensure Locale Ordering Groups
Suite Teardown    Restore Own Language
Test Tags         iam    group    search    langjson


*** Variables ***
${GROUP_SCHEMA}          ${IAM_SCHEMA_DIR}/group.json
${ORIGINAL_LANGUAGE}     ${EMPTY}


*** Test Cases ***
Search Order By Lang Json Name Uses User Locale
    [Documentation]    With the reader set to Vietnamese, ordering by "name" must follow the
    ...    Vietnamese text. The fixtures' Vietnamese names are Xuan/Yen/Zulu, whose ascending
    ...    order is the exact reverse of their English Alpha/Bravo/Charlie.
    Set Own Language    vi-VN
    ${resp}=    Search Groups By Name Order    asc
    ${names}=    Locale Group Names In Result    ${resp}    vi-VN
    Should Be Equal    ${names}    ${{ sorted($names) }}
    ...    msg=Ascending order did not follow the Vietnamese names: ${names}

Search Order By Lang Json Name Follows A Locale Change
    [Documentation]    The test that actually proves locale-awareness. Same request, same
    ...    data, only the acting user's stored language differs -- and because the fixtures
    ...    invert, the resulting order of ids must reverse.
    Set Own Language    vi-VN
    ${vi_resp}=    Search Groups By Name Order    asc
    ${vi_ids}=    Locale Group Ids In Result    ${vi_resp}

    Set Own Language    en-US
    ${en_resp}=    Search Groups By Name Order    asc
    ${en_ids}=    Locale Group Ids In Result    ${en_resp}

    Should Be Equal    ${en_ids}    ${{ list(reversed($vi_ids)) }}
    ...    msg=Changing the reader's language did not change the order: vi=${vi_ids} en=${en_ids}

Search Order By Lang Json Name Descending Uses User Locale
    Set Own Language    vi-VN
    ${resp}=    Search Groups By Name Order    desc
    ${names}=    Locale Group Names In Result    ${resp}    vi-VN
    Should Be Equal    ${names}    ${{ sorted($names, reverse=True) }}
    ...    msg=Descending order did not follow the Vietnamese names: ${names}

Search Filter By Lang Json Name Uses User Locale
    [Documentation]    "contains" matches the Vietnamese text for a Vietnamese reader. The
    ...    English half of the very same document must not match.
    Set Own Language    vi-VN
    ${resp}=    Search Groups Where Name Contains    Xuan
    Search Results Should Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[charlie]

    ${resp}=    GET On Session    api    ${GROUP_API}
    ...    params=${{ {'graph': '{"if":["name","*","Charlie"]}'} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Not Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[charlie]

Search Equals On Lang Json Name Uses User Locale
    [Documentation]    The originally requested shape: WHERE name->>'vi-VN' = '...'. Before
    ...    this feature "=" compared the whole jsonb document and could match nothing at all.
    Set Own Language    vi-VN
    ${exact}=    Set Variable    Xuan ${LOCALE_GROUP_SUFFIX}
    ${graph}=    Evaluate    json.dumps({'if': ['name', '=', $exact]})    modules=json
    ${resp}=    GET On Session    api    ${GROUP_API}    params=${{ {'graph': $graph} }}
    Response Should Be Search Success    ${resp}    ${GROUP_SCHEMA}    size=50    page=0    item_count=1
    Search Results Should Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[charlie]

Search In On Lang Json Name Uses User Locale
    Set Own Language    vi-VN
    ${graph}=    Evaluate
    ...    json.dumps({'if': ['name', 'in', 'Xuan ' + $LOCALE_GROUP_SUFFIX, 'Yen ' + $LOCALE_GROUP_SUFFIX]})
    ...    modules=json
    ${resp}=    GET On Session    api    ${GROUP_API}    params=${{ {'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[charlie]
    Search Results Should Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[bravo]
    Search Results Should Not Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[alpha]

Search Explicit Language Overrides User Locale
    [Documentation]    The stored setting is only the default. A client that names a language
    ...    still wins, which is what lets an export or a report pin its own locale.
    Set Own Language    vi-VN
    ${resp}=    GET On Session    api    ${GROUP_API}
    ...    params=${{ {'language': 'en-US', 'graph': '{"if":["name","*","Charlie"]}'} }}
    Response Status Should Be    ${resp}    200
    Search Results Should Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[charlie]

Search Order By Lang Json Missing Locale Sorts Last
    [Documentation]    Pins the strict, no-fallback rule. The English-only group has no
    ...    Vietnamese text, so for a Vietnamese reader it sorts as NULL -- last on ascending --
    ...    and is not reachable by a Vietnamese filter. It does NOT fall back to its English
    ...    name, which would sort it into the middle of the alphabet while showing blank.
    Set Own Language    vi-VN
    ${resp}=    Search Groups By Name Order    asc
    ${ids}=    Locale Group Ids In Result    ${resp}
    Should Be Equal    ${ids}[-1]    ${LOCALE_GROUP_IDS}[en_only]
    ...    msg=A row with no text in the reader's language must sort last, got ${ids}

    ${resp}=    Search Groups Where Name Contains    Delta
    Search Results Should Not Contain Id    ${resp}    ${LOCALE_GROUP_IDS}[en_only]

Search With Invalid Language Fails
    [Documentation]    The language reaches SQL as a literal, so a malformed one is rejected
    ...    at the edge rather than passed down to match nothing.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${GROUP_API}
    ...    params=${{ {'language': 'not a locale'} }}    expected_status=any
    Should Be True    ${{ $resp.status_code >= 400 }}
    ...    msg=A malformed language should be a client error, got ${resp.status_code}


*** Keywords ***
Search Groups By Name Order
    [Arguments]    ${direction}
    ${graph}=    Evaluate    json.dumps({'order': [['name', $direction]]})    modules=json
    ${resp}=    GET On Session    api    ${GROUP_API}    params=${{ {'graph': $graph, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp}

Search Groups Where Name Contains
    [Arguments]    ${fragment}
    ${graph}=    Evaluate    json.dumps({'if': ['name', '*', $fragment]})    modules=json
    ${resp}=    GET On Session    api    ${GROUP_API}    params=${{ {'graph': $graph} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp}

Locale Group Ids In Result
    [Documentation]    Result-order ids, narrowed to the groups this suite created so that
    ...    unrelated rows cannot decide whether an ordering assertion passes.
    [Arguments]    ${resp}
    ${ids}=    Evaluate
    ...    [i['id'] for i in $resp.json()['items'] if isinstance(i.get('name'), dict) and $LOCALE_GROUP_SUFFIX in str(i['name'])]
    RETURN    ${ids}

Remember Own Language
    [Documentation]    A language is per-user state that outlives this suite, so it is put
    ...    back in teardown rather than left wherever the last test set it.
    ${resp}=    Get Settings At Level    ${SETTINGS_USER_API}
    Response Status Should Be    ${resp}    200
    ${item}=    Find Setting Item    ${resp}    ${SETTING_LANGUAGE}
    Set Global Variable    ${ORIGINAL_LANGUAGE}    ${item}[value]

Set Own Language
    [Arguments]    ${value}
    ${resp}=    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': $SETTING_LANGUAGE, 'value': $value}] }}
    Response Status Should Be    ${resp}    200

Restore Own Language
    ${value}=    Get Variable Value    ${ORIGINAL_LANGUAGE}    ${EMPTY}
    IF    not $value    RETURN
    Set Settings At Level    ${SETTINGS_USER_API}
    ...    ${{ [{'name': $SETTING_LANGUAGE, 'value': $value}] }}    expected_status=any
