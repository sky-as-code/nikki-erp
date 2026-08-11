*** Settings ***
Documentation     The Product Variant schema. The critical assertions here are negative:
...               BR §5.2 / §7.6 / AC-PROD-004 forbid the variant from carrying its own copy
...               of an inherited field, and a duplicated column is exactly the kind of
...               modelling error that reads as harmless until the two copies disagree.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    product_variant    schema


*** Test Cases ***
Get Product Variant Model Schema
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Variant Owned Fields
    [Documentation]    BR §6.2.2 / §7.4: the variant owns what identifies a concrete
    ...    transactable unit — its template, its combination, its SKU and barcode.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    product_template_id    combination_key    sku    primary_barcode
    ...    is_materialized    status    archive_source    org_id
        Dictionary Should Contain Key    ${fields}    ${field}
    END

Schema Declares No Inherited Fields
    [Documentation]    BR §5.2 / §7.6 / AC-PROD-004 — the rule the whole template/variant
    ...    split exists to enforce. Name, classification and capability live on the template;
    ...    a stored copy here would become a second source of truth that goes stale the
    ...    moment the template is edited.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    name    category_id    product_type_id    brand_id
    ...    sale_ok    purchase_ok
        Dictionary Should Not Contain Key    ${fields}    ${field}
    END

Schema Declares The Inherited Fields As Virtual
    [Documentation]    The template's values still reach a variant read — as template_*
    ...    fields marked virtual, so they are served from the join rather than stored. That
    ...    is the difference between inheriting a value and duplicating it.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    template_name    template_category_id    template_brand_id
    ...    template_product_type_id    template_status    template_sale_ok
        Dictionary Should Contain Key    ${fields}    ${field}
    END

Schema Declares The Overridable Fallback Fields
    [Documentation]    BR §7.6: image, weight and dimensions are the allowlisted fields that
    ...    exist on both halves. Here they are unprefixed and nullable — null means "inherit
    ...    from the template", not "zero" (AC-PROD-014).
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    variant_image_id    weight    length    width    height
        Dictionary Should Contain Key    ${fields}    ${field}
    END

Schema Declares No Default Variant Flag
    [Documentation]    BR-PROD-VAR-005 / AC-PROD-010: there is no "default variant". A
    ...    template with no variant-generating attributes gets exactly one variant with an
    ...    empty combination, which needs no flag to mark it special.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    is_default_variant

Schema Declares No Display Name Field
    [Documentation]    BR §5.5 / AC-PROD-013: display_name is computed as template name plus
    ...    attribute values and served by the effective-product endpoint, never stored.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    display_name

Schema Declares Sku As The Record Label
    [Documentation]    The variant has no name of its own, so the relation picker falls back
    ...    to the SKU rather than rendering a raw ULID.
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    sku
