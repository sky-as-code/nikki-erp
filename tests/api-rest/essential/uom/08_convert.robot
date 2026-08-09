*** Settings ***
Documentation     The shared conversion capability of BR-UOM-ESS-013. The success cases
...               are the requirement's own worked examples, so a failure here means the
...               engine disagrees with the specification rather than with this suite.
...               Runs before DELETE because it needs the units alive.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Conversion Fixtures
Test Tags         essential    uom    convert


*** Test Cases ***
Convert Two Tonnes To Grams Succeeds
    [Documentation]    BR-UOM-ESS-013's worked example, verbatim: with kg as the
    ...    reference, 2 ton x 1000 / 0.001 = 2,000,000 g.
    ${resp}=    POST On Session    api    ${UOM_CONVERT_API}
    ...    json=${{ {'quantity': '2', 'source_uom_id': $TONNE_UOM_ID, 'target_uom_id': $GRAM_UOM_ID} }}
    Response Status Should Be    ${resp}    200
    Quantity Should Equal    ${resp.json()}[exact_quantity]    2000000

Convert To Reference Succeeds
    [Documentation]    BR-UOM-ESS-007: 500 g normalizes to 0.5 kg.
    ${resp}=    POST On Session    api    ${UOM_TO_REFERENCE_API}
    ...    json=${{ {'quantity': '500', 'source_uom_id': $GRAM_UOM_ID} }}
    Response Status Should Be    ${resp}    200
    Quantity Should Equal    ${resp.json()}[exact_quantity]    0.5
    Should Be Equal    ${resp.json()}[reference_uom_id]    ${REFERENCE_UOM_ID}

Convert Keeps Intermediate Precision
    [Documentation]    BR-UOM-ESS-018: 1 yard is 0.9144 m exactly. The unrounded result
    ...    must carry that through — this is the assertion a float64 pipeline fails.
    ${resp}=    POST On Session    api    ${UOM_CONVERT_API}
    ...    json=${{ {'quantity': '1', 'source_uom_id': $YARD_UOM_ID, 'target_uom_id': $METRE_UOM_ID} }}
    Response Status Should Be    ${resp}    200
    Quantity Should Equal    ${resp.json()}[exact_quantity]    0.9144

Convert Applies Target Rounding
    [Documentation]    BR-UOM-ESS-015/016: rounding is applied once, at the end, and is a
    ...    step rather than a digit count. The metre rounds to 0.01, so 0.9144 m reads
    ...    back as 0.91 while exact_quantity keeps the full value.
    ${resp}=    POST On Session    api    ${UOM_CONVERT_API}
    ...    json=${{ {'quantity': '1', 'source_uom_id': $YARD_UOM_ID, 'target_uom_id': $METRE_UOM_ID} }}
    Response Status Should Be    ${resp}    200
    Quantity Should Equal    ${resp.json()}[quantity]    0.91
    Quantity Should Equal    ${resp.json()}[exact_quantity]    0.9144

Convert Across Categories Fails
    [Documentation]    BR-UOM-ESS-012 / UOM-ESS-INV-06: the category is the conversion
    ...    boundary. Weight to length has no meaning, so it is refused rather than
    ...    silently producing a number.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_CONVERT_API}
    ...    json=${{ {'quantity': '1', 'source_uom_id': $GRAM_UOM_ID, 'target_uom_id': $METRE_UOM_ID} }}
    ...    expected_status=any
    Response Should Be Uom Category Mismatch Error    ${resp}

Convert To Archived Uom Fails
    [Documentation]    BR-UOM-ESS-019 / UOM-ESS-INV-11: an archived UoM must not be the
    ...    target of a new conversion.
    [Tags]    negative
    ${resp}=    GET On Session    api    ${UOM_API}/${TONNE_UOM_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uom.json    200
    ${resp}=    POST On Session    api    ${UOM_API}/${TONNE_UOM_ID}/archived
    ...    json=${{ {'etag': $item['etag'], 'is_archived': True} }}
    Response Should Be Update Success    ${resp}    count=1

    ${resp}=    POST On Session    api    ${UOM_CONVERT_API}
    ...    json=${{ {'quantity': '1', 'source_uom_id': $GRAM_UOM_ID, 'target_uom_id': $TONNE_UOM_ID} }}
    ...    expected_status=any
    Response Should Be Uom Target Archived Error    ${resp}

Convert From Archived Uom Succeeds
    [Documentation]    BR-UOM-ESS-020: the mirror of the rule above. A historical quantity
    ...    recorded in a since-archived UoM must stay convertible, or archiving would
    ...    retroactively break old documents.
    ${resp}=    POST On Session    api    ${UOM_CONVERT_API}
    ...    json=${{ {'quantity': '2', 'source_uom_id': $TONNE_UOM_ID, 'target_uom_id': $GRAM_UOM_ID} }}
    Response Status Should Be    ${resp}    200
    Quantity Should Equal    ${resp.json()}[exact_quantity]    2000000

Convert With Not Found Uom Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_CONVERT_API}
    ...    json=${{ {'quantity': '1', 'source_uom_id': $NOT_FOUND_ID, 'target_uom_id': $GRAM_UOM_ID} }}
    ...    expected_status=any
    Response Status Should Be    ${resp}    400


*** Keywords ***
Ensure Conversion Fixtures
    [Documentation]    A weight category (kg reference, gram, tonne) and a length category
    ...    (metre reference, yard) — enough to exercise every conversion rule including
    ...    the cross-category refusal.
    Ensure Reference Uom
    Ensure Foreign Uom Category
    ${gram}=    Get Variable Value    ${GRAM_UOM_ID}    ${EMPTY}
    IF    not $gram
        ${id}=    Create Conversion Uom    ${UOMCAT_ID}    Robot Conv Gram    cg    smaller    0.001    1
        Set Global Variable    ${GRAM_UOM_ID}    ${id}
    END
    ${tonne}=    Get Variable Value    ${TONNE_UOM_ID}    ${EMPTY}
    IF    not $tonne
        ${id}=    Create Conversion Uom    ${UOMCAT_ID}    Robot Conv Tonne    ct    bigger_equal    1000    0.001
        Set Global Variable    ${TONNE_UOM_ID}    ${id}
    END
    # The foreign category's reference metre is the conversion target for the yard.
    Set Global Variable    ${METRE_UOM_ID}    ${FOREIGN_UOM_ID}
    ${resp}=    GET On Session    api    ${UOM_API}/${METRE_UOM_ID}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uom.json    200
    ${resp}=    PATCH On Session    api    ${UOM_API}/${METRE_UOM_ID}
    ...    json=${{ {'rounding': '0.01', 'etag': $item['etag']} }}
    Response Should Be Update Success    ${resp}    count=1
    ${yard}=    Get Variable Value    ${YARD_UOM_ID}    ${EMPTY}
    IF    not $yard
        ${id}=    Create Conversion Uom    ${FOREIGN_UOMCAT_ID}    Robot Conv Yard    cy    smaller    0.9144    0.01
        Set Global Variable    ${YARD_UOM_ID}    ${id}
    END

Create Conversion Uom
    [Arguments]    ${category_id}    ${words}    ${symbol_words}    ${uom_type}    ${factor}    ${rounding}
    ${name}=    Unique Display Name    ${words}
    ${symbol}=    Unique Symbol    ${symbol_words}
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $category_id, 'uom_type': $uom_type, 'factor': $factor, 'rounding': $rounding, 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}

Quantity Should Equal
    [Documentation]    Quantities travel as decimal strings, so compare them numerically
    ...    rather than as text: "0.5", "0.50" and "0.500" are the same quantity, and which
    ...    one the server emits is a formatting detail, not a contract.
    [Arguments]    ${actual}    ${expected}
    Should Be True
    ...    ${{ __import__('decimal').Decimal(str($actual)) == __import__('decimal').Decimal(str($expected)) }}
    ...    msg=Expected quantity ${expected} but got ${actual}
