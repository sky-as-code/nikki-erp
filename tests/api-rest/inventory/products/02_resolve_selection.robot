*** Settings ***
Documentation     Selection resolution (BR §14.4): turning a template plus chosen attribute
...               values into the concrete variant a transaction line must reference. The
...               payload validation is the point — without it a malformed `selections` would
...               decode to an empty list, resolve to the empty combination (a real and
...               different variant) and answer 200 with the wrong product.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
Test Tags         inventory    products    resolve


*** Test Cases ***
Resolve With No Selections Succeeds
    [Documentation]    BR §4.5: a template with no variant-generating attributes resolves on
    ...    an empty selection, so absence is legitimate rather than an error.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'template_id': $PRODUCT_TEMPLATE_ID} }}
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}    combination_key
    Dictionary Should Contain Key    ${resp.json()}    materialized

Resolve Without Template Id Fails
    [Documentation]    resolve_selection identifies its template through the body rather than
    ...    the path, so a missing template_id is a client error the caller can fix, not a 404.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Template Id Required Error    ${resp}

Resolve With Empty Template Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'template_id': ''} }}    expected_status=any
    Response Should Be Template Id Required Error    ${resp}

Resolve With Malformed Selections Fails
    [Documentation]    The rule this endpoint's ParamSchema exists for: `selections` must be
    ...    a list. A string silently decoding to an empty list would resolve to the empty
    ...    combination and return the wrong variant with a 200.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'template_id': $PRODUCT_TEMPLATE_ID, 'selections': 'not-a-list'} }}
    ...    expected_status=any
    Response Should Be Selections Malformed Error    ${resp}

Resolve With Not Found Template Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'template_id': $NOT_FOUND_ID} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Resolving against a template that does not exist must not succeed
