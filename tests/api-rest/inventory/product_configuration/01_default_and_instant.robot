*** Settings ***
Documentation     The default variant and the instant reconcile matrix (CR01 §6-§7, CR02 §7.2):
...               a template starts with one empty-key complete; adding one instant value extends
...               that same record (id, SKU, history kept); adding more creates new records and never
...               copies SKU or stock; adding several values at once never picks one to inherit V0.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Instant Catalogue
Suite Teardown    Release V0 Move
Test Tags         inventory    product_configuration    instant


*** Test Cases ***
A Template With Only Never Attributes Keeps One Empty Default
    [Documentation]    AC04 (CR02 §13): a never attribute generates nothing, so the template keeps
    ...    exactly one complete with the empty key; saving again creates no second one.
    ${template_id}    ${etag}=    Create Product Template    Robot Never Only Template
    ${line}=    Template Line    ${FINISH_ID}    ${FINISHES}[matte]    ${FINISHES}[gloss]
    Configure Template Successfully    ${template_id}    ${{ [$line] }}
    Configure Template Successfully    ${template_id}    ${{ [$line] }}
    ${variants}=    Search Template Variants    ${template_id}    include_archived=${True}
    Length Should Be    ${variants}    1
    Should Be Equal    ${variants}[0][combination_key]    ${EMPTY}
    Should Be Equal    ${variants}[0][variant_kind]    complete

Adding One Value Extends The Default Variant
    [Documentation]    AC05 / ADD28: V0 has a SKU and stock history (a move); adding Color=Red
    ...    rewrites V0's own combination to Red. Same id, same SKU, no new record.
    ${template_id}    ${etag}=    Create Product Template    Robot Extend Template
    ${variants}=    Search Template Variants    ${template_id}
    ${v0}=    Set Variable    ${variants}[0][id]
    ${sku}=    Unique Code    V0SKU
    ${resp}=    Patch Record    ${PRODUCT_VARIANT_API}    ${v0}    ${{ {'sku': $sku} }}
    Response Status Should Be    ${resp}    200
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${v0}
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    Configure Template Successfully    ${template_id}    ${{ [$line] }}
    ${variants}=    Search Template Variants    ${template_id}    include_archived=${True}
    Length Should Be    ${variants}    1
    Should Be Equal    ${variants}[0][id]    ${v0}
    Should Be Equal    ${variants}[0][sku]    ${sku}
    Should Be Equal    ${variants}[0][combination_key]    ${COLOR_ID}:${COLORS}[red]
    ${links}=    GET On Session    api    ${VARIANT_ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["product_variant_id","=","' + $v0 + '"]}'} }}
    Response Status Should Be    ${links}    200
    Length Should Be    ${links.json()}[items]    1    msg=The extend writes the variant's join row
    Set Suite Variable    ${EXTEND_TEMPLATE_ID}    ${template_id}
    Set Suite Variable    ${V0_ID}    ${v0}
    Set Suite Variable    ${V0_SKU}    ${sku}
    Set Suite Variable    ${V0_TRANSFER_ID}    ${transfer_id}
    Set Suite Variable    ${V0_MOVE_ID}    ${move_id}

Adding A Second Value Creates Only The New Combination
    [Documentation]    AC06: Red keeps its record; Blue is created with no SKU copied.
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    Configure Template Successfully    ${EXTEND_TEMPLATE_ID}    ${{ [$line] }}
    ${variants}=    Search Template Variants    ${EXTEND_TEMPLATE_ID}
    Length Should Be    ${variants}    2
    ${red}=    Variant By Id    ${variants}    ${V0_ID}
    Should Be Equal    ${red}[combination_key]    ${COLOR_ID}:${COLORS}[red]
    ${blue}=    Evaluate    next(v for v in $variants if v['id'] != '${V0_ID}')
    Should Be Equal    ${blue}[combination_key]    ${COLOR_ID}:${COLORS}[blue]
    Should Be True    ${{ not $blue.get('sku') }}    msg=A new combination must not inherit V0's SKU

Adding A Single Value Dimension Extends Every Variant
    [Documentation]    AC11: Material=Cotton (one value) is appended to Red and Blue alike; both keep
    ...    their ids.
    ${before}=    Search Template Variants    ${EXTEND_TEMPLATE_ID}
    ${color}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    ${material}=    Template Line    ${MATERIAL_ID}    ${MATERIALS}[cotton]
    Configure Template Successfully    ${EXTEND_TEMPLATE_ID}    ${{ [$color, $material] }}
    ${after}=    Search Template Variants    ${EXTEND_TEMPLATE_ID}
    ${ids_before}=    Evaluate    sorted(v['id'] for v in $before)
    ${ids_after}=    Evaluate    sorted(v['id'] for v in $after)
    Lists Should Be Equal    ${ids_after}    ${ids_before}
    FOR    ${variant}    IN    @{after}
        Should Contain    ${variant}[combination_key]    ${MATERIAL_ID}:${MATERIALS}[cotton]
    END
    ${red}=    Variant By Id    ${after}    ${V0_ID}
    Should Be Equal    ${red}[sku]    ${V0_SKU}

Adding Two Values At Once Does Not Hand V0 To Either
    [Documentation]    AC07: Red and Blue added together to a template whose V0 is used: both are
    ...    created fresh and V0 is archived by the configuration (system_sync), not reused.
    ${template_id}    ${etag}=    Create Product Template    Robot Two At Once Template
    ${variants}=    Search Template Variants    ${template_id}
    ${v0}=    Set Variable    ${variants}[0][id]
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${v0}
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    Configure Template Successfully    ${template_id}    ${{ [$line] }}
    ${active}=    Search Template Variants    ${template_id}
    Length Should Be    ${active}    2
    ${ids}=    Evaluate    [v['id'] for v in $active]
    Should Not Contain    ${ids}    ${v0}
    ${all}=    Search Template Variants    ${template_id}    include_archived=${True}
    ${old}=    Variant By Id    ${all}    ${v0}
    Should Be Equal    ${old}[is_archived]    ${True}
    Should Be Equal    ${old}[archive_source]    system_sync
    [Teardown]    Release Stock Move    ${transfer_id}    ${move_id}

An Unused V0 Is Deleted When Two Values Arrive At Once
    [Documentation]    AC07 / PROD-020: without usage the invalid default is deleted, not archived.
    ${template_id}    ${etag}=    Create Product Template    Robot Two At Once Unused
    ${variants}=    Search Template Variants    ${template_id}
    ${v0}=    Set Variable    ${variants}[0][id]
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    Configure Template Successfully    ${template_id}    ${{ [$line] }}
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${v0}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Two By Three Instant Creates Six Completes
    [Documentation]    ADD11: 2 colours x 3 sizes, all instant: 6 complete, 0 base.
    ${color}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    ${size}=    Template Line    ${SIZE_ID}    ${SIZES}[s]    ${SIZES}[m]    ${SIZES}[l]
    ${template_id}=    Create Configured Template    ${{ [$color, $size] }}    words=Robot Two By Three
    ${counts}=    Count Template Variants    ${template_id}
    Should Be Equal As Integers    ${counts}[complete]    6
    Should Be Equal As Integers    ${counts}[base]    0


*** Keywords ***
Ensure Instant Catalogue
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Cfg Color
    ${size}    ${sizes}=    Create Attribute With Values    instant    s    m    l    words=Robot Cfg Size
    ${material}    ${materials}=    Create Attribute With Values    instant    cotton    wool    words=Robot Cfg Material
    ${finish}    ${finishes}=    Create Attribute With Values    never    matte    gloss    words=Robot Cfg Finish
    Set Suite Variable    ${COLOR_ID}    ${color}
    Set Suite Variable    ${COLORS}    ${colors}
    Set Suite Variable    ${SIZE_ID}    ${size}
    Set Suite Variable    ${SIZES}    ${sizes}
    Set Suite Variable    ${MATERIAL_ID}    ${material}
    Set Suite Variable    ${MATERIALS}    ${materials}
    Set Suite Variable    ${FINISH_ID}    ${finish}
    Set Suite Variable    ${FINISHES}    ${finishes}

Release V0 Move
    ${transfer_id}=    Get Variable Value    ${V0_TRANSFER_ID}    ${EMPTY}
    IF    $transfer_id    Release Stock Move    ${transfer_id}    ${V0_MOVE_ID}
