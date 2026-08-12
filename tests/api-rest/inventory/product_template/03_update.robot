*** Settings ***
Documentation     Updating Product Templates, including the status transitions of BR §6.1.2.
...               The success case runs first (it consumes and rotates the saved etag);
...               negatives follow.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
Test Tags         inventory    product_template    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Template
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $PRODUCT_TEMPLATE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Activate Succeeds
    [Documentation]    draft -> active is the normal publication step.
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'status': 'active', 'etag': $PRODUCT_TEMPLATE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Discontinue Does Not Archive
    [Documentation]    BR-PROD-TPL-004 / AC-PROD-018 — the rule this whole field pair exists
    ...    for. `status` is the business lifecycle and `is_archived` is system visibility;
    ...    they are deliberately independent, so a discontinued product stays visible in
    ...    discontinued-product listings instead of vanishing. Collapsing the two would make
    ...    that listing impossible to build.
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'status': 'discontinued', 'etag': $PRODUCT_TEMPLATE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_template.json    200
    Should Be Equal    ${item}[status]    discontinued
    Should Be Equal    ${item}[is_archived]    ${False}
    ...    msg=Discontinuing a template must not archive it (BR-PROD-TPL-004)
    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${item}[etag]

Reactivate Succeeds
    [Documentation]    Discontinuation is reversible: the later suites expect a live
    ...    template, and a one-way transition would make that impossible to restore.
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'status': 'active', 'etag': $PRODUCT_TEMPLATE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Update Capability Flags Succeeds
    [Documentation]    BR §7.6: variants inherit sale_ok, so turning it off here withdraws
    ...    the whole product line from sale in one edit rather than variant by variant.
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'sale_ok': False, 'etag': $PRODUCT_TEMPLATE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'sale_ok': True, 'etag': $PRODUCT_TEMPLATE_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_TEMPLATE_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${etag}

Update With Invalid Status Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'status': 'bla_bla_status', 'etag': $PRODUCT_TEMPLATE_ETAG} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=status must reject a value outside draft/active/discontinued

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Template
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $PRODUCT_TEMPLATE_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $PRODUCT_TEMPLATE_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
