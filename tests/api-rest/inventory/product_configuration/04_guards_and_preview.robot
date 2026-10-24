*** Settings ***
Documentation     The guards around a configuration save (CR01 §5.2, §9, PROD-010, PROD-016): the
...               pre-create cap (all or nothing), the template etag, one transaction with the
...               template fields, row-addressed errors, foreign values, key conflicts, and the
...               preview that dry-runs the same payload.
Resource          resources/inventory_configuration.resource
Resource          resources/iam.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Guard Fixture
Test Tags         inventory    product_configuration    guard


*** Test Cases ***
Preview Reports The Plan Without Writing
    [Documentation]    PROD-010: adding Red to the empty default previews one extend; nothing
    ...    changes until the PATCH.
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${resp}=    Preview Template Configuration    ${TEMPLATE_ID}    ${{ [$line] }}
    Response Status Should Be    ${resp}    200
    ${preview}=    Set Variable    ${resp.json()}
    Should Be Equal As Integers    ${preview}[counts][extend]    1
    Should Be Equal As Integers    ${preview}[counts][create]    0
    Should Be Equal As Integers    ${preview}[pre_create_count]    1
    Should Be Equal As Integers    ${preview}[possible_complete_count]    1
    Should Be Equal    ${preview}[changes_identity]    ${True}
    Should Be Equal    ${preview}[extend][0][new_combination_key]    ${COLOR_ID}:${COLORS}[red]
    ${variants}=    Search Template Variants    ${TEMPLATE_ID}
    Should Be Equal    ${variants}[0][combination_key]    ${EMPTY}    msg=A preview must not write

Preview Lists Archive And Create Outcomes
    [Documentation]    PROD-010: replacing a used Red by Blue previews one archive (with its
    ...    reason) and one create.
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    Configure Template Successfully    ${TEMPLATE_ID}    ${{ [$line] }}
    ${variants}=    Search Template Variants    ${TEMPLATE_ID}
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${variants}[0][id]
    ${blue}=    Template Line    ${COLOR_ID}    ${COLORS}[blue]
    ${resp}=    Preview Template Configuration    ${TEMPLATE_ID}    ${{ [$blue] }}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[counts][archive]    1
    Should Be Equal As Integers    ${resp.json()}[counts][create]    1
    Should Be Equal    ${resp.json()}[archive][0][variant_id]    ${variants}[0][id]
    Should Be Equal    ${resp.json()}[archive][0][archive_reason]    ever_used
    [Teardown]    Release Stock Move    ${transfer_id}    ${move_id}

Preview With A Stale Etag Fails
    [Tags]    negative
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${TEMPLATE_ID}/preview_attribute_configuration
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': '1000000000000000000', 'attribute_lines': [$line]} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Apply After The Configuration Changed Fails
    [Documentation]    AC26: the client previewed at etag E; someone saved meanwhile; applying with
    ...    E is refused, so the client must preview again.
    [Tags]    negative
    ${stale}=    Read Etag    ${PRODUCT_TEMPLATE_API}    ${TEMPLATE_ID}
    ${resp}=    Patch Record    ${PRODUCT_TEMPLATE_API}    ${TEMPLATE_ID}    ${{ {'short_name': {'en-US': 'moved on'}} }}
    Response Status Should Be    ${resp}    200
    ${line}=    Template Line    ${COLOR_ID}    ${COLORS}[blue]
    ${resp}=    PATCH On Session    api    ${PRODUCT_TEMPLATE_API}/${TEMPLATE_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $stale, 'attribute_lines': [$line]} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

A Bad Line Rolls Back The Template Fields Too
    [Documentation]    PROD-009: the name change and the configuration share one transaction; the
    ...    unknown attribute in row 1 refuses both, and the error names the row.
    [Tags]    negative
    ${before}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${TEMPLATE_ID}    id,name
    ${good}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${bad}=    Template Line    ${NOT_FOUND_ID}    ${COLORS}[red]
    ${resp}=    Configure Template    ${TEMPLATE_ID}    ${{ [$good, $bad] }}    name=${{ {'en-US': 'Robot Renamed'} }}
    Response Should Contain Business Error    ${resp}    common:err_not_found    attribute_lines[1].attribute_id
    ${after}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${TEMPLATE_ID}    id,name
    Should Be Equal    ${after}[name]    ${before}[name]

A Value Of Another Attribute Is Refused
    [Documentation]    AC27 / ADD39: a value that does not belong to the line's attribute (or to the
    ...    org) is refused; nothing is written across attributes.
    [Tags]    negative
    ${line}=    Template Line    ${COLOR_ID}    ${SIZES}[s]
    ${resp}=    Configure Template    ${TEMPLATE_ID}    ${{ [$line] }}
    Response Should Contain Business Error    ${resp}    inventory:product_template.value_not_of_attribute
    ...    attribute_lines[0].value_ids

A Value From Another Organization Is Refused
    [Documentation]    AC27: ids that do not exist in the caller's org are not found, never borrowed.
    [Tags]    negative
    ${other}=    Not Found Id Variant    7
    ${line}=    Template Line    ${COLOR_ID}    ${other}
    ${resp}=    Configure Template    ${TEMPLATE_ID}    ${{ [$line] }}
    Should Be Equal As Integers    ${resp.status_code}    400

Exceeding The Pre-create Cap Creates Nothing
    [Documentation]    AC23 / ADD37: 6^4 = 1296 instant combinations is over the 1000 cap. The save
    ...    is refused with the count and the limit, and the template keeps its one default.
    [Tags]    negative
    ${template_id}    ${etag}=    Create Product Template    Robot Cap Template
    ${lines}=    Create List
    FOR    ${index}    IN RANGE    4
        ${attr}    ${values}=    Create Attribute With Values    instant    v0    v1    v2    v3    v4    v5
        ...    words=Robot Cap Attribute ${index}
        Append To List    ${lines}    ${{ {'attribute_id': $attr, 'value_ids': list($values.values())} }}
    END
    ${resp}=    Configure Template    ${template_id}    ${lines}
    ${error}=    Response Should Contain Business Error    ${resp}    inventory:product_template.too_many_variants
    Should Be Equal As Integers    ${error}[vars][count]    1296
    Should Be Equal As Integers    ${error}[vars][limit]    1000
    ${variants}=    Search Template Variants    ${template_id}    include_archived=${True}
    Length Should Be    ${variants}    1

Extending Onto A Key Another Record Holds Is A Conflict
    [Documentation]    CR01 §6.2: a variant may take a new single value only when no record,
    ...    archived included, holds the target key. Red (used) is archived when lines are
    ...    cleared; re-adding Red would extend the new empty default onto Red's key, so the save
    ...    reports a conflict instead of merging.
    [Tags]    negative
    ${template_id}    ${etag}=    Create Product Template    Robot Conflict Template
    ${red}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    Configure Template Successfully    ${template_id}    ${{ [$red] }}
    ${variants}=    Search Template Variants    ${template_id}
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${variants}[0][id]
    Configure Template Successfully    ${template_id}    ${{ [] }}
    ${resp}=    Configure Template    ${template_id}    ${{ [$red] }}
    Response Should Contain Business Error    ${resp}    inventory:product_template.combination_conflict
    [Teardown]    Release Stock Move    ${transfer_id}    ${move_id}

Duplicate Lines For One Attribute Are Refused
    [Tags]    negative
    ${a}=    Template Line    ${COLOR_ID}    ${COLORS}[red]
    ${b}=    Template Line    ${COLOR_ID}    ${COLORS}[blue]
    ${resp}=    Configure Template    ${TEMPLATE_ID}    ${{ [$a, $b] }}
    Should Be Equal As Integers    ${resp.status_code}    400

Only Sales Price Extra Is Editable On A Template Value
    [Documentation]    BR-PRICE-VARIANT-001: the per-template surcharge is edited in place with
    ...    PATCH; selection, sequence and references stay with the configuration payload, and the
    ...    row cannot be created, deleted or archived over REST.
    ${attr}    ${values}=    Create Attribute With Values    never    plain    words=Robot Price Extra
    ${line}=    Template Line    ${attr}    ${values}[plain]
    ${template_id}=    Create Configured Template    ${{ [$line] }}    words=Robot Price Extra Template
    ${resp}=    GET On Session    api    ${TEMPLATE_ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': 'id', 'graph': '{"if":["attribute_value_id","=","' + $values['plain'] + '"]}'} }}
    Response Status Should Be    ${resp}    200
    ${row_id}=    Set Variable    ${resp.json()}[items][0][id]
    ${resp}=    Patch Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${row_id}    ${{ {'sales_price_extra': '12500'} }}
    Response Status Should Be    ${resp}    200
    ${row}=    Read Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${row_id}    id,sales_price_extra,is_selected
    Should Be Equal As Numbers    ${row}[sales_price_extra]    12500
    Should Be Equal    ${row}[is_selected]    ${True}
    ${resp}=    Patch Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${row_id}
    ...    ${{ {'sales_price_extra': '1', 'is_selected': False} }}
    Response Should Contain Business Error    ${resp}
    ...    inventory:product_template_attribute_value.field_read_only    is_selected
    ${row}=    Read Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${row_id}    id,sales_price_extra,is_selected
    Should Be Equal As Numbers    ${row}[sales_price_extra]    12500
    Should Be Equal    ${row}[is_selected]    ${True}
    ${resp}=    POST On Session    api    ${TEMPLATE_ATTRIBUTE_VALUE_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'attribute_value_id': $values['plain']} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    DELETE On Session    api    ${TEMPLATE_ATTRIBUTE_VALUE_API}/${row_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    405
    ${resp}=    Set Record Archived    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${row_id}
    Should Be True    ${resp.status_code} in (404, 405)

Template Values Of One Line List Their Value Name And Code
    [Documentation]    The extra-price list filters by template line and shows each value by name
    ...    and code, read through the edges; the related fields are not writable.
    ${attr}    ${values}=    Create Attribute With Values    dynamic    big    small    words=Robot Price List
    ${line}=    Template Line    ${attr}    ${values}[big]    ${values}[small]
    ${template_id}=    Create Configured Template    ${{ [$line] }}    words=Robot Price List Template
    ${resp}=    GET On Session    api    ${TEMPLATE_ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': 'id,template_attribute_id', 'graph': '{"if":["attribute_value_id","=","' + $values['big'] + '"]}'} }}
    Response Status Should Be    ${resp}    200
    ${line_id}=    Set Variable    ${resp.json()}[items][0][template_attribute_id]
    ${resp}=    GET On Session    api    ${TEMPLATE_ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["template_attribute_id","=","' + $line_id + '"]}'} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[items]
    Length Should Be    ${items}    2
    ${codes}=    Evaluate    sorted(row['attribute_value_code'] for row in $items)
    Should Be Equal    ${codes}    ${{ ['big', 'small'] }}
    ${big}=    Evaluate    next(row for row in $items if row['attribute_value_code'] == 'big')
    Should Be Equal    ${big}[attribute_value_name][en-US]    BIG
    Dictionary Should Contain Key    ${big}    sales_price_extra
    Dictionary Should Contain Key    ${big}    is_selected
    ${resp}=    Patch Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${big}[id]    ${{ {'sales_price_extra': '3000'} }}
    Response Status Should Be    ${resp}    200
    ${row}=    Read Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${big}[id]
    ...    id,sales_price_extra,attribute_value_code,attribute_name
    Should Be Equal As Numbers    ${row}[sales_price_extra]    3000
    Should Be Equal    ${row}[attribute_value_code]    big
    Should Start With    ${row}[attribute_name][en-US]    Robot Price List
    ${resp}=    Patch Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${big}[id]    ${{ {'attribute_value_code': 'x'} }}
    Should Be Equal As Integers    ${resp.status_code}    400


Updating A Template Value Needs Update Value On The Attribute
    [Documentation]    PFU-010: PATCH of a template value's extra price asserts update_value on
    ...    inventory_product_attribute, not update on the configuration row. A user holding only
    ...    the system User role is refused. Needs PLAIN_USER_USERNAME / PLAIN_USER_PASSWORD;
    ...    skips without them.
    [Tags]    negative    permission
    ${username}=    Get Variable Value    ${PLAIN_USER_USERNAME}    ${EMPTY}
    ${password}=    Get Variable Value    ${PLAIN_USER_PASSWORD}    ${EMPTY}
    IF    not $username or not $password
        Skip    No plain-role account configured; set PLAIN_USER_USERNAME and PLAIN_USER_PASSWORD
    END
    ${attr}    ${values}=    Create Attribute With Values    never    guarded    words=Robot Price Guard
    ${line}=    Template Line    ${attr}    ${values}[guarded]
    ${template_id}=    Create Configured Template    ${{ [$line] }}    words=Robot Price Guard Template
    ${resp}=    GET On Session    api    ${TEMPLATE_ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': 'id', 'graph': '{"if":["attribute_value_id","=","' + $values['guarded'] + '"]}'} }}
    Response Status Should Be    ${resp}    200
    ${row_id}=    Set Variable    ${resp.json()}[items][0][id]
    Open Session As    plain_user    ${username}    ${password}
    ${resp}=    PATCH On Session    plain_user    ${TEMPLATE_ATTRIBUTE_VALUE_API}/${row_id}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'sales_price_extra': '999'} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user without update_value on inventory_product_attribute must not change an extra price
    ${row}=    Read Record    ${TEMPLATE_ATTRIBUTE_VALUE_API}    ${row_id}    id,sales_price_extra
    Should Be Equal As Numbers    ${row}[sales_price_extra]    0


*** Keywords ***
Ensure Guard Fixture
    ${color}    ${colors}=    Create Attribute With Values    instant    red    blue    words=Robot Guard Color
    ${size}    ${sizes}=    Create Attribute With Values    instant    s    m    words=Robot Guard Size
    ${template_id}    ${etag}=    Create Product Template    Robot Guard Template
    Set Suite Variable    ${COLOR_ID}    ${color}
    Set Suite Variable    ${COLORS}    ${colors}
    Set Suite Variable    ${SIZE_ID}    ${size}
    Set Suite Variable    ${SIZES}    ${sizes}
    Set Suite Variable    ${TEMPLATE_ID}    ${template_id}
