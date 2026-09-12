*** Settings ***
Documentation     Cross-module usage check: Essential asks every module that may hold a
...               uom_id before letting a unit be deleted or re-scaled, and a unit still
...               referenced is refused.
...
...               Runs AFTER 09_delete, which removes the unit that suite owns. This one
...               creates and cleans up its own unit, so the two never share state.
Resource          resources/essential.resource
Resource          ../../resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Referenced Uom
Suite Teardown    Delete Usage Check Fixtures
Test Tags         essential    uom    usage_check


*** Test Cases ***
Delete Is Refused While Inventory References The Unit
    [Documentation]    The stock configuration names this unit, so Inventory answers "in
    ...    use" and the delete is refused with resource_in_use rather than succeeding and
    ...    letting the database blank the reference (the foreign key is ON DELETE SET NULL).
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOM_API}/${REFERENCED_UOM_ID}
    ...    params=${{ {'org_id': $UOM_ORG_ID} }}    expected_status=any
    Response Should Be Resource In Use Error    ${resp}

Refusal Names The Blocking Module
    [Documentation]    BR-DEL-004: the client is told which module to go and clear the
    ...    reference in. The record itself is never disclosed — that is the dependant's data.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${UOM_API}/${REFERENCED_UOM_ID}
    ...    params=${{ {'org_id': $UOM_ORG_ID} }}    expected_status=any
    Should Contain    ${resp.json()}[0][message]    inventory
    ...    msg=The refusal must name the module still referencing the unit

Rescaling Is Refused While The Unit Is Referenced
    [Documentation]    BR-UOM-ESS-020. Changing the factor of a referenced unit would
    ...    reinterpret every quantity already recorded in it.
    ...
    ...    The new factor stays >= 1 so the unit remains a valid `bigger_equal`: a factor
    ...    that also broke BR-UOM-ESS-009 would raise a second violation and the response
    ...    would no longer isolate the rule under test.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${UOM_API}/${REFERENCED_UOM_ID}
    ...    json=${{ {'factor': '24', 'etag': $REFERENCED_UOM_ETAG, 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Uom Immutable While In Use Error    ${resp}    factor

Renaming Is Allowed While The Unit Is Referenced
    [Documentation]    Only factor, type and category are frozen. A name is a label, so
    ...    changing it reinterprets nothing and must not cost a round trip to four modules.
    ${resp}=    PATCH On Session    api    ${UOM_API}/${REFERENCED_UOM_ID}
    ...    json=${{ {'name': {'en-US': 'Robot Renamed Unit'}, 'etag': $REFERENCED_UOM_ETAG, 'org_id': $UOM_ORG_ID} }}
    ${etag}=    Response Should Be Update Success    ${resp}
    Set Global Variable    ${REFERENCED_UOM_ETAG}    ${etag}

Delete Succeeds Once The Reference Is Gone
    [Documentation]    The other half of the rule: the unit becomes deletable the moment
    ...    nothing references it, so the check reports actual usage rather than merely
    ...    refusing everything.
    Delete Stock Product Config Fixture
    ${resp}=    DELETE On Session    api    ${UOM_API}/${REFERENCED_UOM_ID}
    ...    params=${{ {'org_id': $UOM_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1
    Set Global Variable    ${REFERENCED_UOM_ID}    ${EMPTY}


*** Keywords ***
Ensure Referenced Uom
    [Documentation]    A unit of its own plus an Inventory stock configuration naming it,
    ...    which is the reference that must block the delete.
    Ensure Reference Uom
    Ensure Product Template Under Test

    ${name}=    Unique Display Name    Robot Referenced Unit
    ${symbol}=    Unique Symbol    ru
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'bigger_equal', 'factor': '12', 'rounding': '1', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${REFERENCED_UOM_ID}    ${id}
    Set Global Variable    ${REFERENCED_UOM_ETAG}    ${etag}

    ${resp}=    POST On Session    api    ${STOCK_PRODUCT_CONFIG_API}
    ...    json=${{ {'product_template_id': $PRODUCT_TEMPLATE_ID, 'inventory_uom_id': $id, 'org_id': $INV_ORG_ID} }}
    ${config_id}    ${config_etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${STOCK_CONFIG_ID}    ${config_id}

Delete Stock Product Config Fixture
    [Documentation]    Removes the only reference to the unit, so the next delete is
    ...    expected to succeed. Idempotent: the teardown may have run it already.
    ${id}=    Get Variable Value    ${STOCK_CONFIG_ID}    ${EMPTY}
    IF    not $id    RETURN
    DELETE On Session    api    ${STOCK_PRODUCT_CONFIG_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Set Global Variable    ${STOCK_CONFIG_ID}    ${EMPTY}

Delete Usage Check Fixtures
    [Documentation]    Best-effort cleanup in reference order: the configuration first,
    ...    then the unit it points at.
    Delete Stock Product Config Fixture
    ${uom_id}=    Get Variable Value    ${REFERENCED_UOM_ID}    ${EMPTY}
    IF    $uom_id
        DELETE On Session    api    ${UOM_API}/${uom_id}
        ...    params=${{ {'org_id': $UOM_ORG_ID} }}    expected_status=any
        Set Global Variable    ${REFERENCED_UOM_ID}    ${EMPTY}
    END
