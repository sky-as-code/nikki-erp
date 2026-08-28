*** Settings ***
Documentation     The Vendor Product Price schema.
...
...               Two things are pinned here that a reader would otherwise have to infer, and both
...               are load-bearing.
...
...               product_variant_id is NULLABLE, and that is the specificity mechanism rather than
...               laxity: a row with no variant prices the whole template, and a row naming one
...               beats it. Making the column required would silently remove the ability to quote a
...               product line as a whole.
...
...               There is NO status field. A vendor price is master data; is_archived is how a
...               withdrawn offer is retired, and adding a status would create a second, competing
...               notion of "not in use" (PRICE-INV-022).
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Test Tags         purchase    vendor_product_price    schema


*** Test Cases ***
Get Vendor Product Price Model Schema
    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Core Fields
    [Documentation]    The four that make a price a price — the amount, its unit, its currency and
    ...    the quantity it applies from — plus who is offering it and for what.
    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${name}    IN    vendor_id    product_template_id    product_variant_id
    ...    purchase_uom_id    currency_id    min_quantity    unit_price    valid_from    valid_to
    ...    lead_time_days    sequence    vendor_product_code    vendor_product_name    org_id
        Dictionary Should Contain Key    ${fields}    ${name}
    END

Schema Declares No Status Field
    [Documentation]    PRICE-INV-022. Archival is the only "not in use" this resource has, and a
    ...    status beside it would let a row be archived-but-active or active-but-expired with
    ...    nothing to say which wins.
    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    status

Schema Declares An Archived Field
    [Documentation]    Section 25: an archived quote stops pricing anything new but stays readable,
    ...    so an order that resolved through it still names something a reader can open.
    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Contain Key    ${resp.json()}[fields]    is_archived

Product Variant Is Not Required For Create
    [Documentation]    A template-wide quote is a first-class thing, not a degenerate case. If this
    ...    became required, every price would have to name a variant and the specificity ladder of
    ...    PRICE-INV-018 would have nothing to rank.
    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${field}=    Set Variable    ${resp.json()}[fields][product_variant_id]
    ${required}=    Get From Dictionary    ${field}    required_for_create    default=${False}
    Should Not Be True    ${required}

Money Fields Are Decimals Rather Than Floats
    [Documentation]    Money is never a float here. The scale of 6 matches
    ...    purchase_order_line.unit_price, which is what this resource feeds — a coarser scale here
    ...    would round a quote on its way onto the line it prices.
    ${resp}=    GET On Session    api    ${VENDOR_PRICE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    Should Be Equal    ${fields}[unit_price][data_type][type]    decimal
    Should Be Equal    ${fields}[min_quantity][data_type][type]    decimal
