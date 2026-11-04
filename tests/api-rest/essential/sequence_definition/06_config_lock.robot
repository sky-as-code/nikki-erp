*** Settings ***
Documentation     Configuration lock after the first issuance (SEQ-CFG-01, AC25, AC45). NextValue
...               has no REST route, so the number is issued the way production issues it: by
...               creating a stock transfer, numbered from its org's inventory.transfer definition.
Resource          resources/sequence.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Issued Sequence Definition
Test Tags         essential    sequence_definition    config_lock


*** Test Cases ***
First Issuance Stamps First Issued At
    ${item}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    Should Not Be Empty    ${item}[first_issued_at]
    Should Be Equal    ${item}[owner_module]    inventory
    Should Be Equal    ${item}[format_owner]    consumer

Changing The Padding After Issuance Is Refused
    [Documentation]    AC25: a different padding would change how every future code compares
    ...    with the ones already printed.
    [Tags]    negative
    ${item}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_ISSUED_DEFINITION_ID}
    ...    json=${{ {'padding': 7, 'etag': $item['etag']} }}    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_CONFIG_LOCKED}    padding

Changing The Increment, Templates Or Time Zone After Issuance Is Refused
    [Tags]    negative
    ${item}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_ISSUED_DEFINITION_ID}
    ...    json=${{ {'increment': 5, 'suffix_template': '-Z', 'timezone': 'Asia/Tokyo', 'etag': $item['etag']} }}
    ...    expected_status=any
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_CONFIG_LOCKED}    increment
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_CONFIG_LOCKED}    suffix_template
    Response Should Be Sequence Error    ${resp}    ${SEQ_ERR_CONFIG_LOCKED}    timezone
    ${after}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    Should Be Equal As Integers    ${after}[increment]    ${item}[increment]
    Should Be Equal    ${after}[timezone]    ${item}[timezone]

Resending The Stored Configuration Is Not A Change
    [Documentation]    A client may send the whole form back; an unchanged value is no violation.
    ${item}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_ISSUED_DEFINITION_ID}
    ...    json=${{ {'padding': $item['padding'], 'timezone': $item['timezone'], 'description': 'Robot note', 'etag': $item['etag']} }}
    Response Should Be Update Success    ${resp}    count=1    previous_etag=${item}[etag]

Renaming After Issuance Succeeds And Leaves Issued Codes Alone
    [Documentation]    AC45: name and description stay editable; the code already handed out does
    ...    not change, and the next transfer continues the same counter.
    ${item}=    Read Sequence Definition    ${SEQ_ISSUED_DEFINITION_ID}
    ${name}=    Unique Display Name    Robot Renamed Sequence
    ${resp}=    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_ISSUED_DEFINITION_ID}
    ...    json=${{ {'name': {'en-US': $name}, 'etag': $item['etag']} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1    previous_etag=${item}[etag]
    ${transfer}=    GET On Session    api    ${STOCK_TRANSFER_API}/${SEQ_ISSUED_TRANSFER_ID}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Should Be Equal    ${transfer.json()}[item][transfer_number]    ${SEQ_ISSUED_TRANSFER_NUMBER}
    [Teardown]    PATCH On Session    api    ${SEQ_DEFINITION_API}/${SEQ_ISSUED_DEFINITION_ID}
    ...    json=${{ {'name': $item['name'], 'etag': $etag} }}    expected_status=any
