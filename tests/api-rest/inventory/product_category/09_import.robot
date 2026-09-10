*** Settings ***
Documentation     Import and bulk create on Product Categories, the pilot resource of the
...               generic import feature (POST {resource}/import and POST {resource}/bulk
...               on the composable engine). Files are generated per run so every code and
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
    ...    ID,Mã,Tên,Danh mục cha
    ...    ${run}-1,imp1${run},Import One ${run},
    ...    ${run}-2,imp2${run},Import Two ${run},${parent}
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    2
    Should Be Equal As Integers    ${resp.json()}[updated_count]    0
    Should Be Equal As Integers    ${resp.json()}[error_count]    0
    Should Be Equal As Integers    ${resp.json()}[total_rows]    2
    ${child}=    Find Category By Code    imp2${run}
    Should Be Equal    ${child}[parent_category_id]    ${PRODUCT_CATEGORY_ID}
    ...    msg=The reference cell must resolve to the parent's id
    Should Be Equal    ${child}[external_id]    ${run}-2
    Should Be Equal    ${child}[source_system]    import
    [Teardown]    Delete Categories By Codes    imp1${run}    imp2${run}

Re-Import By External Id Updates Instead Of Duplicating
    [Documentation]    Deduplication: the same (source_system, external_id) pair is an update.
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_first_${run}.csv
    ...    ID,Mã,Tên,Danh mục cha
    ...    ${run}-1,upd${run},Before ${run},
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    ${csv}=    Write Csv    import_second_${run}.csv
    ...    ID,Mã,Tên,Danh mục cha
    ...    ${run}-1,upd${run},After ${run},
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    0
    Should Be Equal As Integers    ${resp.json()}[updated_count]    1
    ${item}=    Find Category By Code    upd${run}
    Should Be Equal    ${item}[name][en-US]    After ${run}
    [Teardown]    Delete Categories By Codes    upd${run}

Import Reports An Unknown Reference Per Row And Keeps The Others
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_ref_${run}.csv
    ...    ID,Mã,Tên,Danh mục cha
    ...    ${run}-1,refok${run},Ref Ok ${run},
    ...    ${run}-2,refbad${run},Ref Bad ${run},No Such Category ${run}
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[error_count]    1
    Should Be Equal As Integers    ${resp.json()}[errors][0][row]    2
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_reference_not_found
    Should Be Equal    ${resp.json()}[errors][0][field]    parent_category_id
    [Teardown]    Delete Categories By Codes    refok${run}

Import Creates A Missing Referenced Category When Asked
    [Documentation]    "Create referenced record if not existing" creates the target with its
    ...    label only. The category domain service derives the code from the name, so the
    ...    parent comes into being with code "missing_parent_{run}" and the row points at it.
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_create_${run}.csv
    ...    ID,Mã,Tên,Danh mục cha
    ...    ${run}-1,crt${run},Create Ref ${run},Missing Parent ${run}
    ${mapping}=    Evaluate    dict($DEFAULT_MAPPING, create_missing_references=True)
    ${resp}=    Import File    ${csv}    ${mapping}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[error_count]    0
    ${parent}=    Find Category By Code    missing_parent_${run}
    Should Be Equal    ${parent}[name][en-US]    Missing Parent ${run}
    ${child}=    Find Category By Code    crt${run}
    Should Be Equal    ${child}[parent_category_id]    ${parent}[id]
    [Teardown]    Delete Categories By Codes    crt${run}    missing_parent_${run}

Create Without Code Derives It From The Name
    [Documentation]    The same rule on the plain create path: a category posted with a name and
    ...    no code gets the codified name, and a second one with the same name gets a suffix.
    ${run}=    Unique Suffix
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'name': {'en-US': 'Đồ uống ' + $run}, 'org_id': $INV_ORG_ID} }}    expected_status=any
    ${first}    ${etag}=    Response Should Be Create Success    ${resp}
    ${item}=    Find Category By Code    do_uong_${run}
    Should Be Equal    ${item}[id]    ${first}
    ${resp}=    POST On Session    api    ${PRODUCT_CATEGORY_API}
    ...    json=${{ {'name': {'en-US': 'Đồ uống ' + $run}, 'org_id': $INV_ORG_ID} }}    expected_status=any
    ${second}    ${etag}=    Response Should Be Create Success    ${resp}
    ${item}=    Find Category By Code    do_uong_${run}_2
    Should Be Equal    ${item}[id]    ${second}
    [Teardown]    Delete Categories By Codes    do_uong_${run}    do_uong_${run}_2

Import Reports A Duplicate External Id Within The File
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_dup_${run}.csv
    ...    ID,Mã,Tên,Danh mục cha
    ...    ${run}-1,dupa${run},Dup A ${run},
    ...    ${run}-1,dupb${run},Dup B ${run},
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[errors][0][row]    2
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_duplicate_external_id
    [Teardown]    Delete Categories By Codes    dupa${run}

Import Reports An Unconvertible Cell Per Row
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_cell_${run}.csv
    ...    ID,Mã,Tên,Thứ tự
    ...    ${run}-1,cell${run},Cell ${run},not-a-number
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'ID', 'target': 'external_id'}, {'source': 'Mã', 'target': 'code'}, {'source': 'Tên', 'target': 'name'}, {'source': 'Thứ tự', 'target': 'sequence'}]}
    ${resp}=    Import File    ${csv}    ${mapping}
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    0
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_cell_invalid
    Should Be Equal    ${resp.json()}[errors][0][field]    sequence

Import Without A Mandatory Column Is Refused
    [Documentation]    external_id is always mandatory for import; code and name are required
    ...    for create. Leaving external_id unmapped is a 400 naming the field.
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_mandatory_${run}.csv
    ...    Mã,Tên
    ...    mand${run},Mandatory ${run}
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'Mã', 'target': 'code'}, {'source': 'Tên', 'target': 'name'}]}
    ${resp}=    Import File    ${csv}    ${mapping}
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    err_import_mandatory_unmapped
    Should Contain    ${resp.json()}[0][vars][fields]    external_id
    ${found}=    Search Categories By Code    mand${run}
    Should Be Empty    ${found}    msg=A refused import must write nothing

Import With An Unknown Target Field Is Refused
    [Tags]    negative
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_unknown_${run}.csv
    ...    ID,Mã,Tên
    ...    ${run}-1,unk${run},Unknown ${run}
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'ID', 'target': 'external_id'}, {'source': 'Mã', 'target': 'code'}, {'source': 'Tên', 'target': 'bla_bla_field'}]}
    ${resp}=    Import File    ${csv}    ${mapping}
    Should Be Equal As Integers    ${resp.status_code}    400
    ${keys}=    Evaluate    [e['key'] for e in $resp.json()]
    Should Contain    ${keys}    err_import_mapping_unknown_field

Import With An Unsupported Extension Is Refused
    [Tags]    negative
    ${run}=    Unique Suffix
    ${path}=    Write Csv    import_ext_${run}.xls
    ...    ID,Mã,Tên
    ...    ${run}-1,ext${run},Ext ${run}
    ${resp}=    Import File    ${path}    ${DEFAULT_MAPPING}    filename=import_ext_${run}.xls
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    err_file_type_not_allowed

Import Without A File Is Refused
    [Tags]    negative
    ${files}=    Evaluate    {'mapping': (None, json.dumps($DEFAULT_MAPPING))}
    ${resp}=    POST On Session    api    ${IMPORT_API}
    ...    files=${files}    params=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    err_import_file_required

Bulk Create Writes Items Then Updates Them By External Id
    ${run}=    Unique Suffix
    ${items}=    Evaluate    [{'code': 'blk1' + $run, 'name': {'en-US': 'Bulk One ' + $run}, 'external_id': $run + '-b1'}, {'code': 'blk2' + $run, 'name': {'en-US': 'Bulk Two ' + $run}, 'external_id': $run + '-b2'}]
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
    ${item}=    Find Category By Code    blk1${run}
    Should Be Equal    ${item}[name][en-US]    Bulk One Renamed ${run}
    Should Be Equal    ${item}[source_system]    manual
    [Teardown]    Delete Categories By Codes    blk1${run}    blk2${run}

Bulk Create Reports A Rejected Item And Writes The Rest
    [Tags]    negative
    ${run}=    Unique Suffix
    ${items}=    Evaluate    [{'code': 'bad' + $run, 'external_id': $run + '-bad'}, {'code': 'good' + $run, 'name': {'en-US': 'Good ' + $run}, 'external_id': $run + '-good'}]
    ${resp}=    POST On Session    api    ${BULK_API}
    ...    json=${{ {'org_id': $INV_ORG_ID, 'items': $items} }}    expected_status=any
    Response Status Should Be    ${resp}    200
    Should Be Equal As Integers    ${resp.json()}[created_count]    1
    Should Be Equal As Integers    ${resp.json()}[errors][0][row]    1
    Should Be Equal    ${resp.json()}[errors][0][code]    err_import_validation
    [Teardown]    Delete Categories By Codes    good${run}

Bulk Create Without Items Is Refused
    [Tags]    negative
    ${resp}=    POST On Session    api    ${BULK_API}
    ...    json=${{ {'org_id': $INV_ORG_ID} }}    expected_status=any
    Should Be Equal As Integers    ${resp.status_code}    400
    Should Be Equal    ${resp.json()}[0][key]    err_import_items_required

Plain User Is Refused Import
    [Documentation]    Import writes records, so it is gated like create.
    [Tags]    permission
    Create Plain User Session
    ${run}=    Unique Suffix
    ${csv}=    Write Csv    import_denied_${run}.csv
    ...    ID,Mã,Tên
    ...    ${run}-1,den${run},Denied ${run}
    ${resp}=    Import File    ${csv}    ${DEFAULT_MAPPING}    alias=plain_user
    Should Be Equal As Integers    ${resp.status_code}    403
    ...    msg=A user holding only the system User role must not import product categories


*** Keywords ***
Set Default Mapping
    [Documentation]    The mapping every happy-path test sends: the three mandatory targets plus
    ...    the parent reference, keyed by the Vietnamese headers the generated files carry.
    ${mapping}=    Evaluate    {'language_code': 'en-US', 'create_missing_references': False, 'columns': [{'source': 'ID', 'target': 'external_id'}, {'source': 'Mã', 'target': 'code'}, {'source': 'Tên', 'target': 'name'}, {'source': 'Danh mục cha', 'target': 'parent_category_id'}]}
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

Search Categories By Code
    [Documentation]    Lists the categories of the org whose code is one of the given codes.
    ...    The default page is wide enough for the handful of rows a test writes.
    [Arguments]    @{codes}
    ${wanted}=    Create List    @{codes}
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}
    ...    params=${{ {'org_id': $INV_ORG_ID, 'size': 200} }}
    Response Status Should Be    ${resp}    200
    # A comprehension is its own Python scope, so the list goes in through the namespace.
    ${scope}=    Create Dictionary    wanted=${wanted}    items=${resp.json()}[items]
    ${found}=    Evaluate    [i for i in items if i.get('code') in wanted]    namespace=${scope}
    RETURN    ${found}

Find Category By Code
    [Arguments]    ${code}
    ${found}=    Search Categories By Code    ${code}
    Length Should Be    ${found}    1    msg=Expected exactly one category with code ${code}
    ${resp}=    GET On Session    api    ${PRODUCT_CATEGORY_API}/${found}[0][id]
    ...    params=${{ {'org_id': $INV_ORG_ID} }}
    Response Status Should Be    ${resp}    200
    RETURN    ${resp.json()}[item]

Delete Categories By Codes
    [Arguments]    @{codes}
    ${found}=    Search Categories By Code    @{codes}
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
