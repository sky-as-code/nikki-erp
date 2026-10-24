*** Settings ***
Documentation     Deleting the Product Template under test — always the LAST suite, doubling
...               as cleanup. The delete guard of BR-PROD-TPL-005 / AC-PROD-021 is the point:
...               a template whose variants may already be referenced by a transaction must
...               be archived rather than deleted, and the error must say so.
Resource          resources/inventory_configuration.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Template Under Test
Test Tags         inventory    product_template    delete


*** Test Cases ***
Delete With A Used Variant Fails
    [Documentation]    BR-PROD-TPL-005 / AC-PROD-021 / D7: a template whose variant is used (here a
    ...    stock move references it) is archived, never hard-deleted. The engine answers with a
    ...    field-level business error rather than letting the FK surface as a 500.
    [Tags]    negative
    Ensure Product Variant Under Test
    ${transfer_id}    ${move_id}=    Use Variant In Stock Move    ${PRODUCT_VARIANT_ID}
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Template Has Variants Error    ${resp}
    [Teardown]    Release Stock Move    ${transfer_id}    ${move_id}

Delete Removes Unused Variants With The Template
    [Documentation]    D7: unused variants no longer block a delete; they go in the same
    ...    transaction. Runs on its own template so the shared one stays for the tests below.
    ${template_id}    ${template_etag}=    Create Product Template    Robot Delete Unused Template
    ${key}=    Unique Code    delcomb
    ${variant_id}    ${variant_etag}=    Create Product Variant    ${template_id}    ${key}
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${template_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1
    ${resp}=    GET On Session    api    ${PRODUCT_VARIANT_API}/${variant_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete Succeeds
    [Documentation]    With its variants gone the template deletes normally — the guard is
    ...    about surviving references, not about templates being undeletable.
    Delete Inventory Fixture    ${PRODUCT_VARIANT_API}    PRODUCT_VARIANT_ID
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Should Be Delete Success    ${resp}    count=1

Delete Again With Same Id Fails
    [Documentation]    Idempotency check on the just-deleted id; clears the globals so the
    ...    template under test is not reused after this point.
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${PRODUCT_TEMPLATE_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
    Set Global Variable    ${PRODUCT_TEMPLATE_ID}    ${EMPTY}
    Set Global Variable    ${PRODUCT_TEMPLATE_ETAG}    ${EMPTY}

Delete With Not Found Id Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/${NOT_FOUND_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Delete With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    DELETE On Session    api    ${PRODUCT_TEMPLATE_API}/not-existing-1234567890123
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id
