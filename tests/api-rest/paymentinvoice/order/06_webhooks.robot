*** Settings ***
Documentation     The gateway callbacks.
...
...               These are the one part of the module that is not an engine action: they are
...               called by MoMo, by NextPay and by the bank, none of which can present a
...               user's authorization. Each authenticates its own caller instead — by
...               signature, by decryption, or by a bearer this deployment issued.
...
...               What is pinned here is the refusal path, which is what a stock deployment
...               can exercise: an unsigned or unauthenticated caller must never be able to
...               settle an order, and must never be told whether the order exists. Enumerating
...               order codes through the callback is exactly what that would allow.
Resource          resources/paymentinvoice.resource
Suite Setup       Create Authorized API Session
Test Tags         paymentinvoice    order    webhooks


*** Variables ***
${WEBHOOK_MOMO}            /v1/paymentinvoice/webhooks/momo
${WEBHOOK_MPOS}            /v1/paymentinvoice/webhooks/mpos
${WEBHOOK_VIETQR_TOKEN}    /v1/paymentinvoice/webhooks/vietqr/token_generate
${WEBHOOK_VIETQR_SYNC}     /v1/paymentinvoice/webhooks/vietqr/transaction_sync


*** Test Cases ***
The Momo Callback Answers Without Authorization
    [Documentation]    MoMo cannot present a user session, so the endpoint must be reachable
    ...    unauthenticated. It answers 204 whatever the outcome because MoMo retries anything
    ...    else, and a callback it cannot fix by resending is not worth a retry.
    ${resp}=    POST On Session    api    ${WEBHOOK_MOMO}
    ...    json=${{ {'orderId': 'NOSUCHORDER00000000', 'resultCode': 0, 'transId': 1} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    204

An Unsigned Momo Callback Settles Nothing
    [Documentation]    The body carries no valid signature, so it did not come from MoMo. The
    ...    reply is deliberately identical to a valid callback's: telling an unsigned caller
    ...    that their signature was wrong helps them work out a correct one.
    ${resp}=    POST On Session    api    ${WEBHOOK_MOMO}
    ...    json=${{ {'orderId': 'FORGED0000000000000', 'resultCode': 0, 'transId': 9, 'signature': 'forged'} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    204

The Momo Callback Survives A Malformed Body
    [Documentation]    A body that will not parse must not take the endpoint down; MoMo would
    ...    retry it forever.
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${WEBHOOK_MOMO}
    ...    data={ "orderId":    headers=${headers}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    204

The Mpos Callback Answers Without Authorization
    ${resp}=    POST On Session    api    ${WEBHOOK_MPOS}
    ...    json=${{ {'merchantId': 'robot', 'reqData': 'not-decryptable'} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    204

An Undecryptable Mpos Callback Settles Nothing
    [Documentation]    The envelope is encrypted under the merchant secret, so being able to
    ...    decrypt it IS the authentication. A body that will not decrypt did not come from
    ...    mPOS and must be refused rather than acted on.
    ${resp}=    POST On Session    api    ${WEBHOOK_MPOS}
    ...    json=${{ {'merchantId': 'robot', 'reqData': 'AAAAAAAAAAAAAAAAAAAAAA=='} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    204

The VietQr Token Endpoint Refuses A Caller With No Credentials
    [Documentation]    VietQR's integration has the partner host this endpoint: the bank logs
    ...    in here and gets a bearer of our issuing. A caller presenting nothing must be
    ...    refused — this is the endpoint that guards marking orders paid.
    ${resp}=    POST On Session    api    ${WEBHOOK_VIETQR_TOKEN}
    ...    json=${{ {} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    401

The VietQr Token Endpoint Refuses Wrong Credentials
    [Tags]    negative
    ${headers}=    Create Dictionary    Authorization=Basic bm9ib2R5Om5vcGU=
    ${resp}=    POST On Session    api    ${WEBHOOK_VIETQR_TOKEN}
    ...    json=${{ {} }}    headers=${headers}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    401

The VietQr Sync Endpoint Refuses A Caller With No Bearer
    [Documentation]    Its reply is non-standard and must stay byte-exact: the bank reads the
    ...    body rather than the status code and keys off errorReason, where 001 is
    ...    "no such transaction".
    ${resp}=    POST On Session    api    ${WEBHOOK_VIETQR_SYNC}
    ...    json=${{ {'orderId': 'NOSUCHORDER00000000', 'referencenumber': 'x'} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    401
    ${body}=    Set Variable    ${resp.json()}
    Should Be Equal    ${body}[error]    ${True}
    Should Be Equal    ${body}[errorReason]    001

The VietQr Sync Endpoint Refuses A Forged Bearer
    [Documentation]    A token this deployment did not sign must not verify. The alternative
    ...    is that anyone able to reach the URL can mark any order paid.
    [Tags]    negative
    ${headers}=    Create Dictionary
    ...    Authorization=Bearer eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJhdHRhY2tlciJ9.
    ${resp}=    POST On Session    api    ${WEBHOOK_VIETQR_SYNC}
    ...    json=${{ {'orderId': 'NOSUCHORDER00000000', 'referencenumber': 'x'} }}
    ...    headers=${headers}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    401
    ...    msg=An unsigned or forged bearer must never be accepted
