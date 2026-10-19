*** Settings ***
Documentation     Import and bulk create on Product Categories, the pilot resource of the
...               generic import feature (POST {resource}/import and POST {resource}/bulk
...               on the composable engine). Files are generated per run so every name and
...               external id is unique; every row written here is deleted at the end of
...               the test that wrote it.
...
...               The plain-user case follows 03_permissions.robot: it needs an account
...               holding only the system User role and skips when none is configured.
Library           Collections
Library           OperatingSystem
Library           RequestsLibrary
Resource          resources/inventory.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Inventory Org
...               AND    Set Default Mapping
Test Tags         inventory    product_category    import


*** Variables ***
${IMPORT_API}     ${PRODUCT_CATEGORY_API}/import
${BULK_API}       ${PRODUCT_CATEGORY_API}/bulk


*** Test Cases ***
Import Csv With A Resolved Reference Creates Rows
    [Documentation]    Two rows, one naming the category under test as its parent by label.
    ...    The label resolves to the parent's id; both rows are created; nothing is reported.
    Ensure Product Category Under Test
    ${parent}=    Get Category Label    ${PRODUCT_CATEGORY_ID}
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_ok_${run}.csv
    ...    ID,Tên,Danh mục cha
    ...    ${run}-1,Import One ${run},
    ...    ${run}-2,Import Two ${run},${parent}
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    2
    Should Be Equal As Integers    ${resp.json()}[updated_count]    0
    Should Be Equal As Integers    ${resp.json()}[error_count]    0
    Should Be Equal As Integers    ${resp.json()}[total_rows]    2
    ${child}=    Find Category By Name    Import Two ${run}
    Should Be Equal    ${child}[parent_category_id]    ${PRODUCT_CATEGORY_ID}
    ...    msg=The reference cell must resolve to the parent's id
    Should Be Equal    ${child}[external_id]    ${run}-2
    Should Be Equal    ${child}[source_system]    import
    [Teardown]    Delete Categories By Names    Import One ${run}    Import Two ${run}

Re-Import By External Id Updates Instead Of Duplicating
    [Documentation]    Deduplication: the same (source_system, external_id) pair is an update.
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_first_${run}.csv
    ...    ID,Tên,Danh mục cha
    ...    ${run}-1,Before ${run},
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    ${csv}=    Write Csv    import_second_${run}.csv
    ...    ID,Tên,Danh mục cha
    ...    ${run}-1,After ${run},
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    0
    Should Be Equal As Integers    ${resp.json()}[updated_count]    1
    ${item}=    Find Category By Name    After ${run}
    Should Be Equal    ${item}[name][en-US]    After ${run}
    [Teardown]    Delete Categories By Names    Before ${run}    After ${run}

Import Reports An Unknown Reference Per Row And Keeps The Others
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_ref_${run}.csv
    ...    ID,Tên,Danh mục cha
    ...    ${run}-1,Ref Ok ${run},
    ...    ${run}-2,Ref Bad ${run},No Such Category ${run}
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[error_count]    1
    Should Be Equal As Integers    ${resp.json()}[errors][0][row]    2
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_reference_not_found
    Should Be Equal    ${resp.json()}[errors][0][field]    parent_category_id
    [Teardown]    Delete Categories By Names    Ref Ok ${run}

Import Creates A Missing Referenced Category When Asked
    [Documentation]    "Create referenced record if not existing" creates the target with its
    ...    label only, so the parent comes into being named "Missing Parent {run}" under the
    ...    org's root and the row points at it.
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_create_${run}.csv
    ...    ID,Tên,Danh mục cha
    ...    ${run}-1,Create Ref ${run},Missing Parent ${run}
    ${mapping}=    Evaluate    dict($DEFAULT_MAPPING, create_missing_references=True)
    ${resp}=    Import File    ${csv}    ${mapping}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[error_count]    0
    ${parent}=    Find Category By Name    Missing Parent ${run}
    Should Be Equal    ${parent}[name][en-US]    Missing Parent ${run}
    ${child}=    Find Category By Name    Create Ref ${run}
    Should Be Equal    ${child}[parent_category_id]    ${parent}[id]
    Ensure Root Product Category
    Should Be Equal    ${parent}[parent_category_id]    ${ROOT_CATEGORY_ID}
    ...    msg=A referenced category created by the import has no parent of its own, so it lands under Default
    [Teardown]    Delete Categories By Names    Create Ref ${run}    Missing Parent ${run}

Import Reports A Duplicate External Id Within The File
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_dup_${run}.csv
    ...    ID,Tên,Danh mục cha
    ...    ${run}-1,Dup A ${run},
    ...    ${run}-1,Dup B ${run},
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[errors][0][row]    2
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_duplicate_external_id
    [Teardown]    Delete Categories By Names    Dup A ${run}

Import Reports An Unconvertible Cell Per Row
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_cell_${run}.csv
    ...    ID,Tên,Thứ tự,Danh mục cha
    ...    ${run}-1,Cell ${run},not-a-number,
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'ID', 'target': 'external_id'}, {'source': 'Tên', 'target': 'name'}, {'source': 'Thứ tự', 'target': 'display_order'}, {'source': 'Danh mục cha', 'target': 'parent_category_id'}]}
    ${resp}=    Import File    ${csv}    ${mapping}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    0
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_cell_invalid
    Should Be Equal    ${resp.json()}[errors][0][field]    display_order

Import Without A Mandatory Column Is Refused
    [Documentation]    external_id is always mandatory for import; name is required
    ...    for create. Leaving external_id unmapped is a 400 naming the field.
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_mandatory_${run}.csv
    ...    Tên
    ...    Mandatory ${run}
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'Tên', 'target': 'name'}]}
    ${resp}=    Import File    ${csv}    ${mapping}
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    common:err_import_mandatory_unmapped
    Should Contain    ${resp.json()}[0][vars][fields]    external_id
    ${found}=    Search Categories By Name    Mandatory ${run}
    Should Be Empty    ${found}    msg=A refused import must write nothing

Import With An Unknown Target Field Is Refused
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_unknown_${run}.csv
    ...    ID,Tên
    ...    ${run}-1,Unknown ${run}
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'ID', 'target': 'external_id'}, {'source': 'Tên', 'target': 'bla_bla_field'}]}
    ${resp}=    Import File    ${csv}    ${mapping}
    Should Be Equal As Integers    ${resp.status_code}    400
    ${keys}=    Evaluate    [e['key'] for e in $resp.json()]
    Should Contain    ${keys}    common:err_import_mapping_unknown_field

Import With An Unsupported Extension Is Refused
    [Tags]    negative
    ${run}=    Unique Suffix
    ${path}=    Write Csv    import_ext_${run}.xls
    ...    ID,Tên
    ...    ${run}-1,Ext ${run}
    ${resp}=    Import File    ${path}    ${DEFAULT_MAPPING}    filename=import_ext_${run}.xls
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    common:err_file_type_not_allowed

Import Without A File Is Refused
    [Tags]    negative
    ${files}=    Evaluate    {'mapping': (None, json.dumps($DEFAULT_MAPPING))}
    ${resp}=    POST On Session    api    ${IMPORT_API}
    ...    files=${files}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    common:err_import_file_required

Bulk Create Writes Items Then Updates Them By External Id
    ${run}=    Unique Suffix
    ${items}=    Evaluate    [{'name': {'en-US': 'Bulk One ' + $run}, 'external_id': $run + '-b1'}, {'name': {'en-US': 'Bulk Two ' + $run}, 'external_id': $run + '-b2'}]
    ${resp}=    POST On Session    api    ${BULK_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'items': $items} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    2
    Should Be Equal As Integers    ${resp.json()}[error_count]    0
    ${renamed}=    Evaluate    [dict($items[0], name={'en-US': 'Bulk One Renamed ' + $run})]
    ${resp}=    POST On Session    api    ${BULK_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'items': $renamed} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[updated_count]    1
    Should Be Equal As Integers    ${resp.json()}[created_count]    0
    ${item}=    Find Category By Name    Bulk One Renamed ${run}
    Should Be Equal    ${item}[name][en-US]    Bulk One Renamed ${run}
    Should Be Equal    ${item}[source_system]    manual
    [Teardown]    Delete Categories By Names    Bulk One Renamed ${run}    Bulk Two ${run}

Bulk Create Reports A Rejected Item And Writes The Rest
    [Tags]    negative
    ${run}=    Unique Suffix
    ${items}=    Evaluate    [{'external_id': $run + '-bad'}, {'name': {'en-US': 'Good ' + $run}, 'external_id': $run + '-good'}]
    ${resp}=    POST On Session    api    ${BULK_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'items': $items} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[errors][0][row]    1
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_validation
    [Teardown]    Delete Categories By Names    Good ${run}

Bulk Create Without Items Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${BULK_API}
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    common:err_import_items_required

Plain User Is Refused Import
    [Documentation]    Import writes records, so it is gated like create.
    [Tags]    permission
    Create Plain User Session
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_denied_${run}.csv
    ...    ID,Tên
    ...    ${run}-1,Denied ${run}
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}    alias=plain_user
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not import product categories


*** Keywords ***
Set Default Mapping
    [Documentation]    The mapping every happy-path test sends: the two mandatory targets plus
    ...    the parent reference, keyed by the Vietnamese headers the generated files carry.
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'ID', 'target': 'external_id'}, {'source': 'Tên', 'target': 'name'}, {'source': 'Danh mục cha', 'target': 'parent_category_id'}]}
    Set Suite Variable    ${DEFAULT_MAPPING}    ${mapping}

Write Csv
    [Documentation]    Writes the given lines as a UTF-8 csv under the output directory and
    ...    returns its path. Generated per run so unique columns never collide.
    [Arguments]    ${name}    @{lines}
    ${path}=    Set Variable    ${OUTPUT DIR}/import_data/${name}
    ${content}=    Catenate    SEPARATOR=\n    @{lines}
    Create File    ${path}    ${content}\n    encoding=UTF-8
    RETURN    ${path}

Import File
    [Documentation]    POST {resource}/import as a browser would: a multipart form with the
    ...    "file" part and the "mapping" JSON part, org_id in the query string.
    [Arguments]    ${path}    ${mapping}    ${alias}=api    ${filename}=${EMPTY}
    ${name}=    Evaluate    $filename or os.path.basename($path)    modules=os
    ${files}=    Evaluate    {'file': ($name, open($path, 'rb'), 'application/octet-stream')}
    ${resp}=    POST On Session    ${alias}    ${IMPORT_API}
    ...    files=${files}    data=${{ {'mapping': json.dumps($mapping)} }}
    ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    RETURN    ${resp}

Get Category Label
    [Arguments]    ${id}
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${id}    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item][name][en-US]

Search Categories By Name
    [Documentation]    Lists the categories of the org whose en-US name is one of the given names,
    ...    filtering on the server: an org accumulates categories across runs, so a single
    ...    unfiltered page cannot be relied on to hold the rows a test wrote. Names carry the
    ...    run suffix, so a contains filter plus an exact match on the en-US text is unique.
    [Arguments]    @{names}
    ${found}=    Create List
    FOR    ${name}    IN    @{names}
        ${graph}=    Evaluate    json.dumps({"if": ["name", "*", $name]})    json
        ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}
        ...    params=${{ {'org_id': $INV_ORG_ID, 'graph': $graph} }}
        Response Status Should Be    ${resp}    200
        ${exact}=    Evaluate
        ...    (lambda items, target: [i for i in items if (i.get('name') or {}).get('en-US') == target])($resp.json()['items'], $name)
        ${found}=    Combine Lists    ${found}    ${exact}
    END
    RETURN    ${found}

Find Category By Name
    [Arguments]    ${name}
    ${found}=    Search Categories By Name    ${name}
    Length Should Be    ${found}    1    msg=Expected exactly one category named ${name}
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${found}[0][id]
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item]

Delete Categories By Names
    [Arguments]    @{names}
    ${found}=    Search Categories By Name    @{names}
    FOR    ${item}    IN    @{found}
        DELETE On Session    api    ${PRODUCT_CATEGORY_API}/${item}[id]
        ...    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    END

Create Plain User Session
    [Documentation]    Mirrors inventory/products/03_permissions.robot: signs in as the
    ...    plain-role account under its own alias, or skips when none is configured.
    ${username}=    Get Variable Value    ${PLAIN_USER_USERNAME}    ${EMPTY}
    ${password}=    Get Variable Value    ${PLAIN_USER_PASSWORD}    ${EMPTY}
    IF    not $username or not $password
        Skip    No plain-role account configured; set PLAIN_USER_USERNAME and PLAIN_USER_PASSWORD to run the import permission check
    END
    Create Anonymous API Session    alias=plain_user_signin
    ${resp}=    POST On Session    plain_user_signin    ${SIGNIN_API}/start
    ...    json=${{ {'username': $username} }}
    ${attempt_id}=    Set Variable    ${resp.json()}[attempt_id]
    ${resp}=    POST On Session    plain_user_signin    ${SIGNIN_API}/continue
    ...    json=${{ {'attempt_id': $attempt_id, 'passwords': {'password': $password}} }}
    Should Be True    ${resp.json()}[done]    msg=Plain-user sign-in flow did not complete (done != true)
    ${certs}=    Evaluate    ($CLIENT_CERT, $CLIENT_KEY)
    ${headers}=    Create Dictionary    Authorization=Bearer ${resp.json()}[data][access_token]
    Create Client Cert Session    plain_user    ${API_HOST}    headers=${headers}
    ...    client_certs=${certs}    verify=${SSL_VERIFY}    disable_warnings=${1}
