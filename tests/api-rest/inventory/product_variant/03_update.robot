*** Settings ***
Documentation     Updating Product Variants. The success case runs first (it consumes and
...               rotates the saved etag); negatives follow. The immutability of
...               product_template_id and the partial-update merge behind BR-PROD-VAR-002
...               are the two rules worth pinning here.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Variant Under Test
Test Tags         inventory    product_variant    update


*** Test Cases ***
Update Succeeds
    ${sku}=    Unique Code    updatedsku
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'sku': $sku, 'etag': $PRODUCT_VARIANT_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Update Leaves The Combination Untouched
    [Documentation]    An update is partial, so the engine overlays the submitted fields onto
    ...    the stored record and validates the result. A SKU-only edit must therefore not be
    ...    read as clearing the combination key — that would either fail uniqueness against
    ...    the empty combination or silently change which product this SKU is.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[combination_key]    ${PRODUCT_VARIANT_COMBINATION}
    ...    msg=A partial update must not disturb the combination key
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${item}[etag]

Discontinue Does Not Archive
    [Documentation]    BR §6.2.2: status lets one variant be withdrawn while its siblings
    ...    stay on sale, and it is independent of is_archived exactly as on the template.
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'status': 'discontinued', 'etag': $PRODUCT_VARIANT_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/product_variant.json    200
    Should Be Equal    ${item}[status]    discontinued
    Should Be Equal    ${item}[is_archived]    ${False}
    ...    msg=Discontinuing a variant must not archive it
    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${item}[etag]

Reactivate Succeeds
    [Documentation]    The later suites expect a live variant, and the archive rules of
    ...    BR-PROD-VAR-006 key off active variants, so the state must be restorable.
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'status': 'active', 'etag': $PRODUCT_VARIANT_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Update Overriding Dimensions Succeeds
    [Documentation]    AC-PROD-014 / BR §7.6: a variant may override the template's fallback
    ...    weight and dimensions. Decimals travel as strings — a JSON number would be parsed
    ...    as a float64 and lose precision before it ever reached the server.
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'weight': '2.25', 'etag': $PRODUCT_VARIANT_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Update With Duplicate Combination Fails
    [Documentation]    BR-PROD-VAR-002 applies to updates as well as creates: moving a
    ...    variant onto a combination a sibling already holds is the same collision, just
    ...    reached from the other side.
    [Tags]    negative
    ${key}=    Unique Code    rival
    ${rival_id}    ${rival_etag}=    Create Product Variant    ${PRODUCT_TEMPLATE_ID}    ${key}
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'combination_key': $key, 'etag': $PRODUCT_VARIANT_ETAG} }}
    ...    expected_status=any
    Response Should Be Duplicate Combination Error    ${resp}
    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${rival_id}    expected_status=any

Update Keeping Its Own Combination Succeeds
    [Documentation]    The uniqueness check must not mistake the record being edited for a
    ...    rival. Re-submitting a variant's own combination key is a no-op, not a collision —
    ...    which is why the lookup fetches two rows and skips the record's own id.
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'combination_key': $PRODUCT_VARIANT_COMBINATION, 'etag': $PRODUCT_VARIANT_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_VARIANT_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_VARIANT_ETAG}    ${etag}

Update Reparenting To Another Template Fails
    [Documentation]    product_template_id is declared no_update: re-parenting a variant
    ...    would silently reinterpret every transaction that already references it.
    [Tags]    negative
    ${template_id}    ${template_etag}=    Create Product Template    Robot Reparent Target
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'product_template_id': $template_id, 'etag': $PRODUCT_VARIANT_ETAG} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=product_template_id is immutable; re-parenting must be refused
    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}    expected_status=any

Update Writing A Computed Field Is Refused
    [Documentation]    template_name is copied from the template on read and has no column of
    ...    its own. Writing it is reported rather than silently dropped, so a client learns the
    ...    value it sent was never going to be kept.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'template_name': 'Renamed Via Variant', 'etag': $PRODUCT_VARIANT_ETAG} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Writing a computed field must be refused, not quietly ignored

Update Writing An Edge Is Refused
    [Documentation]    An edge is hydrated from the peer schema, so echoing a GET response
    ...    straight back into a PATCH must be reported rather than silently dropped — the
    ...    value would never have been stored either way.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'template': {'id': $PRODUCT_TEMPLATE_ID}, 'etag': $PRODUCT_VARIANT_ETAG} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Writing an edge field must be refused, not quietly ignored

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${sku}=    Unique Code    stalesku
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${PRODUCT_VARIANT_ID}
    ...    json=${{ {'sku': $sku, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $PRODUCT_VARIANT_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_VARIANT_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $PRODUCT_VARIANT_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
