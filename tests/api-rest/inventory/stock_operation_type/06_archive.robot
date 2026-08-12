*** Settings ***
Documentation     Archiving the Stock Operation Type under test, rotating the saved etag.
...               The type is unarchived again so the later suites see it live.
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session
...               AND    Ensure Stock Operation Type Under Test
Test Tags         inventory    stock_operation_type    archive


*** Test Cases ***
Archive Succeeds
    [Documentation]    STOCK-INV-015: the operation type is the one Stock resource that
    ...    supports archive, and AC-STOCK-029 requires it to carry is_archived for it.
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}/archived
    ...    json=${{ {'etag': $STOCK_OPERATION_TYPE_ETAG, 'is_archived': True} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1
    ...    previous_etag=${STOCK_OPERATION_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${etag}

Archived Operation Type Is Still Readable
    [Documentation]    AC-STOCK-031: a historical transfer still resolves the type it was
    ...    created with, so archiving must not make the record unreadable.
    ${resp}=    GET On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}
    ${item}=    Item Should Match Schema    ${resp}
    ...    ${INVENTORY_SCHEMA_DIR}/stock_operation_type.json    200
    Should Be Equal    ${item}[is_archived]    ${True}

Unarchive Succeeds
    [Documentation]    BR §4.2.1.6: restoring makes the type usable for new transfers again.
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}/archived
    ...    json=${{ {'etag': $STOCK_OPERATION_TYPE_ETAG, 'is_archived': False} }}
    ${etag}=    Response Should Be Update Success    ${resp}    count=1
    ...    previous_etag=${STOCK_OPERATION_TYPE_ETAG}
    IF    $etag is not None    Set Global Variable    ${STOCK_OPERATION_TYPE_ETAG}    ${etag}

Archive With Not Found Id Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/${NOT_FOUND_ID}/archived
    ...    json=${{ {'etag': $STOCK_OPERATION_TYPE_ETAG, 'is_archived': True} }}    expected_status=any
    Response Should Be Not Found Error    ${resp}

Archive With Unmatched Etag Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}/archived
    ...    json=${{ {'etag': '___________________', 'is_archived': True} }}    expected_status=any
    Response Should Be Etag Unmatched Error    ${resp}

Archive With Missing Required Fields Fails
    [Tags]    negative
    ${resp}=    POST On Session    api    ${STOCK_OPERATION_TYPE_API}/${STOCK_OPERATION_TYPE_ID}/archived
    ...    json=${{ {} }}    expected_status=any
    Response Should Be Missing Fields Error    ${resp}    etag    is_archived
