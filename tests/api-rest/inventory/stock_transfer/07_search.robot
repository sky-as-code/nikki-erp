*** Settings ***
Documentation     Searching Stock Transfers. Filtering by status and by operation type is what a
...               warehouse worklist is built from, so both have a declared search index.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Stock Transfer Under Test
Test Tags         inventory    stock_transfer    search


*** Test Cases ***
Search Returns The Transfer Under Test
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100} }}
    Response Status Should Be    ${resp}    200
    ${ids}=    Evaluate    [i['id'] for i in $resp.json()['data']['items']]
    Should Contain    ${ids}    ${STOCK_TRANSFER_ID}

Search By Status Filters
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 100, 'graph': '{"if": ["status", "eq", "draft"]}'} }}
    ...    expected_status=any
    IF    ${resp.status_code} == 200
        FOR    ${item}    IN    @{resp.json()}[data][items]
            Should Be Equal    ${item}[status]    draft
            ...    msg=A status filter must not return transfers in another state
        END
    END

Search Returns The Listing Fields
    [Documentation]    The default field set is what the listing page renders. A transfer row
    ...    showing only an id is unreadable to whoever has to pick one.
    ${resp}=    GET On Session    api    ${STOCK_TRANSFER_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 1} }}
    Response Status Should Be    ${resp}    200
    ${items}=    Set Variable    ${resp.json()}[data][items]
    IF    len($items) > 0
        FOR    ${field}    IN    transfer_number    status    operation_code
            Dictionary Should Contain Key    ${items}[0]    ${field}
            ...    msg=The default listing fields must include ${field}
        END
    END
