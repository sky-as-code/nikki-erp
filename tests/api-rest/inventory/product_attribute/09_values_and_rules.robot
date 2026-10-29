*** Settings ***
Documentation     Attribute and value rules of CR-INV-PRODUCT-VARIANTS (PROD-001..PROD-007): the
...               nested `values` delta saved atomically with the attribute, value-name
...               de-duplication, colour hex, display type, the variant-creation-mode lock, and the
...               archive/delete guards driven by template usage. Every test builds its own
...               attributes, so this file does not depend on the attribute under test that
...               08_delete removed.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
Test Tags         inventory    product_attribute    rules


*** Test Cases ***
Create With Nested Values Saves Them In One Request
    [Documentation]    PROD-001 / AC02: an attribute and its values are created together.
    ${attr}    ${values}=    Create Attribute With Values    instant    red    green    blue
    Length Should Be    ${values}    3
    ${item}=    Read Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}
    ...    id,current_template_usage_count,historical_template_usage_count,can_change_variant_creation
    Should Be Equal As Integers    ${item}[current_template_usage_count]    0
    Should Be Equal As Integers    ${item}[historical_template_usage_count]    0
    Should Be Equal    ${item}[can_change_variant_creation]    ${True}

Create With One Bad Value Row Saves Nothing And Names The Row
    [Documentation]    AC02: a bad row (here a malformed colour) rolls the whole create back, and
    ...    the error points at that row by path.
    [Tags]    negative
    ${code}=    Unique Code    badrow
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $code}, 'data_type': 'option', 'variant_creation_mode': 'instant', 'org_id': $INV_ORG_ID, 'values': {'add': [{'code': 'A', 'name': {'en-US': 'A'}}, {'code': 'B', 'name': {'en-US': 'B'}, 'color_hex': '#12345G'}]}} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    common:err_invalid_string_pattern    values.add[1].color_hex
    ${resp}=    GET On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["code","=","' + $code + '"]}'} }}
    Response Status Should Be    ${resp}    200
    Should Be Empty    ${resp.json()}[items]    msg=A refused create must not leave the attribute behind

Create Values Accepts Only Add
    [Tags]    negative
    ${code}=    Unique Code    addonly
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $code}, 'data_type': 'option', 'variant_creation_mode': 'instant', 'org_id': $INV_ORG_ID, 'values': {'remove': ['01M5N0ERKCX83C11H5SHSNEZ92']}} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_attribute.values_create_add_only

Patch Applies A Values Delta Atomically
    [Documentation]    PROD-001: add, update and remove in one PATCH, together with an attribute
    ...    column. ADD08: renaming and recolouring keeps the value id.
    ${attr}    ${values}=    Create Attribute With Values    never    red    green
    ${value_etag}=    Read Etag    ${ATTRIBUTE_VALUE_API}    ${values}[red]
    ${delta}=    Evaluate    {'add': [{'code': 'BLUE', 'name': {'en-US': 'Blue'}, 'color_hex': '#0000FF'}], 'update': [{'id': $values['red'], 'etag': $value_etag, 'name': {'en-US': 'Crimson'}, 'color_hex': '#DC143C'}], 'remove': [$values['green']]}
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'sequence': 7, 'values': $delta} }}
    Response Status Should Be    ${resp}    200
    ${after}=    Read Attribute Values    ${attr}
    Dictionary Should Contain Key    ${after}    blue
    Dictionary Should Not Contain Key    ${after}    green
    Should Be Equal    ${after}[red]    ${values}[red]
    ${red}=    Read Record    ${ATTRIBUTE_VALUE_API}    ${values}[red]    id,name,color_hex
    Should Be Equal    ${red}[name][en-US]    Crimson
    Should Be Equal    ${red}[color_hex]    \#DC143C
    ${item}=    Read Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    id,sequence
    Should Be Equal As Integers    ${item}[sequence]    7

Patch With A Bad Row Rolls Back The Attribute Change
    [Documentation]    AC02: a refused row leaves both the values and the attribute's own columns
    ...    untouched; the error names the row.
    [Tags]    negative
    ${attr}    ${values}=    Create Attribute With Values    never    red
    ${delta}=    Evaluate    {'add': [{'code': 'OK', 'name': {'en-US': 'Fine'}}, {'code': 'DUP', 'name': {'en-US': 'RED'}}]}
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'sequence': 9, 'values': $delta} }}
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.name_duplicate    values.add[1].name
    ${after}=    Read Attribute Values    ${attr}
    Length Should Be    ${after}    1
    ${item}=    Read Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    id,sequence
    Should Not Be Equal    ${item.get('sequence')}    ${9}

Value Names Are Unique After Normalization Including Archived Values
    [Documentation]    AC28 / PROD-002: trimmed, case-folded names clash with an archived value
    ...    too; the duplicate is refused, never merged or restored.
    [Tags]    negative
    ${attr}    ${values}=    Create Attribute With Values    never    red
    ${resp}=    Set Record Archived    ${ATTRIBUTE_VALUE_API}    ${values}[red]
    Response Status Should Be    ${resp}    200
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $attr, 'code': 'RED2', 'name': {'en-US': ' RED '}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.name_duplicate    name

Colour Hex Must Be Six Hex Digits
    [Documentation]    PROD-003: `color_hex` is `#RRGGBB`.
    [Tags]    negative
    ${attr}    ${values}=    Create Attribute With Values    never    red
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $attr, 'code': 'BAD', 'name': {'en-US': 'Bad'}, 'color_hex': 'red', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    common:err_invalid_string_length    color_hex
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $attr, 'code': 'GOOD', 'name': {'en-US': 'Good'}, 'color_hex': '#a1B2c3', 'org_id': $INV_ORG_ID} }}
    Response Should Be Create Success    ${resp}

Multi Checkbox Requires Mode Never
    [Documentation]    PROD-004 / AC25 / ADD07: multi_checkbox is refused with instant; changing both
    ...    to never in one request is accepted while the attribute is unused.
    [Tags]    negative
    ${code}=    Unique Code    multi
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $code}, 'data_type': 'option', 'variant_creation_mode': 'instant', 'display_type': 'multi_checkbox', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_attribute.multi_checkbox_requires_never    display_type
    ${attr}    ${values}=    Create Attribute With Values    instant    a
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'display_type': 'multi_checkbox'} }}
    Response Should Contain Business Error    ${resp}    inventory:product_attribute.multi_checkbox_requires_never    display_type
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'display_type': 'multi_checkbox', 'variant_creation_mode': 'never'} }}
    Response Status Should Be    ${resp}    200

Display Type Is An Enum
    [Tags]    negative
    ${code}=    Unique Code    disp
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}
    ...    json=${{ {'code': $code, 'name': {'en-US': $code}, 'data_type': 'option', 'variant_creation_mode': 'never', 'display_type': 'carousel', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

Mode Changes Freely While Unused
    [Documentation]    ADD01: no template uses the attribute, so its mode may change; nothing is
    ...    generated.
    ${attr}    ${values}=    Create Attribute With Values    instant    a
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'variant_creation_mode': 'dynamic'} }}
    Response Status Should Be    ${resp}    200

Mode Is Locked While A Template Uses The Attribute
    [Documentation]    ADD02 / AC25 / PROD-005: an enabled template line locks the mode; the error
    ...    lists the template. ADD05: sending the same mode is accepted as a no-op.
    [Tags]    negative
    ${attr}    ${values}=    Create Attribute With Values    instant    a    b
    ${line}=    Template Line    ${attr}    ${values}[a]
    ${template_id}=    Create Configured Template    ${{ [$line] }}    words=Robot Mode Lock Template
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'variant_creation_mode': 'dynamic'} }}
    ${error}=    Response Should Contain Business Error    ${resp}    inventory:product_attribute.mode_locked
    ...    variant_creation_mode
    Should Contain    ${error}[vars][template_ids]    ${template_id}
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'variant_creation_mode': 'instant'} }}
    Response Status Should Be    ${resp}    200
    ${item}=    Read Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}
    ...    id,current_template_usage_count,can_change_variant_creation
    Should Be Equal As Integers    ${item}[current_template_usage_count]    1
    Should Be Equal    ${item}[can_change_variant_creation]    ${False}
    Set Suite Variable    ${USED_ATTRIBUTE_ID}    ${attr}
    Set Suite Variable    ${USED_VALUES}    ${values}
    Set Suite Variable    ${USED_TEMPLATE_ID}    ${template_id}

Mode Stays Locked When Only An Archived Template Uses It
    [Documentation]    ADD03: lines of an archived template still lock the mode.
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_TEMPLATE_API}    ${USED_TEMPLATE_ID}
    Response Status Should Be    ${resp}    200
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${USED_ATTRIBUTE_ID}    ${{ {'variant_creation_mode': 'never'} }}
    Response Should Contain Business Error    ${resp}    inventory:product_attribute.mode_locked
    ...    variant_creation_mode
    ${resp}=    Set Record Archived    ${PRODUCT_TEMPLATE_API}    ${USED_TEMPLATE_ID}    ${False}
    Response Status Should Be    ${resp}    200

Archiving A Used Attribute Or Selected Value Fails
    [Documentation]    AC21 / ADD10: archive is refused while a template applies the attribute or
    ...    selects the value, and the error lists the template; the template is not changed.
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_ATTRIBUTE_API}    ${USED_ATTRIBUTE_ID}
    ${error}=    Response Should Contain Business Error    ${resp}    inventory:product_attribute.in_use
    Should Contain    ${error}[vars][template_ids]    ${USED_TEMPLATE_ID}
    ${resp}=    Set Record Archived    ${ATTRIBUTE_VALUE_API}    ${USED_VALUES}[a]
    ${error}=    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.in_use
    Should Contain    ${error}[vars][template_ids]    ${USED_TEMPLATE_ID}
    ${config}=    Read Template Configuration    ${USED_TEMPLATE_ID}
    Length Should Be    ${config}[attribute_lines]    1

An Unselected Value Of A Used Attribute Can Be Archived
    [Documentation]    ADD09 / AC03: value b was never selected by the template, so it is free to
    ...    archive; adding or archiving catalogue values never changes the template.
    ${resp}=    Set Record Archived    ${ATTRIBUTE_VALUE_API}    ${USED_VALUES}[b]
    Response Status Should Be    ${resp}    200
    ${config}=    Read Template Configuration    ${USED_TEMPLATE_ID}
    Should Be Equal    ${config}[attribute_lines][0][value_ids]    ${{ [$USED_VALUES['a']] }}

Deleting A Value Or Attribute With Template History Fails
    [Documentation]    AC22 / PROD-006: once the line is removed the value may be archived, but
    ...    neither it nor its attribute may be deleted: template history references them.
    [Tags]    negative
    Configure Template Successfully    ${USED_TEMPLATE_ID}    ${{ [] }}
    ${item}=    Read Record    ${PRODUCT_ATTRIBUTE_API}    ${USED_ATTRIBUTE_ID}
    ...    id,current_template_usage_count,historical_template_usage_count,can_change_variant_creation
    Should Be Equal As Integers    ${item}[current_template_usage_count]    0
    Should Be Equal As Integers    ${item}[historical_template_usage_count]    1
    Should Be Equal    ${item}[can_change_variant_creation]    ${True}
    ${resp}=    DELETE On Session    api    ${ATTRIBUTE_VALUE_API}/${USED_VALUES}[a]
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.has_history    id
    ${resp}=    DELETE On Session    api    ${PRODUCT_ATTRIBUTE_API}/${USED_ATTRIBUTE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_attribute.has_template_history    id
    ${resp}=    Set Record Archived    ${ATTRIBUTE_VALUE_API}    ${USED_VALUES}[a]
    Response Status Should Be    ${resp}    200

Mode Changes Again Once No Template Uses It
    [Documentation]    ADD04 / AC25 adjusted: the lock follows current usage, not history.
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${USED_ATTRIBUTE_ID}    ${{ {'variant_creation_mode': 'dynamic'} }}
    Response Status Should Be    ${resp}    200

An Archived Attribute Is Read Only Until Unarchived
    [Tags]    negative
    ${resp}=    Set Record Archived    ${PRODUCT_ATTRIBUTE_API}    ${USED_ATTRIBUTE_ID}
    Response Status Should Be    ${resp}    200
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${USED_ATTRIBUTE_ID}    ${{ {'sequence': 3} }}
    Response Should Contain Business Error    ${resp}    inventory:product_attribute.archived_read_only
    ${resp}=    Set Record Archived    ${PRODUCT_ATTRIBUTE_API}    ${USED_ATTRIBUTE_ID}    ${False}
    Response Status Should Be    ${resp}    200

A Value Name Must Not Be Blank In Every Language
    [Documentation]    PROD-001: a name whose every locale is empty after trimming is refused, in
    ...    the nested delta and on the value resource alike.
    [Tags]    negative
    ${attr}    ${values}=    Create Attribute With Values    never    red
    ${delta}=    Evaluate    {'add': [{'code': 'BLANK', 'name': {'en-US': ''}}]}
    ${resp}=    Patch Record    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${{ {'values': $delta} }}
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.name_blank    values.add[0].name
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $attr, 'code': 'BLANK', 'name': {'en-US': ' ' * 2, 'vi-VN': ''}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.name_blank    name
    ${resp}=    Patch Record    ${ATTRIBUTE_VALUE_API}    ${values}[red]    ${{ {'name': {'en-US': ' '}} }}
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.name_blank    name
    ${after}=    Read Attribute Values    ${attr}
    Length Should Be    ${after}    1

A Value Of An Archived Attribute Cannot Be Created Or Updated
    [Documentation]    An archived attribute is read only, its values included.
    [Tags]    negative
    ${attr}    ${values}=    Create Attribute With Values    never    red
    ${resp}=    Set Record Archived    ${PRODUCT_ATTRIBUTE_API}    ${attr}
    Response Status Should Be    ${resp}    200
    ${resp}=    POST On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    json=${{ {'attribute_id': $attr, 'code': 'LATE', 'name': {'en-US': 'Late'}, 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.parent_archived    attribute_id
    ${resp}=    Patch Record    ${ATTRIBUTE_VALUE_API}    ${values}[red]    ${{ {'name': {'en-US': 'Crimson'}} }}
    Response Should Contain Business Error    ${resp}    inventory:product_attribute_value.parent_archived    attribute_id
    ${resp}=    Set Record Archived    ${PRODUCT_ATTRIBUTE_API}    ${attr}    ${False}
    Response Status Should Be    ${resp}    200
    ${resp}=    Patch Record    ${ATTRIBUTE_VALUE_API}    ${values}[red]    ${{ {'name': {'en-US': 'Crimson'}} }}
    Response Status Should Be    ${resp}    200
