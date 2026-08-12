*** Settings ***
Documentation     Shared HTTP session management and constants for API REST tests.
...               App-specific values (host, certs, credentials) come exclusively from
...               the environment variable file selected with --variablefile.
Library           Collections
Library           RequestsLibrary
Library           ${CURDIR}/SchemaValidator.py


*** Variables ***
# Visibly fake but format-valid ULID (first char must be 0-7) shared by every
# not-found test case. See "docs/06. API testing.md" rule 8.
${NOT_FOUND_ID}                 0A000000000000000000000001
${NOT_FOUND_ID_PREFIX}          0A00000000000000000000000
${COMMON_SCHEMA_DIR}            ${CURDIR}/schemas/common


*** Keywords ***
Not Found Id Variant
    [Documentation]    Returns a distinct fake-but-valid ULID for the given digit 1-9.
    [Arguments]    ${digit}
    RETURN    ${NOT_FOUND_ID_PREFIX}${digit}

Not Found Id List
    [Documentation]    Returns ${count} distinct fake-but-valid ULIDs (count 1-9).
    [Arguments]    ${count}=5
    ${ids}=    Create List
    FOR    ${digit}    IN RANGE    1    ${{ int($count) + 1 }}
        Append To List    ${ids}    ${NOT_FOUND_ID_PREFIX}${digit}
    END
    RETURN    ${ids}

Create Anonymous API Session
    [Documentation]    Session with mTLS client certificate but no Authorization header.
    [Arguments]    ${alias}=api
    ${certs}=    Evaluate    ($CLIENT_CERT, $CLIENT_KEY)
    Create Client Cert Session    ${alias}    ${API_HOST}
    ...    client_certs=${certs}    verify=${SSL_VERIFY}    disable_warnings=${1}
