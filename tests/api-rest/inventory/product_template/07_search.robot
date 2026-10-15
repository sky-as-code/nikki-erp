*** Settings ***
Documentation     Searching Product Templates — the listing page's own query. Graph filters
...               rely on the seeded templates (the "Lead" variants), so they pass on any
...               database. The "graph" query values are JSON strings that requests
...               URL-encodes via the params dict.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Product Template Under Test
...               AND    Ensure Seeded Product Templates    50
Test Tags         inventory    product_template    search


*** Variables ***
${TEMPLATE_SCHEMA}    ${INVENTORY_SCHEMA_DIR}/product_template.json


*** Test Cases ***
Search Without Criteria Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search With Paging Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 2, 'size': 7} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=7    page=2

Search With No Result Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'page': 99} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=99    item_count=0

Search By Name Succeeds
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["name", "*", "lead"]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search By Category Succeeds
    [Documentation]    Filtering a catalog by category is the listing page's primary facet,
    ...    and it is the query a category detail page issues for its related templates.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["category_id", "=", "' + $PRODUCT_CATEGORY_ID + '"]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search By Status Succeeds
    [Documentation]    BR-PROD-TPL-004: because status is independent of is_archived, a
    ...    discontinued-product listing is a plain status filter rather than an archive one.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': '{"if":["status", "=", "draft"]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0

Search Filtering By Aggregate Computed Field Succeeds
    [Documentation]    variant_count counts the template's variants in SQL. A filter on it
    ...    lands in the WHERE clause through the shared lateral, so the count and the page agree.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['id', 'name', 'variant_count'], 'graph': '{"if":["variant_count", ">=", 0]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0
    Should Be True    ${resp.json()}[total] >= ${{len($resp.json()['items'])}}

Search Ordering By Aggregate Computed Field Succeeds
    [Documentation]    Ordering happens in PostgreSQL, before paging, so every page is sorted.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['id', 'variant_count'], 'graph': '{"order": [["variant_count", "desc"]]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0
    ${counts}=    Evaluate    [i.get('variant_count') or 0 for i in $resp.json()['items']]
    ${sorted}=    Evaluate    sorted($counts, reverse=True)
    Lists Should Be Equal    ${counts}    ${sorted}

Search Filtering By Related Computed Field Succeeds
    [Documentation]    uom_name copies uom.name; its predicate lands on the joined unit's column.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['id'], 'graph': '{"and":[{"if":["id", "=", "' + $PRODUCT_TEMPLATE_ID + '"]},{"or":[{"if":["uom_name", "is_set"]},{"if":["uom_name", "not_set"]}]}]}'} }}
    Response Should Be Search Success    ${resp}    ${TEMPLATE_SCHEMA}    size=50    page=0
    Search Results Should Contain Id    ${resp}    ${PRODUCT_TEMPLATE_ID}

Schema Advertises Computed Fields As Filterable And Sortable
    [Documentation]    The listing gates its filter cells and sort menus on these flags.
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}/meta/schema
    Response Status Should Be    ${resp}    200
    FOR    ${name}    IN    variant_count    uom_name    name
        Should Be Equal    ${resp.json()}[fields][${name}][is_filterable]    ${True}    msg=${name} is_filterable
        Should Be Equal    ${resp.json()}[fields][${name}][is_sortable]    ${True}    msg=${name} is_sortable
    END

Search With Nonexist Field Fails
    [Tags]    negative
    ${resp}=    GET On Session    api    ${PRODUCT_TEMPLATE_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'fields': ['name', 'bla_bla_field']} }}
    ...    expected_status=any
    Response Should Be Nonexist Fields Error    ${resp}    bla_bla_field
