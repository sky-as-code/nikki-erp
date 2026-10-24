*** Settings ***
Documentation     Selection resolution (BR §14.4, CR01 §5.3, CR02 §5): turning a template plus
...               chosen attribute values into the complete variant a transaction line must
...               reference, creating it on demand. The payload validation comes first — without it
...               a malformed `selections` would decode to an empty list, resolve to the empty
...               combination (a real and different variant) and answer 200 with the wrong product.
...               The rest runs against a mixed template: Color instant x Size dynamic x Wrap never.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
...               AND    Ensure Mixed Resolve Template
Test Tags         inventory    products    resolve


*** Test Cases ***
Resolve With No Selections Succeeds
    [Documentation]    BR §4.5 / AC04: a template with no variant-generating attributes resolves an
    ...    empty selection to the empty-key default it was created with. The legacy
    ...    `template_id` body field is still accepted beside `product_template_id`.
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'org_id': $INV_ORG_ID, 'template_id': $PRODUCT_TEMPLATE_ID} }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[combination_key]    ${EMPTY}
    Should Be Equal    ${resp.json()}[outcome]    existing
    Should Be Equal    ${resp.json()}[materialized]    ${False}

Resolve Without Template Id Fails
    [Documentation]    resolve_selection identifies its template through the body rather than
    ...    the path, so a missing product_template_id is a client error the caller can fix.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Template Id Required Error    ${resp}

Resolve With Empty Template Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'org_id': $INV_ORG_ID, 'product_template_id': ''} }}    expected_status=any
    Response Should Be Template Id Required Error    ${resp}

Resolve With Malformed Selections Fails
    [Documentation]    `selections` must be a list. A string silently decoding to an empty list
    ...    would resolve to the empty combination and return the wrong variant with a 200.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'org_id': $INV_ORG_ID, 'product_template_id': $PRODUCT_TEMPLATE_ID, 'selections': 'not-a-list'} }}
    ...    expected_status=any
    Response Should Be Selections Malformed Error    ${resp}

Resolve With Not Found Template Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${{ {'org_id': $INV_ORG_ID, 'product_template_id': $NOT_FOUND_ID} }}    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200
    ...    msg=Resolving against a template that does not exist must not succeed

Lookup Without Materialize Reports Missing And Writes Nothing
    [Documentation]    AC09: reading a combination never creates it. With
    ...    materialize_if_missing=false a valid but absent combination answers `missing` and names
    ...    its base; the template still holds only its two bases.
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED, $SEL_M] }}    materialize=${False}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[outcome]    missing
    Should Be Equal    ${resp.json()}[base_variant_id]    ${BASE_RED_ID}
    Dictionary Should Not Contain Key    ${resp.json()}    variant_id
    ${counts}=    Count Template Variants    ${MIXED_TEMPLATE_ID}
    Should Be Equal As Integers    ${counts}[complete]    0
    Should Be Equal As Integers    ${counts}[base]    2

First Resolve Creates One Complete Under Its Base
    [Documentation]    ADD16: resolving Red-M creates exactly one complete variant linked to base
    ...    Red; the base is unchanged. The never value is echoed back, never stored (CR02 §5).
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED, $SEL_M, $SEL_GIFT] }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[outcome]    created
    Should Be Equal    ${resp.json()}[materialized]    ${True}
    Should Be Equal    ${resp.json()}[base_variant_id]    ${BASE_RED_ID}
    Length Should Be    ${resp.json()}[never_selections]    1
    Should Be Equal    ${resp.json()}[never_selections][0][value_id]    ${SEL_GIFT}[value_id]
    ${key}=    Evaluate    '|'.join(a + ':' + v for a, v in sorted([($SEL_RED['attribute_id'], $SEL_RED['value_id']), ($SEL_M['attribute_id'], $SEL_M['value_id'])]))
    Should Be Equal    ${resp.json()}[combination_key]    ${key}
    ...    msg=The never dimension must stay out of the combination key
    Set Suite Variable    ${RED_M_ID}    ${resp.json()}[variant_id]
    ${variant}=    Read Record    ${PRODUCT_VARIANT_API}    ${RED_M_ID}    id,variant_kind,base_variant_id,is_transaction_eligible
    Should Be Equal    ${variant}[variant_kind]    complete
    Should Be Equal    ${variant}[base_variant_id]    ${BASE_RED_ID}
    Should Be Equal    ${variant}[is_transaction_eligible]    ${True}

Resolving Another Size Creates Another Complete
    [Documentation]    ADD17: Red-S is a new complete; Red-M is untouched.
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED, $SEL_S] }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[outcome]    created
    Should Not Be Equal    ${resp.json()}[variant_id]    ${RED_M_ID}
    ${counts}=    Count Template Variants    ${MIXED_TEMPLATE_ID}
    Should Be Equal As Integers    ${counts}[complete]    2

Resolving Again With Another Never Value Returns The Same Complete
    [Documentation]    ADD18 / ADD19 / AC10: the same always+dynamic selection with a different
    ...    never value, or in another order, answers the same id. The back-to-back pair stands in
    ...    for the concurrent case: Robot has no cheap parallel request (the only Python allowed is
    ...    the schema validator), and the race path — unique violation, re-read the winner — was
    ...    exercised with four parallel calls outside Robot during T08 (one created, three existing).
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED, $SEL_M, $SEL_PLAIN] }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[outcome]    existing
    Should Be Equal    ${resp.json()}[variant_id]    ${RED_M_ID}
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_M, $SEL_RED] }}
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[variant_id]    ${RED_M_ID}
    ${counts}=    Count Template Variants    ${MIXED_TEMPLATE_ID}
    Should Be Equal As Integers    ${counts}[complete]    2

Resolve With A Missing Dimension Fails
    [Documentation]    ADD20: every always and dynamic dimension needs exactly one value; nothing
    ...    is chosen implicitly.
    [Tags]    negative
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED] }}
    Response Should Contain Business Error    ${resp}    inventory:product_template.selection_missing_attribute

Resolve With Two Values Of One Attribute Fails
    [Tags]    negative
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED, $SEL_M, $SEL_S] }}
    Response Should Contain Business Error    ${resp}    inventory:product_template.selection_duplicate_attribute
    ...    selections

Resolve With A Stale Expected Etag Fails
    [Documentation]    The caller may pin the configuration version it rendered (AC26).
    [Tags]    negative
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED, $SEL_M] }}    expected_etag=1000000000000000000
    Response Should Be Etag Unmatched Error    ${resp}

Resolving A Manually Archived Complete Neither Duplicates Nor Restores It
    [Documentation]    ADD29 / AC16: a user archive is never undone by resolve and never shadowed
    ...    by a copy; the caller is told to unarchive it.
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${RED_M_ID}
    Response Status Should Be    ${resp}    200
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_RED, $SEL_M] }}
    Response Should Contain Business Error    ${resp}    inventory:product_variant.manually_archived
    ${variants}=    Search Template Variants    ${MIXED_TEMPLATE_ID}    include_archived=${True}
    ${completes}=    Variants Of Kind    ${variants}    complete
    Length Should Be    ${completes}    2
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${RED_M_ID}    ${False}
    Response Status Should Be    ${resp}    200

A Manually Archived Base Blocks Resolve Through It
    [Documentation]    ADD31: archiving base Blue stops new completes under it.
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${BASE_BLUE_ID}
    Response Status Should Be    ${resp}    200
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_BLUE, $SEL_S] }}
    Response Should Contain Business Error    ${resp}    inventory:product_variant.base_archived
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${BASE_BLUE_ID}    ${False}
    Response Status Should Be    ${resp}    200

Deleting A Base With Completes Fails
    [Documentation]    ADD30: a base never cascades its completes away.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${BASE_RED_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_variant.base_has_complete_variants    id

A Base Is Not Transaction Eligible
    [Documentation]    ADD21: a base variant cannot be named by a new stock move, and its computed
    ...    eligibility says so.
    [Tags]    negative
    ${base}=    Read Record    ${PRODUCT_VARIANT_API}    ${BASE_BLUE_ID}    id,is_transaction_eligible
    Should Be Equal    ${base}[is_transaction_eligible]    ${False}
    ${transfer_id}    ${resp}=    Post Stock Move For Variant    ${BASE_BLUE_ID}
    Response Should Contain Business Error    ${resp}    inventory:product_variant.not_transaction_eligible
    ...    product_variant_id
    DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Resolving On An Archived Template Fails
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_TEMPLATE_API}    ${MIXED_TEMPLATE_ID}
    Response Status Should Be    ${resp}    200
    ${resp}=    Resolve Selection    ${MIXED_TEMPLATE_ID}    ${{ [$SEL_BLUE, $SEL_M] }}
    Response Should Contain Business Error    ${resp}    inventory:product_template.archived
    ${resp}=    Set Record Archived    ${PRODUCT_TEMPLATE_API}    ${MIXED_TEMPLATE_ID}    ${False}
    Response Status Should Be    ${resp}    200

Removing The Last Dynamic Line Gives Completes Their Base Keys
    [Documentation]    ADD27: with Size removed, Color alone is instant, so the completes Red and
    ...    Blue are pre-created under the keys the bases held. The bases and the Size completes
    ...    become invalid, and a complete never takes over a base id (CR02 §7.2.5). A complete
    ...    may share its key with an old base: uniqueness is per variant kind.
    ${line_color}=    Template Line    ${COLOR_ID}    ${SEL_RED}[value_id]    ${SEL_BLUE}[value_id]
    Configure Template Successfully    ${MIXED_TEMPLATE_ID}    ${{ [$line_color] }}
    ${variants}=    Search Template Variants    ${MIXED_TEMPLATE_ID}
    ${completes}=    Variants Of Kind    ${variants}    complete
    ${bases}=    Variants Of Kind    ${variants}    base
    Length Should Be    ${completes}    2
    Should Be Empty    ${bases}
    ${keys}=    Evaluate    sorted(v['combination_key'] for v in $completes)
    ${expected}=    Evaluate    sorted([$RED_KEY, $BLUE_KEY])
    Lists Should Be Equal    ${keys}    ${expected}
    ${ids}=    Evaluate    {v['id'] for v in $completes}
    Should Not Contain    ${ids}    ${BASE_RED_ID}    msg=A base id must never become a complete
    Should Not Contain    ${ids}    ${RED_M_ID}    msg=A Size complete must not be merged into Red


*** Keywords ***
Ensure Mixed Resolve Template
    [Documentation]    Color{Red, Blue} instant x Size{S, M} dynamic x Wrap{Gift, Plain} never:
    ...    two bases, no complete, four possible completes (ADD12, ADD13).
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Resolve Color
    ${size}    ${sizes}=    Create Attribute With Values    dynamic    s    m    words=Robot Resolve Size
    ${wrap}    ${wraps}=    Create Attribute With Values    never    gift    plain    words=Robot Resolve Wrap
    ${lines}=    Create List
    ...    ${{ {'attribute_id': $color, 'value_ids': [$colors['red'], $colors['blue']]} }}
    ...    ${{ {'attribute_id': $size, 'value_ids': [$sizes['s'], $sizes['m']]} }}
    ...    ${{ {'attribute_id': $wrap, 'value_ids': [$wraps['gift'], $wraps['plain']]} }}
    ${template_id}=    Create Configured Template    ${lines}    words=Robot Resolve Template
    Set Suite Variable    ${MIXED_TEMPLATE_ID}    ${template_id}
    Set Suite Variable    ${COLOR_ID}    ${color}
    Set Suite Variable    ${SEL_RED}    ${{ {'attribute_id': $color, 'value_id': $colors['red']} }}
    Set Suite Variable    ${SEL_BLUE}    ${{ {'attribute_id': $color, 'value_id': $colors['blue']} }}
    Set Suite Variable    ${SEL_S}    ${{ {'attribute_id': $size, 'value_id': $sizes['s']} }}
    Set Suite Variable    ${SEL_M}    ${{ {'attribute_id': $size, 'value_id': $sizes['m']} }}
    Set Suite Variable    ${SEL_GIFT}    ${{ {'attribute_id': $wrap, 'value_id': $wraps['gift']} }}
    Set Suite Variable    ${SEL_PLAIN}    ${{ {'attribute_id': $wrap, 'value_id': $wraps['plain']} }}
    Set Suite Variable    ${RED_KEY}    ${color}:${colors}[red]
    Set Suite Variable    ${BLUE_KEY}    ${color}:${colors}[blue]
    ${variants}=    Search Template Variants    ${template_id}
    ${bases}=    Variants Of Kind    ${variants}    base
    Length Should Be    ${bases}    2    msg=A mixed template pre-creates one base per instant combination (ADD12)
    ${red}=    Evaluate    next(v['id'] for v in $bases if v['combination_key'] == '${RED_KEY}')
    ${blue}=    Evaluate    next(v['id'] for v in $bases if v['combination_key'] == '${BLUE_KEY}')
    Set Suite Variable    ${BASE_RED_ID}    ${red}
    Set Suite Variable    ${BASE_BLUE_ID}    ${blue}
    ${counts}=    Count Template Variants    ${template_id}
    Should Be Equal As Integers    ${counts}[possible]    4
    ...    msg=The never dimension must not multiply the possible completes (ADD13)
