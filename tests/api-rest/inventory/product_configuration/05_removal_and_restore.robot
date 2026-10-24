*** Settings ***
Documentation     Removing values and lines, restoring, and the archive precedence (CR01 §7,
...               CR02 §8, PROD-008, PROD-012, PROD-017, PROD-020): an invalid variant is deleted
...               only when nothing uses it, otherwise archived by the configuration (system_sync);
...               re-selecting restores the same id; a removed line keeps its history and a re-add
...               opens a new revision; a user archive outranks every automatic restore.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Removal Fixture
Suite Teardown    Release Removal Move
Test Tags         inventory    product_configuration    removal


*** Test Cases ***
Removing A Used Value Archives Its Variant
    [Documentation]    AC12 / AC20: Red has stock history (a move) but no stock; removing Red keeps
    ...    Blue and archives Red with archive_source=system_sync. The old record stays readable.
    ${blue}=    Template Line    ${COLOR_ID}    ${COLORS}[blue]
    Configure Template Successfully    ${TEMPLATE_ID}    ${{ [$blue] }}
    ${all}=    Search Template Variants    ${TEMPLATE_ID}    include_archived=${True}
    ${red}=    Variant By Id    ${all}    ${RED_ID}
    Should Be Equal    ${red}[is_archived]    ${True}
    Should Be Equal    ${red}[archive_source]    system_sync
    ${active}=    Search Template Variants    ${TEMPLATE_ID}
    Length Should Be    ${active}    1
    Should Be Equal    ${active}[0][id]    ${BLUE_ID}

A Configuration Archived Variant Cannot Be Unarchived By Hand
    [Documentation]    D3: the combination is no longer part of the configuration.
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${RED_ID}    ${False}
    Response Should Contain Business Error    ${resp}    inventory:product_variant.combination_not_valid    is_archived

Reselecting The Value Restores The Same Variant
    [Documentation]    AC15: Red comes back with its own id; no duplicate is created.
    ${both}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    Configure Template Successfully    ${TEMPLATE_ID}    ${{ [$both] }}
    ${active}=    Search Template Variants    ${TEMPLATE_ID}
    Length Should Be    ${active}    2
    ${red}=    Variant By Id    ${active}    ${RED_ID}
    Should Be True    ${{ not $red.get('archive_source') }}

Removing An Unused Value Deletes Its Variant
    [Documentation]    PROD-020: Blue was never used, so removing it deletes the record.
    ${red}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    Configure Template Successfully    ${TEMPLATE_ID}    ${{ [$red] }}
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${BLUE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

A Removed Line Keeps History And A Re-add Opens A New Revision
    [Documentation]    ADD35 / PROD-008: removing Size disables its line (revision 1 stays); adding
    ...    it back creates revision 2 with a fresh mode snapshot.
    ...    Runs on a template with no used variant, so every reconcile step is a plain
    ...    extend/delete and no archived record holds a target key.
    ${red}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${size}=    Template Line    ${SIZE_ID}    ${SIZES}[s]
    ${template_id}=    Create Configured Template    ${{ [$red, $size] }}    words=Robot Revision Template
    Configure Template Successfully    ${template_id}    ${{ [$red] }}
    Configure Template Successfully    ${template_id}    ${{ [$red, $size] }}
    ${config}=    Read Template Configuration    ${template_id}
    ${line}=    Evaluate    next(l for l in $config['attribute_lines'] if l['attribute_id'] == '${SIZE_ID}')
    Should Be Equal As Integers    ${line}[application_revision]    2
    Should Be Equal    ${line}[variant_creation_snapshot]    instant
    ${resp}=    GET On Session    api    ${TEMPLATE_ATTRIBUTE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'include_archived': 'true', 'fields': 'id,is_enabled,application_revision', 'graph': '{"and":[{"if":["product_template_id","=","' + $template_id + '"]},{"if":["attribute_id","=","' + $SIZE_ID + '"]}]}'} }}
    Response Status Should Be    ${resp}    200
    ${revisions}=    Evaluate    sorted((l['application_revision'], l.get('is_enabled')) for l in $resp.json()['items'])
    Should Be Equal    ${revisions}    ${{ [(1, False), (2, True)] }}

A User Archive Survives A Template Unarchive
    [Documentation]    AC17 / PROD-017: a variant the user archived stays archived when its template
    ...    is archived and unarchived; only the cascade's own archives are lifted.
    ${both}=    Template Line    ${COLOR_ID}    ${COLORS}[red]    ${COLORS}[blue]
    ${template_id}=    Create Configured Template    ${{ [$both] }}    words=Robot Precedence Template
    ${variants}=    Search Template Variants    ${template_id}
    ${user_id}=    Set Variable    ${variants}[0][id]
    ${cascade_id}=    Set Variable    ${variants}[1][id]
    ${resp}=    Set Record Archived    ${PRODUCT_VARIANT_API}    ${user_id}
    Response Status Should Be    ${resp}    200
    ${resp}=    Set Record Archived    ${PRODUCT_TEMPLATE_API}    ${template_id}
    Response Status Should Be    ${resp}    200
    ${all}=    Search Template Variants    ${template_id}    include_archived=${True}
    ${user}=    Variant By Id    ${all}    ${user_id}
    Should Be Equal    ${user}[archive_source]    user    msg=A weaker cascade must not overwrite a user archive
    ${resp}=    Set Record Archived    ${PRODUCT_TEMPLATE_API}    ${template_id}    ${False}
    Response Status Should Be    ${resp}    200
    ${all}=    Search Template Variants    ${template_id}    include_archived=${True}
    ${user}=    Variant By Id    ${all}    ${user_id}
    ${cascade}=    Variant By Id    ${all}    ${cascade_id}
    Should Be Equal    ${user}[is_archived]    ${True}
    Should Be Equal    ${cascade}[is_archived]    ${False}

Deleting The Last Complete Variant Fails
    [Documentation]    D7: the last active complete of a template without dynamic attributes is
    ...    archived with the template, never deleted on its own.
    [Tags]    negative
    ${red}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${template_id}=    Create Configured Template    ${{ [$red] }}    words=Robot Last Complete Template
    ${variants}=    Search Template Variants    ${template_id}
    ${resp}=    DELETE On Session    api    ${PRODUCT_VARIANT_API}/${variants}[0][id]
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_variant.last_complete_variant    id


A Discontinued Template Makes Its Variants Ineligible
    [Documentation]    PROD-019 / D5: `is_transaction_eligible` follows EffectiveProduct.IsSelectable,
    ...    which refuses a discontinued template even while the template and variant are active.
    ${template_id}    ${etag}=    Create Product Template    Robot Discontinued Template
    ${variants}=    Search Template Variants    ${template_id}
    ${before}=    Read Record    ${PRODUCT_VARIANT_API}    ${variants}[0][id]    id,is_transaction_eligible
    Should Be Equal    ${before}[is_transaction_eligible]    ${True}
    ${resp}=    Patch Record    ${PRODUCT_TEMPLATE_API}    ${template_id}    ${{ {'status': 'discontinued'} }}
    Response Status Should Be    ${resp}    200
    ${after}=    Read Record    ${PRODUCT_VARIANT_API}    ${variants}[0][id]    id,is_transaction_eligible
    Should Be Equal    ${after}[is_transaction_eligible]    ${False}

*** Keywords ***
Ensure Removal Fixture
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Removal Color
    ${size}    ${sizes}=    Create Attribute With Values    instant    s    words=Robot Removal Size
    ${both}=    Template Line    ${color}    ${colors}[red]    ${colors}[blue]
    ${template_id}=    Create Configured Template    ${{ [$both] }}    words=Robot Removal Template
    ${variants}=    Search Template Variants    ${template_id}
    ${red_id}=    Evaluate    next(v['id'] for v in $variants if v['combination_key'].endswith('${colors}[red]'))
    ${blue_id}=    Evaluate    next(v['id'] for v in $variants if v['combination_key'].endswith('${colors}[blue]'))
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${red_id}
    Set Suite Variable    ${COLOR_ID}    ${color}
    Set Suite Variable    ${COLORS}    ${colors}
    Set Suite Variable    ${SIZE_ID}    ${size}
    Set Suite Variable    ${SIZES}    ${sizes}
    Set Suite Variable    ${TEMPLATE_ID}    ${template_id}
    Set Suite Variable    ${RED_ID}    ${red_id}
    Set Suite Variable    ${BLUE_ID}    ${blue_id}
    Set Suite Variable    ${REMOVAL_TRANSFER_ID}    ${transfer_id}
    Set Suite Variable    ${REMOVAL_MOVE_ID}    ${move_id}

Release Removal Move
    ${transfer_id}=    Get Variable Value    ${REMOVAL_TRANSFER_ID}    ${EMPTY}
    IF    $transfer_id    Release Stock Move    ${transfer_id}    ${REMOVAL_MOVE_ID}
