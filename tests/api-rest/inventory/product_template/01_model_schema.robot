*** Settings ***
Documentation     The Product Template schema. Beyond the fields it declares, this suite
...               pins the fields it must NOT declare: BR §5.2 / §7.6 / AC-PROD-004 make
...               inheritance the template's job, and a duplicated field on the variant
...               would silently become a second source of truth.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Test Tags         inventory    product_template    schema


*** Test Cases ***
Get Product Template Model Schema
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Schema Declares The Catalog Fields
    [Documentation]    BR §6.1.2 / §7.3: the template owns the catalog-level definition —
    ...    its identity, its classification and its capability flags.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    name    short_name    product_type_id    category_id    brand_id
    ...    sale_ok    purchase_ok    status    org_id
        Dictionary Should Contain Key    ${fields}    ${field}
    END

Schema Declares The Fallback Fields
    [Documentation]    BR §7.6: image, weight and dimensions are the allowlisted fields that
    ...    exist on both halves. The template's copy is the fallback a variant inherits when
    ...    it does not override, which is why they are named default_*.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    default_image_id    default_weight    default_length
    ...    default_width    default_height
        Dictionary Should Contain Key    ${fields}    ${field}
    END

Schema Declares No Variant Owned Fields
    [Documentation]    BR §6.2.2: SKU, barcode and the combination key identify a concrete
    ...    transactable unit. A template with several variants has nothing for them to mean,
    ...    so their presence here would be a modelling error rather than a spare field.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${fields}=    Set Variable    ${resp.json()}[fields]
    FOR    ${field}    IN    sku    primary_barcode    combination_key    archive_source
        Dictionary Should Not Contain Key    ${fields}    ${field}
    END

Schema Declares No Display Name Field
    [Documentation]    BR §5.5 / AC-PROD-013: display_name is computed as template name plus
    ...    attribute values, never stored. A stored copy would go stale the moment the
    ...    template is renamed.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Dictionary Should Not Contain Key    ${resp.json()}[fields]    display_name

Schema Declares A Record Label Field
    [Documentation]    Without record_label_field the frontend relation picker falls back to
    ...    the raw ULID, so a template shows as an opaque id when chosen from a variant.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    Should Be Equal    ${resp.json()}[record_label_field]    name
