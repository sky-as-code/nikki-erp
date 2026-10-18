*** Settings ***
Documentation     Updating Product Categories. The success cases run first (they consume and
...               rotate the saved etag); negatives follow, including the two shapes of the
...               acyclic-tree rule (BR §6.4.3).
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Product Category Under Test
Test Tags         inventory    product_category    update


*** Test Cases ***
Update Succeeds
    ${name}=    Unique Display Name    Robot Updated Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': {'en-US': $name}, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_CATEGORY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${etag}

Update Parent Succeeds
    [Documentation]    Updating a child without touching its parent or name is ordinary tree
    ...    maintenance and leaves its canonical name alone.
    Ensure Child Product Category
    ${before}=    Get Product Category    ${CHILD_CATEGORY_ID}
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${CHILD_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'sequence': 1, 'etag': $CHILD_CATEGORY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${CHILD_CATEGORY_ETAG}
    IF    $etag is not None    Set Global Variable    ${CHILD_CATEGORY_ETAG}    ${etag}
    ${after}=    Get Product Category    ${CHILD_CATEGORY_ID}
    Should Be Equal    ${after}[canonical_name]    ${before}[canonical_name]

Rename Cascades To Descendants
    [Documentation]    Renaming a category rewrites the canonical name of its child and
    ...    grandchild in the same transaction.
    Ensure Child Product Category
    ${grand_code}=    Unique Code    grandcat
    ${grand_name}=    Unique Display Name    Robot Grandchild Category
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'code': $grand_code, 'name': {'en-US': $grand_name}, 'parent_category_id': $CHILD_CATEGORY_ID, 'org_id': $INV_ORG_ID} }}
    ${grand_id}    ${grand_etag}=    Response Should Be Create Success    ${resp}
    ${new_name}=    Unique Display Name    Robot Renamed Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': {'en-US': $new_name}, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_CATEGORY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${etag}
    ${parent}=    Get Product Category    ${PRODUCT_CATEGORY_ID}
    ${child}=    Get Product Category    ${CHILD_CATEGORY_ID}
    ${grand}=    Get Product Category    ${grand_id}
    Should End With    ${parent}[canonical_name][en-US]    / ${new_name}
    Should Be Equal    ${child}[canonical_name][en-US]    ${parent}[canonical_name][en-US] / ${child}[name][en-US]
    Should Be Equal    ${grand}[canonical_name][en-US]    ${child}[canonical_name][en-US] / ${grand_name}
    Set Global Variable    ${CHILD_CATEGORY_ETAG}    ${child}[etag]
    [Teardown]    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${grand_id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Move Cascades To Descendants
    [Documentation]    Moving a category to another parent rewrites its own and its
    ...    descendants' canonical names; moving it back restores them.
    Ensure Child Product Category
    Ensure Root Product Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${CHILD_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_category_id': $ROOT_CATEGORY_ID, 'etag': $CHILD_CATEGORY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${CHILD_CATEGORY_ETAG}
    ${child}=    Get Product Category    ${CHILD_CATEGORY_ID}
    Should Be Equal    ${child}[canonical_name][en-US]    ${ROOT_CATEGORY_CANONICAL}[en-US] / ${child}[name][en-US]
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${CHILD_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_category_id': $PRODUCT_CATEGORY_ID, 'etag': $child['etag']} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${child}[etag]
    IF    $etag is not None    Set Global Variable    ${CHILD_CATEGORY_ETAG}    ${etag}
    ${parent}=    Get Product Category    ${PRODUCT_CATEGORY_ID}
    ${child}=    Get Product Category    ${CHILD_CATEGORY_ID}
    Should Be Equal    ${child}[canonical_name][en-US]    ${parent}[canonical_name][en-US] / ${child}[name][en-US]

Update With Canonical Name Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'canonical_name': {'en-US': 'Fake'}, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Response Should Be Category Server Managed Error    ${resp}    canonical_name

Update With Null Parent Fails
    [Documentation]    Every category has a parent; clearing it is refused, not left to the
    ...    NOT NULL column.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_category_id': None, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400

Rename Root Fails
    [Tags]    negative
    Ensure Root Product Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${ROOT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': {'en-US': 'Everything'}, 'etag': $ROOT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Response Should Be Category Root Immutable Error    ${resp}

Move Root Fails
    [Tags]    negative
    Ensure Root Product Category
    Ensure Product Category Under Test
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${ROOT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_category_id': $PRODUCT_CATEGORY_ID, 'etag': $ROOT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Response Should Be Category Root Immutable Error    ${resp}

Archive Root Fails
    [Tags]    negative
    Ensure Root Product Category
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}/${ROOT_CATEGORY_ID}/archived
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $ROOT_CATEGORY_ETAG, 'is_archived': True} }}
    ...    expected_status=any
    Response Should Be Category Root Immutable Error    ${resp}

Delete Root Fails
    [Tags]    negative
    Ensure Root Product Category
    ${resp}=    DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${ROOT_CATEGORY_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Category Root Immutable Error    ${resp}

Update With Self As Parent Fails
    [Documentation]    BR §6.4.3: a category cannot be its own parent — the degenerate
    ...    one-node cycle, reported separately from the walk because there is no chain.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_category_id': $PRODUCT_CATEGORY_ID, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Response Should Be Category Self Parent Error    ${resp}

Update With Descendant As Parent Fails
    [Documentation]    BR §6.4.3: re-parenting the root under its own child would close a
    ...    loop no upward walk could terminate on.
    [Tags]    negative
    Ensure Child Product Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_category_id': $CHILD_CATEGORY_ID, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Response Should Be Category Cycle Error    ${resp}

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/not-invalid-1234567890123
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $PRODUCT_CATEGORY_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${NOT_FOUND_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'etag': $PRODUCT_CATEGORY_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
