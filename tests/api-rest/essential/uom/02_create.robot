*** Settings ***
Documentation     Creating Units of Measure. The first test saves the unit under test
...               (${UOM_ID}/${UOM_ETAG}) consumed by the later suites and deleted last
...               by 08_delete.robot. The negatives are the business invariants of
...               BR-UOM-ESS-005/006/009/017 — they are the point of this suite.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Reference Uom
Test Tags         essential    uom    create


*** Test Cases ***
Create Smaller Uom Succeeds
    [Documentation]    A gram against the reference kilogram: BR-UOM-ESS-008's worked
    ...    example, where 1 g = 0.001 kg is expressed as the factor, never as rounding.
    ${name}=    Unique Display Name    Robot Gram
    ${symbol}=    Unique Symbol    g
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'smaller', 'factor': '0.001', 'rounding': '1', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${UOM_ID}    ${id}
    Set Global Variable    ${UOM_ETAG}    ${etag}
    Set Global Variable    ${UOM_SYMBOL}    ${symbol}

Create Bigger Equal Uom Succeeds
    ${name}=    Unique Display Name    Robot Tonne
    ${symbol}=    Unique Symbol    t
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'bigger_equal', 'factor': '1000', 'rounding': '0.001', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    Set Global Variable    ${TONNE_UOM_ID}    ${id}

Created Factor Keeps Full Precision
    [Documentation]    BR-UOM-ESS-018: a pound is 0.453592 kg. Reading it back unchanged
    ...    is what proves the decimal column and the string transport actually hold it —
    ...    a float64 round trip is where this silently becomes 0.45359199999999997.
    ${name}=    Unique Display Name    Robot Pound
    ${symbol}=    Unique Symbol    lb
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'smaller', 'factor': '0.453592', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${resp}=    GET On Session    api    ${UOM_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${ESSENTIAL_SCHEMA_DIR}/uom.json    200
    Should Be True    ${{ __import__('decimal').Decimal($item['factor']) == __import__('decimal').Decimal('0.453592') }}
    ...    msg=Factor lost precision in the round trip: ${item}[factor]
    DELETE On Session    api    ${UOM_API}/${id}    expected_status=any

Create Second Reference In Category Fails
    [Documentation]    BR-UOM-ESS-005 / UOM-ESS-INV-09: a category has exactly one
    ...    Reference UoM. Two references would make every factor in the category
    ...    ambiguous about what it is relative to.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Second Reference
    ${symbol}=    Unique Symbol    ref2
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'reference', 'factor': '1', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Uom Duplicate Reference Error    ${resp}

Create Reference With Factor Other Than One Fails
    [Documentation]    BR-UOM-ESS-006: the Reference UoM is the unit everything else is
    ...    measured against, so its own factor can only be 1.
    [Tags]    negative
    # A category with no reference yet, so only BR-UOM-ESS-006 can be violated here.
    Ensure Referenceless Uom Category
    ${name}=    Unique Display Name    Robot Bad Reference
    ${symbol}=    Unique Symbol    badref
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $BARE_UOMCAT_ID, 'uom_type': 'reference', 'factor': '1000', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Uom Reference Factor Error    ${resp}

Create Bigger Equal With Factor Below One Fails
    [Documentation]    BR-UOM-ESS-009.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Bigger
    ${symbol}=    Unique Symbol    badbig
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'bigger_equal', 'factor': '0.001', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Uom Bigger Equal Factor Error    ${resp}

Create Smaller With Factor One Fails
    [Documentation]    BR-UOM-ESS-009: `smaller` is strictly below 1, so the boundary
    ...    value belongs to bigger_equal instead.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Smaller
    ${symbol}=    Unique Symbol    badsml
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'smaller', 'factor': '1', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Uom Smaller Factor Error    ${resp}

Create Smaller With Zero Factor Fails
    [Documentation]    BR-UOM-ESS-009: a zero factor would make every conversion through
    ...    this unit a division by zero.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Zero Factor
    ${symbol}=    Unique Symbol    zero
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'smaller', 'factor': '0', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Uom Smaller Factor Error    ${resp}

Create With Rounding Above One Fails
    [Documentation]    BR-UOM-ESS-017: 0 <= rounding <= 1. A step of exactly 1 is the
    ...    "whole units only" precision of a discrete UoM, so the upper bound is
    ...    inclusive; anything beyond it is not a rounding step at all.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Bad Rounding
    ${symbol}=    Unique Symbol    badrnd
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $symbol, 'category_id': $UOMCAT_ID, 'uom_type': 'bigger_equal', 'factor': '10', 'rounding': '1.5', 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Uom Rounding Range Error    ${resp}

Create With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${UOM_API}    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}
    ...    name    symbol    category_id    uom_type    factor    rounding    org_id

Create With Malformed Payload Fails
    [Tags]    negative
    ${headers}=    Create Dictionary    Content-Type=application/json
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    data={ "symbol": "broken",    headers=${headers}    expected_status=any
    Response Should Be Malformed Payload Error    ${resp}

Create With Duplicate Symbol Fails
    [Documentation]    symbol is unique per organization.
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Duplicate Symbol
    ${resp}=    POST On Session    api    ${UOM_API}
    ...    json=${{ {'name': {'en-US': $name}, 'symbol': $UOM_SYMBOL, 'category_id': $UOMCAT_ID, 'uom_type': 'bigger_equal', 'factor': '10', 'rounding': '0.01', 'org_id': $UOM_ORG_ID} }}
    ...    expected_status=any
    Response Should Be Duplicate Values Error    ${resp}    symbol    org_id
