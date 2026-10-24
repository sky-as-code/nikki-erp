*** Settings ***
Documentation     Pairwise combination exclusions (CR01 §10.5, CR02 §4.2, PROD-015): exclusions are
...               saved with the configuration in catalogue value ids, a base exists only while some
...               completion stays valid, and a configuration that leaves nothing valid is refused.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Exclusion Fixture
Test Tags         inventory    product_configuration    exclusion


*** Test Cases ***
An Excluded Pair Keeps Its Base While Another Completion Is Valid
    [Documentation]    ADD22: Red-S excluded, Red-M allowed -> base Red still exists, and resolving
    ...    Red-S is refused.
    ${excl}=    Exclusion    ${COLORS}[red]    ${SIZES}[s]
    Configure Template Successfully    ${TEMPLATE_ID}    ${LINES}    ${{ [$excl] }}
    ${counts}=    Count Template Variants    ${TEMPLATE_ID}
    Should Be Equal As Integers    ${counts}[base]    2
    Should Be Equal As Integers    ${counts}[possible]    3
    ${resp}=    Resolve Selection    ${TEMPLATE_ID}
    ...    ${{ [{'attribute_id': $COLOR_ID, 'value_id': $COLORS['red']}, {'attribute_id': $SIZE_ID, 'value_id': $SIZES['s']}] }}
    Response Should Contain Business Error    ${resp}    inventory:product_template.selection_excluded
    ${config}=    Read Template Configuration    ${TEMPLATE_ID}
    Length Should Be    ${config}[exclusions]    1

Excluding Every Completion Of Red Withdraws Base Red
    [Documentation]    ADD23: with both sizes excluded for Red, no current base Red remains; Blue
    ...    is kept. The pair order is normalized, so reversed input is accepted.
    ${e1}=    Exclusion    ${COLORS}[red]    ${SIZES}[s]
    ${e2}=    Exclusion    ${SIZES}[m]    ${COLORS}[red]
    Configure Template Successfully    ${TEMPLATE_ID}    ${LINES}    ${{ [$e1, $e2] }}
    ${variants}=    Search Template Variants    ${TEMPLATE_ID}
    Length Should Be    ${variants}    1
    Should Be Equal    ${variants}[0][combination_key]    ${COLOR_ID}:${COLORS}[blue]

Excluding Every Completion Is Refused
    [Documentation]    ADD24 / AC24: nothing valid would remain, so the save is refused and the
    ...    template and its last variant stay as they were.
    [Tags]    negative
    ${before}=    Search Template Variants    ${TEMPLATE_ID}
    ${all}=    Evaluate    [{'left_value_id': c, 'right_value_id': s} for c, s in __import__('itertools').product($COLORS.values(), $SIZES.values())]
    ${resp}=    Configure Template    ${TEMPLATE_ID}    ${LINES}    ${all}
    Response Should Contain Business Error    ${resp}    inventory:product_template.no_valid_combination
    ${after}=    Search Template Variants    ${TEMPLATE_ID}
    Should Be Equal    ${after}    ${before}

Clearing Exclusions Brings Back Base Red
    [Documentation]    Sending an empty list removes the exclusions; omitting the field would have
    ...    left them unchanged.
    Configure Template Successfully    ${TEMPLATE_ID}    ${LINES}    ${{ [] }}
    ${counts}=    Count Template Variants    ${TEMPLATE_ID}
    Should Be Equal As Integers    ${counts}[base]    2
    Should Be Equal As Integers    ${counts}[possible]    4

Exclusions Are Read Only Over Rest
    [Documentation]    D1: junction rows are written only through the template payload.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${TEMPLATE_EXCLUSION_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'product_template_id': $TEMPLATE_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    POST On Session    api    ${TEMPLATE_ATTRIBUTE_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'product_template_id': $TEMPLATE_ID, 'attribute_id': $COLOR_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405


*** Keywords ***
Ensure Exclusion Fixture
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Excl Color
    ${size}    ${sizes}=    Create Attribute With Values    dynamic    s    m    words=Robot Excl Size
    ${lines}=    Create List
    ...    ${{ {'attribute_id': $color, 'value_ids': list($colors.values())} }}
    ...    ${{ {'attribute_id': $size, 'value_ids': list($sizes.values())} }}
    ${template_id}=    Create Configured Template    ${lines}    words=Robot Exclusion Template
    Set Suite Variable    ${COLOR_ID}    ${color}
    Set Suite Variable    ${COLORS}    ${colors}
    Set Suite Variable    ${SIZE_ID}    ${size}
    Set Suite Variable    ${SIZES}    ${sizes}
    Set Suite Variable    ${LINES}    ${lines}
    Set Suite Variable    ${TEMPLATE_ID}    ${template_id}
