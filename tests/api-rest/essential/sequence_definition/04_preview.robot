*** Settings ***
Documentation     The preview action (BR §4.2, AC09): it renders the code a configuration would
...               produce for a sample number and reserves nothing - no counter, no issuance.
...               The renders double as the AC05/AC06 worked examples.
Resource          resources/sequence.resource
Suite Setup       Create Authorized API Session
Test Tags         essential    sequence_definition    preview


*** Test Cases ***
Preview Renders Padded Numbers From Zero In Steps Of Two
    [Documentation]    AC05: initial 0, increment 2, padding 3 issue 000, 002, 004.
    FOR    ${sample}    ${expected}    IN    0    000    2    002    4    004
        ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
        ...    json=${{ {'padding': 3, 'initial_value': 0, 'increment': 2, 'sample_number': int($sample)} }}
        Response Status Should Be    ${resp}    200
        Should Be Equal    ${resp.json()}[formatted_value]    ${expected}
        Should Be True    ${resp.json()}[is_illustration]
    END

Preview Grows Past The Padding Instead Of Truncating
    [Documentation]    AC06: padding is a minimum width; 1000000 is never cut to 000000.
    ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
    ...    json=${{ {'padding': 6, 'initial_value': 999999, 'sample_number': 999999} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[formatted_value]    999999
    ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
    ...    json=${{ {'padding': 6, 'initial_value': 999999, 'sample_number': 1000000} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[formatted_value]    1000000

Preview Renders Prefix, Number And Suffix
    ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
    ...    json=${{ {'prefix_template': 'SO', 'suffix_template': '-T', 'sample_number': 42} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[formatted_value]    SO00042-T
    Should Be Equal    ${resp.json()}[period_key]    NONE

Preview Renders The Period Marker In The Definition's Time Zone
    [Documentation]    AC19: 2026-12-31 18:00 UTC is already 2027 in Asia/Ho_Chi_Minh (UTC+7).
    ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
    ...    json=${{ {'prefix_template': 'INV-%(range_year)s-', 'padding': 6, 'reset_period': 'yearly', 'timezone': 'Asia/Ho_Chi_Minh', 'sample_at': '2026-12-31T18:00:00Z'} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[formatted_value]    INV-2027-000001
    Should Be Equal    ${resp.json()}[period_key]    2027

Preview Of A Pseudorandom Code Is CORE Plus Six Crockford Characters
    [Documentation]    SEQ-004: FF1 over Crockford Base32 at the minimum length 6. The preview
    ...    uses a throwaway key, so only the shape is asserted, never the value.
    ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
    ...    json=${{ {'algorithm': 'pseudorandom', 'prefix_template': 'CORE'} }}
    Response Status Should Be    ${resp}    200
    Should Match Regexp    ${resp.json()}[formatted_value]    ^CORE${CROCKFORD_CHAR}{6}$
    Response Should Carry No Secret    ${resp}

Preview Reserves Nothing
    [Documentation]    AC09: repeated previews write no counter and no issuance.
    ${counters}=    GET On Session    api    ${SEQ_COUNTER_API}    params=${{ {'size': 1} }}
    ${issuances}=    GET On Session    api    ${SEQ_ISSUANCE_API}    params=${{ {'size': 1} }}
    FOR    ${index}    IN RANGE    5
        ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}    json=${{ {'prefix_template': 'PV'} }}
        Response Status Should Be    ${resp}    200
    END
    ${counters_after}=    GET On Session    api    ${SEQ_COUNTER_API}    params=${{ {'size': 1} }}
    ${issuances_after}=    GET On Session    api    ${SEQ_ISSUANCE_API}    params=${{ {'size': 1} }}
    Should Be Equal As Integers    ${counters_after.json()}[total]    ${counters.json()}[total]
    Should Be Equal As Integers    ${issuances_after.json()}[total]    ${issuances.json()}[total]

Preview With An Invalid Template Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
    ...    json=${{ {'prefix_template': '%(nope)s'} }}    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_INVALID_TEMPLATE}    prefix_template

Preview With An Unknown Time Zone Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${SEQ_PREVIEW_API}
    ...    json=${{ {'timezone': 'Mars/Olympus_Mons'} }}    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_INVALID_TIMEZONE}    timezone
