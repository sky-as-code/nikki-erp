*** Settings ***
Documentation     Re-parenting and renaming a location, and what follows from it.
...
...               canonical_name, warehouse_id and hierarchy_depth are derived from the tree, so a
...               move, a cleared parent or a rename has to rewrite them for the location and
...               everything beneath, in one transaction. The cycle rules are what stop a branch
...               being grafted onto itself, which would detach it from the tree entirely while
...               leaving it pointing in a circle.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Move Fixture
Test Tags         inventory    inventory_location    move


*** Test Cases ***
A Root Location Has Its Own Name As Its Canonical Name
    ${item}=    Get Location    ${MOVE_ROOT_ID}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_ROOT_NAME}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    0
    Should Be Equal    ${item.get('warehouse_id')}    ${None}    msg=A location without a parent has no warehouse

A Child Canonical Name Is Its Parent's And Its Own Name
    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_ROOT_NAME} / ${MOVE_BRANCH_NAME}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    1

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_ROOT_NAME} / ${MOVE_BRANCH_NAME} / ${MOVE_LEAF_NAME}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    2

The Canonical Name Cannot Be Set By A Client
    [Tags]    negative
    ${name}=    Unique Display Name    Robot Canonical
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'name': $name, 'canonical_name': 'Forged / Path', 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ...    expected_status=any
    Should Be True    ${resp.status_code} >= 400
    Should Contain    ${resp.text}    inventory_location.canonical_name_server_managed

The Warehouse Is Not The Client's To Choose
    [Documentation]    warehouse_id is auto-generated: it follows the parent, so a value sent on
    ...    create is ignored.
    Ensure Warehouse Under Test
    ${name}=    Unique Display Name    Robot No Warehouse
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}
    ...    json=${{ {'name': $name, 'warehouse_id': $WAREHOUSE_ID, 'location_usage': 'internal', 'org_id': $INV_ORG_ID} }}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    ${item}=    Get Location    ${id}
    Should Be Equal    ${item.get('warehouse_id')}    ${None}
    DELETE On Session    api    ${INVENTORY_LOCATION_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any

Renaming A Location Rewrites Its Subtree
    [Documentation]    The leaf's canonical name carries its ancestors' names, so renaming the
    ...    branch reaches it too.
    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    ${renamed}=    Unique Display Name    Robot renamed
    ${resp}=    PATCH On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'name': $renamed, 'etag': $item['etag']} }}
    Response Should Be Update Success    ${resp}
    Set Global Variable    ${MOVE_BRANCH_NAME}    ${renamed}

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_ROOT_NAME} / ${renamed} / ${MOVE_LEAF_NAME}

Moving A Branch Rewrites The Whole Subtree
    [Documentation]    The leaf is two levels down and nobody named it in the request, but its
    ...    canonical name describes where it is — so it has to change when its ancestor moves.
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_location_id': $MOVE_SECOND_ROOT_ID} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_SECOND_ROOT_NAME} / ${MOVE_BRANCH_NAME}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    1

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item}[canonical_name]
    ...    ${MOVE_SECOND_ROOT_NAME} / ${MOVE_BRANCH_NAME} / ${MOVE_LEAF_NAME}
    ...    msg=A descendant must follow its ancestor

Moving Under A Warehouse Root Puts The Subtree In The Warehouse
    [Documentation]    INV-010. warehouse_id is copied from the parent and cascades to the
    ...    whole subtree.
    Ensure Warehouse Under Test
    ${root}=    Get Warehouse Root Location Id    ${WAREHOUSE_ID}
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_location_id': $root} }}
    Response Status Should Be    ${resp}    200

    ${root_item}=    Get Location    ${root}
    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    Should Be Equal    ${item.get('warehouse_id')}    ${WAREHOUSE_ID}
    Should Be Equal    ${item}[canonical_name]    ${root_item}[canonical_name] / ${MOVE_BRANCH_NAME}
    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item.get('warehouse_id')}    ${WAREHOUSE_ID}    msg=The subtree follows its ancestor

Clearing The Parent Takes The Subtree Out Of The Warehouse
    [Documentation]    INV-010. With no parent a location belongs to no warehouse, and its
    ...    children follow. A plain update does this, not only the move action.
    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    ${resp}=    PATCH On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_location_id': None, 'etag': $item['etag']} }}
    Response Should Be Update Success    ${resp}

    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    Should Be Equal    ${item.get('warehouse_id')}    ${None}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_BRANCH_NAME}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    0

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item.get('warehouse_id')}    ${None}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_BRANCH_NAME} / ${MOVE_LEAF_NAME}

Moving To No Parent Makes A Location A Root
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_LEAF_ID}/move
    ...    json=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item}[canonical_name]    ${MOVE_LEAF_NAME}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    0
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_LEAF_ID}/move
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_location_id': $MOVE_BRANCH_ID} }}
    Response Status Should Be    ${resp}    200

A Location Cannot Become Its Own Parent
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_location_id': $MOVE_BRANCH_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

A Location Cannot Move Under Its Own Descendant
    [Documentation]    Grafting a branch onto its own leaf would cut the pair off from the tree
    ...    and leave them pointing at each other.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_location_id': $MOVE_LEAF_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Moving Under An Unknown Parent Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'org_id': $INV_ORG_ID, 'parent_location_id': $NOT_FOUND_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400


*** Keywords ***
Ensure Move Fixture
    [Documentation]    Two roots and a two-level branch under the first, so a move has
    ...    somewhere to come from, somewhere to go, and a descendant that must follow.
    ${id}=    Get Variable Value    ${MOVE_ROOT_ID}    ${EMPTY}
    IF    $id    RETURN

    ${root}    ${root_name}=    Create Move Location    moveroot    ${EMPTY}
    Set Global Variable    ${MOVE_ROOT_ID}    ${root}
    Set Global Variable    ${MOVE_ROOT_NAME}    ${root_name}

    ${second}    ${second_name}=    Create Move Location    movealt    ${EMPTY}
    Set Global Variable    ${MOVE_SECOND_ROOT_ID}    ${second}
    Set Global Variable    ${MOVE_SECOND_ROOT_NAME}    ${second_name}

    ${branch}    ${branch_name}=    Create Move Location    movebranch    ${root}
    Set Global Variable    ${MOVE_BRANCH_ID}    ${branch}
    Set Global Variable    ${MOVE_BRANCH_NAME}    ${branch_name}

    ${leaf}    ${leaf_name}=    Create Move Location    moveleaf    ${branch}
    Set Global Variable    ${MOVE_LEAF_ID}    ${leaf}
    Set Global Variable    ${MOVE_LEAF_NAME}    ${leaf_name}

Create Move Location
    [Arguments]    ${prefix}    ${parent_id}
    ${name}=    Unique Display Name    Robot ${prefix}
    ${payload}=    Create Dictionary    name=${name}    location_usage=internal    org_id=${INV_ORG_ID}
    IF    $parent_id    Set To Dictionary    ${payload}    parent_location_id    ${parent_id}
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}    json=${payload}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${name}

Get Location
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${id}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    RETURN    ${item}
