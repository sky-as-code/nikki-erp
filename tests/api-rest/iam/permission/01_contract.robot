*** Settings ***
Documentation     Contract of POST /v1/iam/me/test-permissions.
...
...               The endpoint takes an expression from an untrusted request body, so the
...               rules it must never break are: reject anonymous callers; answer 4xx and
...               never 5xx on malformed input; refuse wildcard questions, which would
...               turn it into a grant-enumeration tool; and never accept a subject other
...               than the caller.
Resource          ../../resources/iam.resource
Suite Setup       Create Authorized API Session


*** Test Cases ***
Probe Refuses An Unauthenticated Caller
    [Documentation]    The probe reports the caller's own access, so there must be a
    ...    caller. Without a token there is nothing to answer about.
    Create Anonymous API Session    alias=anon_probe
    ${resp}=    POST On Session    anon_probe    ${TEST_PERM_API}
    ...    json=${{ {'expression': 'read:iam_user:domain'} }}    expected_status=any
    Should Be True    ${resp.status_code} in (401, 403)
    ...    msg=Anonymous probe returned ${resp.status_code}, expected 401/403

Probe Answers A Well Formed Question
    [Documentation]    The shape of a successful answer: a boolean and a list, with the
    ...    list always present so a client needs no null check.
    ${body}=    Probe Permission    api    read:iam_user:domain
    Dictionary Should Contain Key    ${body}    is_granted
    Dictionary Should Contain Key    ${body}    matches
    ${type}=    Evaluate    type($body['matches']).__name__
    Should Be Equal    ${type}    list

Probe Reports Provenance When Granted
    [Documentation]    A granted answer must name the grant path. The administrator
    ...    session used by the whole test run holds broad user permissions, so this
    ...    question is granted and every match must carry a usable source.
    ${body}=    Probe Permission    api    read:iam_user:domain
    Should Be True    ${body}[is_granted]
    Should Not Be Empty    ${body}[matches]
    FOR    ${match}    IN    @{body}[matches]
        Dictionary Should Contain Key    ${match}    source_kind
        Dictionary Should Contain Key    ${match}    source_id
        Dictionary Should Contain Key    ${match}    source_name
        Dictionary Should Contain Key    ${match}    ent_expression
        Should Contain Any    ${match}[source_kind]    direct    group    owner
    END

Probe Rejects Malformed Expressions Without Failing
    [Documentation]    Every one of these is a client mistake, not a server fault. A 500
    ...    here would mean untrusted input reached something that panicked.
    @{malformed}=    Create List
    ...    ${EMPTY}
    ...    read
    ...    read:iam_user
    ...    read:iam_user:domain:extra
    ...    read:iam_user:galaxy
    ...    read:iam_user:domain/ORG1
    ...    ::
    ...    :iam_user:domain
    ...    read::domain
    FOR    ${expression}    IN    @{malformed}
        ${resp}=    Probe Permission Raw    api    ${expression}
        Should Be True    ${resp.status_code} >= 400 and ${resp.status_code} < 500
        ...    msg=Expression "${expression}" returned ${resp.status_code}, expected a 4xx
    END

Probe Rejects Injection Shaped And Oversized Input
    [Documentation]    The expression is used to build an IN-list, so payloads shaped like
    ...    SQL or a very long string must be refused by the grammar, never executed and
    ...    never allowed to exhaust anything.
    ${long}=    Evaluate    'a' * 5000
    @{hostile}=    Create List
    ...    read:iam_user:domain' OR '1'='1
    ...    read:iam_user:domain; DROP TABLE iam_user_permissions
    ...    read:iam_user:domain\n\nread:iam_role:domain
    ...    ${long}
    ...    read:${long}:domain
    FOR    ${expression}    IN    @{hostile}
        ${resp}=    Probe Permission Raw    api    ${expression}
        Should Be True    ${resp.status_code} >= 400 and ${resp.status_code} < 500
        ...    msg=Hostile input returned ${resp.status_code}, expected a 4xx
    END
    # The table must still be there afterwards.
    ${body}=    Probe Permission    api    read:iam_user:domain
    Dictionary Should Contain Key    ${body}    is_granted

Probe Refuses Wildcard Questions
    [Documentation]    Wildcards are legitimate in a stored grant but not in a question:
    ...    a real requirement is always concrete, and answering wildcard questions would
    ...    let any signed-in user enumerate the shape of their own grants and, worse,
    ...    probe for the existence of resources by pattern.
    @{wildcards}=    Create List
    ...    *:iam_user:domain
    ...    read:*:domain
    ...    *:*:domain
    ...    *:*:*
    FOR    ${expression}    IN    @{wildcards}
        ${resp}=    Probe Permission Raw    api    ${expression}
        Should Be True    ${resp.status_code} >= 400 and ${resp.status_code} < 500
        ...    msg=Wildcard "${expression}" returned ${resp.status_code}, expected a 4xx
    END

Probe Ignores Any Attempt To Name Another Subject
    [Documentation]    There is no user_id parameter by design. Sending one must not
    ...    change the subject: the answer must stay the caller's own, whatever extra
    ...    fields the body carries.
    ${honest}=    Probe Permission    api    read:iam_user:domain
    ${resp}=    POST On Session    api    ${TEST_PERM_API}
    ...    json=${{ {'expression': 'read:iam_user:domain', 'user_id': '01JWNXT3EY7FG47VDJTEPTDC98'} }}
    ...    expected_status=any
    IF    ${resp.status_code} == 200
        Should Be Equal    ${resp.json()}[is_granted]    ${honest}[is_granted]
        ...    msg=A user_id in the body changed the answer; the probe is not self-only
    ELSE
        Should Be True    ${resp.status_code} >= 400 and ${resp.status_code} < 500
    END
