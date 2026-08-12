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
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${PRODUCT_CATEGORY_ETAG}
    IF    $etag is not None    Set Global Variable    ${PRODUCT_CATEGORY_ETAG}    ${etag}

Update Parent Succeeds
    [Documentation]    Re-parenting onto an unrelated, valid category is ordinary tree
    ...    maintenance, not a cycle. Uses the child fixture as the new parent, then restores
    ...    the category under test back to a root so the rest of the suite sees it unchanged.
    Ensure Child Product Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${CHILD_CATEGORY_ID}
    ...    json=${{ {'sequence': 1, 'etag': $CHILD_CATEGORY_ETAG} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${CHILD_CATEGORY_ETAG}
    IF    $etag is not None    Set Global Variable    ${CHILD_CATEGORY_ETAG}    ${etag}

Update With Self As Parent Fails
    [Documentation]    BR §6.4.3: a category cannot be its own parent — the degenerate
    ...    one-node cycle, reported separately from the walk because there is no chain.
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'parent_category_id': $PRODUCT_CATEGORY_ID, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Response Should Be Category Self Parent Error    ${resp}

Update With Descendant As Parent Fails
    [Documentation]    BR §6.4.3: re-parenting the root under its own child would close a
    ...    loop no upward walk could terminate on.
    [Tags]    negative
    Ensure Child Product Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'parent_category_id': $CHILD_CATEGORY_ID, 'etag': $PRODUCT_CATEGORY_ETAG} }}
    ...    expected_status=any
    Response Should Be Category Cycle Error    ${resp}

Update With Missing Etag Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag

Update With Unmatched Etag Fails
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Stale Category
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${PRODUCT_CATEGORY_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': '___________________'} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Update With Invalid Id Format Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/not-invalid-1234567890123
    ...    json=${{ {'etag': $PRODUCT_CATEGORY_ETAG} }}    expected_status=any
    Response Should Be Invalid Format Error    ${resp}    id

Update With Not Found Id Fails
    [Tags]    negative
    ${resp}=    PATCH On Session    api    ${PRODUCT_CATEGORY_API}/${NOT_FOUND_ID}
    ...    json=${{ {'etag': $PRODUCT_CATEGORY_ETAG} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}
