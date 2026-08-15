*** Settings ***
Documentation     Re-parenting a location, and the cached path that follows from it.
...
...               complete_path is a cache of what the tree already says, so a move has to
...               rewrite it for the location and everything beneath. The cycle rules are what
...               stop a branch being grafted onto itself, which would detach it from the tree
...               entirely while leaving it pointing in a circle.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Inventory Org    AND    Ensure Move Fixture
Test Tags         inventory    inventory_location    move


*** Test Cases ***
A Root Location Has Its Own Code As Its Path
    ${item}=    Get Location    ${MOVE_ROOT_ID}
    Should Be Equal    ${item}[complete_path]    ${MOVE_ROOT_CODE}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    0

A Child Path Is Its Parent Path And Its Own Code
    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    Should Be Equal    ${item}[complete_path]    ${MOVE_ROOT_CODE}/${MOVE_BRANCH_CODE}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    1

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item}[complete_path]    ${MOVE_ROOT_CODE}/${MOVE_BRANCH_CODE}/${MOVE_LEAF_CODE}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    2

Moving A Branch Rewrites The Whole Subtree
    [Documentation]    The leaf is two levels down and nobody named it in the request, but its
    ...    path describes where it is — so it has to change when its ancestor moves.
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'parent_location_id': $MOVE_SECOND_ROOT_ID} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    Should Be Equal    ${item}[complete_path]    ${MOVE_SECOND_ROOT_CODE}/${MOVE_BRANCH_CODE}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    1

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item}[complete_path]
    ...    ${MOVE_SECOND_ROOT_CODE}/${MOVE_BRANCH_CODE}/${MOVE_LEAF_CODE}
    ...    msg=A descendant path must follow its ancestor

Moving To No Parent Makes A Location A Root
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {} }}
    Response Status Should Be    ${resp}    200

    ${item}=    Get Location    ${MOVE_BRANCH_ID}
    Should Be Equal    ${item}[complete_path]    ${MOVE_BRANCH_CODE}
    Should Be Equal As Integers    ${item}[hierarchy_depth]    0

    ${item}=    Get Location    ${MOVE_LEAF_ID}
    Should Be Equal    ${item}[complete_path]    ${MOVE_BRANCH_CODE}/${MOVE_LEAF_CODE}

A Location Cannot Become Its Own Parent
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'parent_location_id': $MOVE_BRANCH_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

A Location Cannot Move Under Its Own Descendant
    [Documentation]    Grafting a branch onto its own leaf would cut the pair off from the tree
    ...    and leave them pointing at each other.
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'parent_location_id': $MOVE_LEAF_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400

Moving Under An Unknown Parent Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}/${MOVE_BRANCH_ID}/move
    ...    json=${{ {'parent_location_id': $NOT_FOUND_ID} }}    expected_status=any
    Should Be True    ${resp.status_code} >= 400


*** Keywords ***
Ensure Move Fixture
    [Documentation]    Two roots and a two-level branch under the first, so a move has
    ...    somewhere to come from, somewhere to go, and a descendant whose path must follow.
    ${id}=    Get Variable Value    ${MOVE_ROOT_ID}    ${EMPTY}
    IF    $id    RETURN

    ${root}    ${root_code}=    Create Move Location    moveroot    ${EMPTY}
    Set Global Variable    ${MOVE_ROOT_ID}    ${root}
    Set Global Variable    ${MOVE_ROOT_CODE}    ${root_code}

    ${second}    ${second_code}=    Create Move Location    movealt    ${EMPTY}
    Set Global Variable    ${MOVE_SECOND_ROOT_ID}    ${second}
    Set Global Variable    ${MOVE_SECOND_ROOT_CODE}    ${second_code}

    ${branch}    ${branch_code}=    Create Move Location    movebranch    ${root}
    Set Global Variable    ${MOVE_BRANCH_ID}    ${branch}
    Set Global Variable    ${MOVE_BRANCH_CODE}    ${branch_code}

    ${leaf}    ${leaf_code}=    Create Move Location    moveleaf    ${branch}
    Set Global Variable    ${MOVE_LEAF_ID}    ${leaf}
    Set Global Variable    ${MOVE_LEAF_CODE}    ${leaf_code}

Create Move Location
    [Arguments]    ${prefix}    ${parent_id}
    ${name}=    Unique Display Name    Robot ${prefix}
    ${code}=    Unique Code    ${prefix}
    ${payload}=    Create Dictionary    code=${code}    location_usage=internal    org_id=${INV_ORG_ID}
    Set To Dictionary    ${payload}    name    ${{ {'en-US': $name} }}
    IF    $parent_id    Set To Dictionary    ${payload}    parent_location_id    ${parent_id}
    ${resp}=    POST On Session    api    ${INVENTORY_LOCATION_API}    json=${payload}
    ${id}    ${etag}=    Response Should Be Create Success    ${resp}
    RETURN    ${id}    ${code}

Get Location
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${INVENTORY_LOCATION_API}/${id}
    ${item}=    Item Should Match Schema    ${resp}    ${INVENTORY_SCHEMA_DIR}/inventory_location.json    200
    RETURN    ${item}
