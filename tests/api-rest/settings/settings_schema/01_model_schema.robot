*** Settings ***
Documentation     The model schema endpoint every dynamic resource exposes. A settings_schema
...               row whose engine failed to register 404s here, which reads to a caller as a
...               missing record rather than a missing route.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Test Tags         settings    settings_schema    meta


*** Test Cases ***
Settings Schema Model Schema Is Served
    ${resp}=    GET On Session    api    ${SETTINGS_SCHEMA_API}/meta/schema
    Response Should Match Schema    ${resp}    ${COMMON_SCHEMA_DIR}/model_meta_schema.json    200

Settings Schemas Are Not Tenant Scoped
    [Documentation]    A module's declaration is identical for every tenant, so settings_schemas
    ...    deliberately carries no tenant_id. If it ever gains one, boot-time registration runs
    ...    with no tenant in context and the application panics before serving a request — which
    ...    is exactly the failure this assertion exists to catch early.
    ${resp}=    GET On Session    api    ${SETTINGS_SCHEMA_API}/meta/schema
    Response Status Should Be    ${resp}    200
    ${body}=    Set Variable    ${resp.json()}
    ${names}=    Evaluate    list($body.get('fields', {}).keys())
    Should Not Contain    ${names}    tenant_id
    ...    msg=settings_schemas must not be tenant-scoped; see the boot panic this guards against.
