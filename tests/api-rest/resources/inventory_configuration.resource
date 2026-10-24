*** Settings ***
Documentation     Helpers for the product attribute / variant lifecycle suites (CR-INV-PRODUCT-VARIANTS
...               + ADDENDUM-02). A template's attribute configuration is written only through the
...               template's own create/update payload (`attribute_lines`, `exclusions`, the FULL
...               desired state in catalogue ids); the junction resources are read-only to clients.
...               These keywords build catalogue attributes with nested values, save configurations,
...               list a template's variants and resolve selections. Fixtures are left in the
...               database by design: a configured attribute keeps template history and can never
...               be hard-deleted, and every code/name is unique per execution.
...               All keywords expect an authorized session with alias "api".
Library           Collections
Resource          ${CURDIR}/inventory.resource


*** Variables ***
${TEMPLATE_EXCLUSION_API}           /v1/inventory/inventory_product_template_combination_exclusion
${VARIANT_ATTRIBUTE_VALUE_API}      /v1/inventory/inventory_product_variant_attribute_value


*** Keywords ***
# --- Catalogue ------------------------------------------------------------------------

Create Attribute With Values
    [Documentation]    Creates an `option` attribute in the given mode together with its values in
    ...    ONE request (the nested `values.add` delta), and returns (attribute id, {code: value id}).
    ...    Value codes are scoped to the attribute, so short fixed codes are safe.
    [Arguments]    ${mode}    @{codes}    ${words}=Robot Config Attribute    ${display_type}=${None}
    Ensure Inventory Org
    ${name}=    Unique Display Name    ${words}
    ${code}=    Unique Code    cfg
    ${rows}=    Evaluate    [{'code': c, 'name': {'en-US': c.upper()}, 'sequence': i} for i, c in enumerate($codes)]
    ${body}=    Create Dictionary    code=${code}    name=${{ {'en-US': $name} }}    data_type=option
    ...    variant_creation_mode=${mode}    org_id=${INV_ORG_ID}    values=${{ {'add': $rows} }}
    IF    $display_type is not None    Set To Dictionary    ${body}    display_type=${display_type}
    ${resp}=    POST On Session    api    ${PRODUCT_ATTRIBUTE_API}    json=${body}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${values}=    Read Attribute Values    ${id}
    RETURN    ${id}    ${values}

Read Attribute Values
    [Documentation]    Returns {code: id} of every value of an attribute, archived ones included.
    [Arguments]    ${attribute_id}
    ${resp}=    GET On Session    api    ${ATTRIBUTE_VALUE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 200, 'include_archived': 'true', 'fields': 'id,code', 'graph': '{"if":["attribute_id","=","' + $attribute_id + '"]}'} }}
    Response Status Should Be    ${resp}    200
    ${values}=    Evaluate    {item['code']: item['id'] for item in $resp.json()['items']}
    RETURN    ${values}

Read Record
    [Documentation]    GET by id with the given projection, returning the item.
    [Arguments]    ${api}    ${id}    ${fields}=${None}
    ${params}=    Create Dictionary    org_id=${INV_ORG_ID}
    IF    $fields is not None    Set To Dictionary    ${params}    fields=${fields}
    ${resp}=    GET On Session    api    ${api}/${id}    params=${params}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item]

Read Etag
    [Arguments]    ${api}    ${id}
    ${item}=    Read Record    ${api}    ${id}    etag
    RETURN    ${item}[etag]

Patch Record
    [Documentation]    PATCH with the record's current etag; returns the raw response, so a caller
    ...    can assert either outcome.
    [Arguments]    ${api}    ${id}    ${body}
    ${etag}=    Read Etag    ${api}    ${id}
    ${payload}=    Copy Dictionary    ${body}
    Set To Dictionary    ${payload}    org_id=${INV_ORG_ID}    etag=${etag}
    ${resp}=    PATCH On Session    api    ${api}/${id}    json=${payload}    expected_status=any
    RETURN    ${resp}

Set Record Archived
    [Documentation]    Archives or unarchives with the current etag; returns the raw response.
    [Arguments]    ${api}    ${id}    ${archived}=${True}
    ${etag}=    Read Etag    ${api}    ${id}
    ${resp}=    POST On Session    api    ${api}/${id}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $etag, 'is_archived': $archived} }}    expected_status=any
    RETURN    ${resp}

# --- Template configuration -----------------------------------------------------------

Template Line
    [Documentation]    One desired attribute line: a catalogue attribute and its selected values.
    [Arguments]    ${attribute_id}    @{value_ids}
    ${line}=    Create Dictionary    attribute_id=${attribute_id}    value_ids=${value_ids}
    RETURN    ${line}

Exclusion
    [Arguments]    ${left_value_id}    ${right_value_id}
    ${pair}=    Create Dictionary    left_value_id=${left_value_id}    right_value_id=${right_value_id}
    RETURN    ${pair}

Configure Template
    [Documentation]    Saves the FULL desired configuration through the template's PATCH and returns
    ...    the raw response. ${exclusions}=None leaves the exclusions unchanged.
    [Arguments]    ${template_id}    ${lines}    ${exclusions}=${None}    &{extra}
    ${body}=    Create Dictionary    attribute_lines=${lines}    &{extra}
    IF    $exclusions is not None    Set To Dictionary    ${body}    exclusions=${exclusions}
    ${resp}=    Patch Record    ${PRODUCT_TEMPLATE_API}    ${template_id}    ${body}
    RETURN    ${resp}

Configure Template Successfully
    [Arguments]    ${template_id}    ${lines}    ${exclusions}=${None}
    ${resp}=    Configure Template    ${template_id}    ${lines}    ${exclusions}
    Response Status Should Be    ${resp}    200

Preview Template Configuration
    [Documentation]    Dry run of the same payload; returns the raw response.
    [Arguments]    ${template_id}    ${lines}    ${exclusions}=${None}
    ${etag}=    Read Etag    ${PRODUCT_TEMPLATE_API}    ${template_id}
    ${body}=    Create Dictionary    org_id=${INV_ORG_ID}    etag=${etag}    attribute_lines=${lines}
    IF    $exclusions is not None    Set To Dictionary    ${body}    exclusions=${exclusions}
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}/preview_attribute_configuration
    ...    json=${body}    expected_status=any
    RETURN    ${resp}

Read Template Configuration
    [Documentation]    The computed `attribute_configuration` read: {attribute_lines, exclusions}.
    [Arguments]    ${template_id}
    ${item}=    Read Record    ${PRODUCT_TEMPLATE_API}    ${template_id}    attribute_configuration
    RETURN    ${item}[attribute_configuration]

Count Template Variants
    [Documentation]    {base, complete, possible}: the unarchived variants of each kind, counted from
    ...    the variant search, and how many completes the current configuration allows, read from a
    ...    preview that re-sends that configuration unchanged (null when over the threshold).
    [Arguments]    ${template_id}
    ${variants}=    Search Template Variants    ${template_id}
    ${bases}=    Variants Of Kind    ${variants}    base
    ${completes}=    Variants Of Kind    ${variants}    complete
    ${config}=    Read Template Configuration    ${template_id}
    ${lines}=    Evaluate
    ...    [{'attribute_id': l['attribute_id'], 'value_ids': l['value_ids']} for l in $config['attribute_lines']]
    ${resp}=    Preview Template Configuration    ${template_id}    ${lines}    ${config}[exclusions]
    Response Status Should Be    ${resp}    200
    ${counts}=    Create Dictionary    base=${{ len($bases) }}    complete=${{ len($completes) }}
    ...    possible=${resp.json()}[possible_complete_count]
    RETURN    ${counts}

Search Template Variants
    [Documentation]    Every variant of a template. Archived ones are included only on request,
    ...    because the engine hides them by default.
    [Arguments]    ${template_id}    ${include_archived}=${False}
    ${params}=    Create Dictionary    org_id=${INV_ORG_ID}    size=500
    ...    fields=id,etag,combination_key,variant_kind,base_variant_id,is_archived,archive_source,sku
    ...    graph={"if":["product_template_id","=","${template_id}"]}
    IF    ${include_archived}    Set To Dictionary    ${params}    include_archived=true
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}    params=${params}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[items]

Variants Of Kind
    [Arguments]    ${variants}    ${kind}
    ${picked}=    Evaluate    list(filter(lambda v, k=$kind: v.get('variant_kind') == k, $variants))
    RETURN    ${picked}

Variant By Id
    [Arguments]    ${variants}    ${variant_id}
    ${picked}=    Evaluate    list(filter(lambda v, i=$variant_id: v['id'] == i, $variants))
    Length Should Be    ${picked}    1    msg=Variant ${variant_id} not found in ${variants}
    RETURN    ${picked}[0]

Combination Key
    [Documentation]    The key the server builds: "attr:value" pairs sorted by attribute id, "|"-joined.
    [Arguments]    &{pairs}
    ${key}=    Evaluate    '|'.join(a + ':' + v for a, v in sorted($pairs.items()))
    RETURN    ${key}

Create Configured Template
    [Documentation]    A fresh template configured in a second step (create gives the empty default
    ...    first). Returns the template id.
    [Arguments]    ${lines}    ${exclusions}=${None}    ${words}=Robot Configured Template
    ${id}    ${etag}=    Create Product Template    ${words}
    Configure Template Successfully    ${id}    ${lines}    ${exclusions}
    RETURN    ${id}

# --- Resolve --------------------------------------------------------------------------

Selection
    [Arguments]    ${attribute_id}    ${value_id}
    ${pair}=    Create Dictionary    attribute_id=${attribute_id}    value_id=${value_id}
    RETURN    ${pair}

Resolve Selection
    [Documentation]    POST resolve_selection; returns the raw response.
    [Arguments]    ${template_id}    ${selections}    ${materialize}=${True}    ${expected_etag}=${None}
    ${body}=    Create Dictionary    org_id=${INV_ORG_ID}    product_template_id=${template_id}
    ...    selections=${selections}    materialize_if_missing=${materialize}
    IF    $expected_etag is not None    Set To Dictionary    ${body}    expected_etag=${expected_etag}
    ${resp}=    POST On Session    api    ${PRODUCT_TEMPLATE_API}/resolve_selection
    ...    json=${body}    expected_status=any
    RETURN    ${resp}

# --- Usage ----------------------------------------------------------------------------

Post Stock Move For Variant
    [Documentation]    Opens a draft internal transfer and adds one move of the variant; returns
    ...    (transfer id, raw move response). A move row is stock history: the reconcile treats a
    ...    variant it references as used, so it is archived rather than deleted.
    [Arguments]    ${variant_id}
    Ensure Internal Operation Type Under Test
    ${transfer_id}    ${etag}=    Create Stock Transfer    ${INTERNAL_OPERATION_TYPE_ID}
    ${resp}=    POST On Session    api    ${STOCK_MOVE_API}
    ...    json=${{ {'transfer_id': $transfer_id, 'product_variant_id': $variant_id, 'demand_quantity': '1', 'base_demand_quantity': '1', 'source_location_id': $INVENTORY_LOCATION_ID, 'destination_location_id': $STOCK_DEST_LOCATION_ID, 'status': 'draft', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    RETURN    ${transfer_id}    ${resp}

Use Variant In Stock Move
    [Documentation]    Makes a variant "used" through a draft stock move; returns (transfer id, move id).
    [Arguments]    ${variant_id}
    ${transfer_id}    ${resp}=    Post Stock Move For Variant    ${variant_id}
    ${move_id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${transfer_id}    ${move_id}

Release Stock Move
    [Documentation]    Best-effort removal of a move and its transfer, so the variant is unused again.
    [Arguments]    ${transfer_id}    ${move_id}
    DELETE On Session    api    ${STOCK_MOVE_API}/${move_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    DELETE On Session    api    ${STOCK_TRANSFER_API}/${transfer_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

# --- Fixtures -------------------------------------------------------------------------

Ensure Two By Two Template
    [Documentation]    A template configured with two INSTANT attributes of two values each — the
    ...    2x2 fixture BR §8.2 / AC-PROD-011 describes. Saving the configuration reconciles at
    ...    once, so the four variants exist before any generate call. Built fresh rather than
    ...    shared, because counts are only meaningful against a template nothing else touched.
    ${id}=    Get Variable Value    ${MATRIX_TEMPLATE_ID}    ${EMPTY}
    IF    $id    RETURN
    ${attr_a}    ${values_a}=    Create Attribute With Values    instant    a0    a1    words=Robot Matrix Attribute A
    ${attr_b}    ${values_b}=    Create Attribute With Values    instant    b0    b1    words=Robot Matrix Attribute B
    ${line_a}=    Template Line    ${attr_a}    ${values_a}[a0]    ${values_a}[a1]
    ${line_b}=    Template Line    ${attr_b}    ${values_b}[b0]    ${values_b}[b1]
    ${id}=    Create Configured Template    ${{ [$line_a, $line_b] }}    words=Robot Matrix Template
    Set Global Variable    ${MATRIX_TEMPLATE_ID}    ${id}
    Set Global Variable    ${MATRIX_ATTRIBUTE_A_ID}    ${attr_a}
    Set Global Variable    ${MATRIX_ATTRIBUTE_B_ID}    ${attr_b}
