*** Settings ***
Documentation     Buy X Get Y on a draft: the gift waits for confirmation (BR-R12), its quantity is
...               the server's (AC-14, BR-R08), a reprice updates the one gift line instead of adding
...               another (AC-16, BR-R09) and declining removes it.
Resource          resources/sales_promotion.resource
Suite Teardown    Archive Suite Promotions
Test Tags         sales    sales_promotion    buy_x_get_y


*** Test Cases ***
The Gift Waits For Confirmation
    ${id}=    Create Promotion    buy_x_get_y
    Add Rule    ${id}    min_quantity=3    qualification_uom_id=${SALES_UOM_ID}
    ${reward}=    Add Gift Reward    ${id}    1
    Link Scope    ${id}
    Activate Promotion    ${id}
    Set Suite Variable    ${BXGY_ID}    ${id}
    Set Suite Variable    ${BXGY_REWARD}    ${reward}
    ${order}=    Create Promotion Order    quantity=7
    Set Suite Variable    ${BXGY_ORDER}    ${order}
    ${quote}=    Quote Of    ${order}
    Program Should Be    ${quote}    ${id}    selection_required

Seven Bought With Buy Three Get One Gives Two Gifts
    [Documentation]    AC-14: two sets; the seventh unit carries nothing over.
    ${quote}=    Quote Of    ${BXGY_ORDER}    select_promotion_reward    promotion_id=${BXGY_ID}    reward_id=${BXGY_REWARD}
    ${program}=    Program Should Be    ${quote}    ${BXGY_ID}    applied
    Should Be Equal As Numbers    ${program}[gift][quantity]    2
    Should Not Be Empty    ${program}[gift][order_line_id]
    Set Suite Variable    ${GIFT_LINE}    ${program}[gift][order_line_id]

Repricing Keeps One Gift Line
    [Documentation]    AC-16: the gift neither feeds the set count nor duplicates itself.
    ${quote}=    Quote Of    ${BXGY_ORDER}
    ${program}=    Program Should Be    ${quote}    ${BXGY_ID}    applied
    Should Be Equal As Numbers    ${program}[gift][quantity]    2
    Should Be Equal    ${program}[gift][order_line_id]    ${GIFT_LINE}

Declining Removes The Gift Line
    ${quote}=    Quote Of    ${BXGY_ORDER}    select_promotion_reward    promotion_id=${BXGY_ID}    decline=${True}
    Program Should Be    ${quote}    ${BXGY_ID}    declined
    ${resp}=    GET On Session    api    ${SALES_ORDER_API}_line/${GIFT_LINE}    params=${{ {'org_id': $SALES_ORG_ID} }}
    ...    expected_status=any
    Should Not Be Equal As Integers    ${resp.status_code}    200

Selecting Without A Reward Or A Decline Is Refused
    ${resp}=    Order Action    ${BXGY_ORDER}    select_promotion_reward    promotion_id=${BXGY_ID}
    Promotion Error Should Be    ${resp}    promotion.reward_selection_required
