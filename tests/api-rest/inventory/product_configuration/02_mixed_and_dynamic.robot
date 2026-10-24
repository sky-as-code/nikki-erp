*** Settings ***
Documentation     Mixed (instant + dynamic) and dynamic-only configurations (CR02 §3, §7): a mixed
...               template pre-creates one base per instant combination and no complete; a
...               dynamic-only template pre-creates nothing at all; adding a dynamic attribute never
...               rewrites an existing variant; removing every generating attribute gives the empty
...               complete back.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Mixed Catalogue
Test Tags         inventory    product_configuration    mixed


*** Test Cases ***
Mixed Template Pre-creates Bases Only
    [Documentation]    ADD12: 2 colours instant x 3 sizes dynamic -> 2 bases, 0 completes, 6
    ...    possible completes. ADD13: a never attribute with 4 options changes none of that.
    ${color}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    ${size}=    Template Line    ${SIZE_ID}    ${SIZES}[s]    ${SIZES}[m]    ${SIZES}[l]
    ${wrap}=    Template Line    ${WRAP_ID}    ${WRAPS}[w1]    ${WRAPS}[w2]    ${WRAPS}[w3]    ${WRAPS}[w4]
    ${template_id}=    Create Configured Template    ${{ [$color, $size, $wrap] }}    words=Robot Mixed Template
    ${counts}=    Count Template Variants    ${template_id}
    Should Be Equal As Integers    ${counts}[base]    2
    Should Be Equal As Integers    ${counts}[complete]    0
    Should Be Equal As Integers    ${counts}[possible]    6
    ${variants}=    Search Template Variants    ${template_id}
    FOR    ${variant}    IN    @{variants}
        Should Be Equal    ${variant}[variant_kind]    base
    END

Dynamic Only Template Pre-creates Nothing
    [Documentation]    ADD14 / ADD15: no empty base and no complete, even with one dynamic value;
    ...    the unused default goes. AC09: a listing never materializes anything.
    ${size}=    Template Line    ${SIZE_ID}    ${SIZES}[s]
    ${template_id}=    Create Configured Template    ${{ [$size] }}    words=Robot Dynamic Only
    ${variants}=    Search Template Variants    ${template_id}    include_archived=${True}
    Should Be Empty    ${variants}
    ${counts}=    Count Template Variants    ${template_id}
    Should Be Equal As Integers    ${counts}[possible]    1
    ${variants}=    Search Template Variants    ${template_id}
    Should Be Empty    ${variants}    msg=Reading must never create a variant
    Set Suite Variable    ${DYNAMIC_TEMPLATE_ID}    ${template_id}

Adding A Dynamic Value Creates No Complete
    [Documentation]    ADD25: a new dynamic value only widens what resolve may create.
    ${size}=    Template Line    ${SIZE_ID}    ${SIZES}[s]    ${SIZES}[m]
    Configure Template Successfully    ${DYNAMIC_TEMPLATE_ID}    ${{ [$size] }}
    ${variants}=    Search Template Variants    ${DYNAMIC_TEMPLATE_ID}
    Should Be Empty    ${variants}
    ${counts}=    Count Template Variants    ${DYNAMIC_TEMPLATE_ID}
    Should Be Equal As Integers    ${counts}[possible]    2

Adding A Dynamic Attribute To A Used Complete Archives It
    [Documentation]    ADD26 / CR02 §7.2: complete Red holds stock history; adding Size dynamic
    ...    does not turn it into a base or guess its size. It is archived (system_sync) and a new
    ...    base Red is created.
    ${color}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${template_id}=    Create Configured Template    ${{ [$color] }}    words=Robot Add Dynamic Template
    ${variants}=    Search Template Variants    ${template_id}
    ${red}=    Set Variable    ${variants}[0][id]
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${red}
    ${size}=    Template Line    ${SIZE_ID}    ${SIZES}[s]    ${SIZES}[m]
    Configure Template Successfully    ${template_id}    ${{ [$color, $size] }}
    ${all}=    Search Template Variants    ${template_id}    include_archived=${True}
    ${old}=    Variant By Id    ${all}    ${red}
    Should Be Equal    ${old}[is_archived]    ${True}
    Should Be Equal    ${old}[archive_source]    system_sync
    Should Be Equal    ${old}[variant_kind]    complete
    ${active}=    Search Template Variants    ${template_id}
    Length Should Be    ${active}    1
    Should Be Equal    ${active}[0][variant_kind]    base
    Should Not Be Equal    ${active}[0][id]    ${red}
    [Teardown]    Release Stock Move    ${transfer_id}    ${move_id}

Removing Every Generating Attribute Gives The Empty Complete Back
    [Documentation]    AC14 / CR02 §13: with no generating line left the template needs its empty
    ...    complete; a base is never turned into it.
    ${color}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${size}=    Template Line    ${SIZE_ID}    ${SIZES}[s]
    ${template_id}=    Create Configured Template    ${{ [$color, $size] }}    words=Robot Remove All Template
    ${bases}=    Search Template Variants    ${template_id}
    Configure Template Successfully    ${template_id}    ${{ [] }}
    ${variants}=    Search Template Variants    ${template_id}
    Length Should Be    ${variants}    1
    Should Be Equal    ${variants}[0][combination_key]    ${EMPTY}
    Should Be Equal    ${variants}[0][variant_kind]    complete
    Should Not Be Equal    ${variants}[0][id]    ${bases}[0][id]

Few Bases With Many Possible Completes Are Not Capped
    [Documentation]    ADD38: one base and 6^4 = 1296 possible completes is accepted, because the
    ...    cap counts pre-created records only; nothing is mass-created.
    ${color}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${lines}=    Create List    ${color}
    FOR    ${index}    IN RANGE    4
        ${attr}    ${values}=    Create Attribute With Values    dynamic    v0    v1    v2    v3    v4    v5
        ...    words=Robot Wide Dynamic ${index}
        Append To List    ${lines}    ${{ {'attribute_id': $attr, 'value_ids': list($values.values())} }}
    END
    ${template_id}=    Create Configured Template    ${lines}    words=Robot Wide Dynamic Template
    ${counts}=    Count Template Variants    ${template_id}
    Should Be Equal As Integers    ${counts}[base]    1
    Should Be Equal As Integers    ${counts}[complete]    0


*** Keywords ***
Ensure Mixed Catalogue
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Mix Color
    ${size}    ${sizes}=    Create Attribute With Values    dynamic    s    m    l    words=Robot Mix Size
    ${wrap}    ${wraps}=    Create Attribute With Values    never    w1    w2    w3    w4    words=Robot Mix Wrap
    Set Suite Variable    ${COLOR_ID}    ${color}
    Set Suite Variable    ${COLORS}    ${colors}
    Set Suite Variable    ${SIZE_ID}    ${size}
    Set Suite Variable    ${SIZES}    ${sizes}
    Set Suite Variable    ${WRAP_ID}    ${wrap}
    Set Suite Variable    ${WRAPS}    ${wraps}
