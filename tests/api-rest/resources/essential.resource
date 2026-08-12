*** Settings ***
Documentation     Essential-module helpers shared by the UoM suites: the category and
...               unit under test, the reference UoM every category needs before its
...               members can be created, and bulk sample data for Search/Exists.
...               All keywords expect an authorized session with alias "api".
Library           Collections
Resource          ${CURDIR}/../nikki_api/auth.resource
Resource          ${CURDIR}/../nikki_api/assertions.resource
Resource          ${CURDIR}/../nikki_api/data.resource


*** Variables ***
${UOM_API}                  /v1/essential/essential_uom
${UOMCAT_API}               /v1/essential/essential_uomcat
${UOM_CONVERT_API}          /v1/essential/uoms/convert
${UOM_TO_REFERENCE_API}     /v1/essential/uoms/to_reference
${ESSENTIAL_SCHEMA_DIR}     ${CURDIR}/schemas/essential


*** Keywords ***
Ensure Org Under Test
    [Documentation]    Both UoM models declare org_id required-for-create, so every
    ...    fixture needs an organization. ORG_ID from the environment wins when set;
    ...    otherwise the first organization the API returns is used, which keeps the
    ...    suite app-agnostic instead of hardcoding a seeded ULID.
    ${configured}=    Get Variable Value    ${ORG_ID}    ${EMPTY}
    IF    $configured
        Set Global Variable    ${UOM_ORG_ID}    ${configured}
        RETURN
    END
    ${resolved}=    Get Variable Value    ${UOM_ORG_ID}    ${EMPTY}
    IF    $resolved    RETURN
    ${resp}=    GET On Session    api    /v1/iam/iam_org    params=${{ {'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Should Not Be Empty    ${items}
    ...    msg=No organization exists to own the UoM fixtures; set ORG_ID in the environment
    Set Global Variable    ${UOM_ORG_ID}    ${items}[0][id]

Unique Symbol
    [Documentation]    A UoM symbol is unique per organization, so it carries the same
    ...    {words}+{hash} treatment as every other unique column.
    [Arguments]    ${words}=uom
    ${suffix}=    Unique Suffix
    RETURN    ${words}${suffix}

Ensure Uom Category Under Test
    [Documentation]    Creates the category the whole uomcat suite exercises, and that
    ...    the uom suite hangs its units off. Lazy so any suite file runs standalone.
    ${id}=    Get Variable Value    ${UOMCAT_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Org Under Test
    ${name}=    Unique Display Name    Robot Weight Category
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${UOMCAT_ID}    ${id}
    Set Global Variable    ${UOMCAT_ETAG}    ${etag}
    Set Global Variable    ${UOMCAT_NAME}    ${name}

Ensure Reference Uom
    [Documentation]    Every category needs exactly one Reference UoM before conversion
    ...    means anything (BR-UOM-ESS-003). Factor is 1 by invariant (BR-UOM-ESS-006).
    ${id}=    Get Variable Value    ${REFERENCE_UOM_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Uom Category Under Test
    ${name}=    Unique Display Name    Robot Kilogram
    ${symbol}=    Unique Symbol    kg
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'reference', 'factor': '1', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${REFERENCE_UOM_ID}    ${id}
    Set Global Variable    ${REFERENCE_UOM_ETAG}    ${etag}

Ensure Uom Under Test
    [Documentation]    The unit the uom suite exercises: a "gram" against the category's
    ...    reference kilogram, so factor 0.001 makes it a `smaller` UoM.
    ${id}=    Get Variable Value    ${UOM_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Reference Uom
    ${name}=    Unique Display Name    Robot Gram
    ${symbol}=    Unique Symbol    g
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'smaller', 'factor': '0.001', 'rounding': '1', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${UOM_ID}    ${id}
    Set Global Variable    ${UOM_ETAG}    ${etag}
    Set Global Variable    ${UOM_SYMBOL}    ${symbol}

Ensure Foreign Uom Category
    [Documentation]    A second category, so the cross-category conversion rejection of
    ...    BR-UOM-ESS-012 has something to be rejected against.
    ${id}=    Get Variable Value    ${FOREIGN_UOMCAT_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Org Under Test
    ${name}=    Unique Display Name    Robot Length Category
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${FOREIGN_UOMCAT_ID}    ${id}
    ${mname}=    Unique Display Name    Robot Metre
    ${msymbol}=    Unique Symbol    m
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $mname}, 'symbol': $msymbol, 'category_id': $id, 'uom_type': 'reference', 'factor': '1', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ${mid}    ${metag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${FOREIGN_UOM_ID}    ${mid}

Ensure Referenceless Uom Category
    [Documentation]    A category deliberately left without a Reference UoM, so a test can
    ...    exercise a reference-specific rule (BR-UOM-ESS-006) in isolation. Creating a bad
    ...    reference in a category that already has one violates BR-UOM-ESS-005 as well,
    ...    and the response would then carry two violations instead of the one under test.
    ${id}=    Get Variable Value    ${BARE_UOMCAT_ID}    ${EMPTY}
    IF    $id    RETURN
    Ensure Org Under Test
    ${name}=    Unique Display Name    Robot Bare Category
    ${resp}=    POST On Session    api    ${UOMCAT_API}
    ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${BARE_UOMCAT_ID}    ${id}

Ensure Seeded Uom Categories
    [Documentation]    Bulk-creates sample categories for Search/Exists once per
    ...    execution. Every 10th carries "Lead" in the name so content-dependent search
    ...    filters match even on an otherwise empty database. Left in the DB by design.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_UOMCAT_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Org Under Test
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${remainder}=    Evaluate    $index % 10
        IF    $remainder == 0
            ${name}=    Unique Display Name    Seed Lead Category ${index}
        ELSE
            ${name}=    Unique Display Name    Seed Category ${index}
        END
        ${resp}=    POST On Session    api    ${UOMCAT_API}
        ...    json=${{ {'name': {'en-US': $name}, 'org_id': $UOM_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_UOMCAT_IDS}    ${ids}

Ensure Seeded Uoms
    [Documentation]    Bulk-creates sample units in the category under test, all
    ...    `bigger_equal` so no seed collides with the single-reference invariant
    ...    (BR-UOM-ESS-005). Shared by the Exists and Search suites.
    [Arguments]    ${count}=50
    ${ids}=    Get Variable Value    ${SEEDED_UOM_IDS}    ${None}
    IF    $ids is not None and len($ids) >= int($count)    RETURN
    Ensure Reference Uom
    ${ids}=    Create List
    FOR    ${index}    IN RANGE    ${count}
        ${remainder}=    Evaluate    $index % 10
        IF    $remainder == 0
            ${name}=    Unique Display Name    Seed Lead Unit ${index}
        ELSE
            ${name}=    Unique Display Name    Seed Unit ${index}
        END
        ${symbol}=    Unique Symbol    su
        ${resp}=    POST On Session    api    ${UOM_API}
        ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'bigger_equal', 'factor': '10', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
        Response Status Should Be    ${resp}    201
        Append To List    ${ids}    ${resp.json()}[id]
    END
    Set Global Variable    ${SEEDED_UOM_IDS}    ${ids}

Delete Uom Seed Data
    [Documentation]    Best-effort removal of the fixtures the UoM suites created on the
    ...    fly, units before categories — the category FK is ON DELETE NO ACTION, so a
    ...    category with surviving units cannot be removed (BR-UOM-ESS-002).
    ${foreign_uom}=    Get Variable Value    ${FOREIGN_UOM_ID}    ${EMPTY}
    IF    $foreign_uom
        DELETE On Session    api    ${UOM_API}/${foreign_uom}    expected_status=any
        Set Global Variable    ${FOREIGN_UOM_ID}    ${EMPTY}
    END
    ${foreign_cat}=    Get Variable Value    ${FOREIGN_UOMCAT_ID}    ${EMPTY}
    IF    $foreign_cat
        DELETE On Session    api    ${UOMCAT_API}/${foreign_cat}    expected_status=any
        Set Global Variable    ${FOREIGN_UOMCAT_ID}    ${EMPTY}
    END
    # Holds no UoM by design, so it needs no member cleanup before it goes.
    ${bare_cat}=    Get Variable Value    ${BARE_UOMCAT_ID}    ${EMPTY}
    IF    $bare_cat
        DELETE On Session    api    ${UOMCAT_API}/${bare_cat}    expected_status=any
        Set Global Variable    ${BARE_UOMCAT_ID}    ${EMPTY}
    END
    ${reference}=    Get Variable Value    ${REFERENCE_UOM_ID}    ${EMPTY}
    IF    $reference
        DELETE On Session    api    ${UOM_API}/${reference}    expected_status=any
        Set Global Variable    ${REFERENCE_UOM_ID}    ${EMPTY}
    END

Delete Uom Fixtures
    [Documentation]    Teardown for the uom suite: removes everything Delete Uom Seed Data
    ...    does, then the category itself. Units before categories — the category FK is
    ...    ON DELETE NO ACTION, so a category with surviving units cannot be removed.
    # The conversion suite's own units first: Delete Uom Seed Data drops the foreign
    # category, and a category with surviving units cannot be removed.
    FOR    ${name}    IN    GRAM_UOM_ID    TONNE_UOM_ID    YARD_UOM_ID
        ${id}=    Get Variable Value    ${${name}}    ${EMPTY}
        IF    $id
            DELETE On Session    api    ${UOM_API}/${id}    expected_status=any
            Set Global Variable    ${${name}}    ${EMPTY}
        END
    END
    Delete Uom Seed Data
    ${cat}=    Get Variable Value    ${UOMCAT_ID}    ${EMPTY}
    IF    $cat
        DELETE On Session    api    ${UOMCAT_API}/${cat}    expected_status=any
        Set Global Variable    ${UOMCAT_ID}    ${EMPTY}
        Set Global Variable    ${UOMCAT_ETAG}    ${EMPTY}
    END
