*** Settings ***
Documentation     Updating Units of Measure. A partial update must be validated against
...               the record it produces, not the submitted fragment alone — the
...               "Change Type Without Factor" test is what pins that.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Uom Under Test
Test Tags         essential    uom    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Gram
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $UOM_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOM_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOM_ETAG}    ${etag}

Update Factor Within Type Succeeds
    [Documentation]    The unit under test is `smaller`, so any factor in (0, 1) is a
    ...    legal change while no transaction references it (BR-UOM-ESS-020).
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {'factor': '0.002', 'etag': $UOM_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOM_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOM_ETAG}    ${etag}
    # Restore, so later suites see the documented 0.001 gram.
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {'factor': '0.001', 'etag': $UOM_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${UOM_ETAG}
    IF    $etag is not None    Set Global Variable    ${UOM_ETAG}    ${etag}

Change Type Without Factor Fails
    [Documentation]    BR-UOM-ESS-009: submitting only uom_type leaves the stored factor
    ...    of 0.001 in place, which contradicts `bigger_equal`. The rule has to be checked
    ...    against the merged record; validating the fragment alone would let this pass
    ...    and leave a UoM whose type and factor disagree.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {'uom_type': 'bigger_equal', 'etag': $UOM_ETAG} }}    expected_status=any
    Response Should Be Uom Bigger Equal Factor Error    ${resp}

Promote To Second Reference Fails
    [Documentation]    BR-UOM-ESS-005: the category already has a reference, so promoting
    ...    a second unit is refused on update exactly as it is on create.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {'uom_type': 'reference', 'factor': '1', 'etag': $UOM_ETAG} }}
    ...    expected_status=any
    Response Should Be Uom Duplicate Reference Error    ${resp}

Update Reference To Factor Other Than One Fails
    [Documentation]    BR-UOM-ESS-006 holds for the existing reference too, not only at
    ...    the moment it is created.
    [Tags]    negative
    Ensure Reference Uom
    ${resp}=    GET On Session    api    ${UOM_API}/${REFERENCE_UOM_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uom.json    200
    ${resp}=    PATCH On Session    api    ${UOM_API}/${REFERENCE_UOM_ID}
    ...    json=${{ {'factor': '5', 'etag': $item['etag']} }}    expected_status=any
    Response Should Be Uom Reference Factor Error    ${resp}

Update With Rounding Above One Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {'rounding': '1.5', 'etag': $UOM_ETAG} }}    expected_status=any
    Response Should Be Uom Rounding Range Error    ${resp}

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/${UOM_ID}
    ...    json=${{ {'factor': '0.005', 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $UOM_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $UOM_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
