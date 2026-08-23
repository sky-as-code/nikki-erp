*** Settings ***
Documentation     Who may edit what. Tenant and organization configuration decides how the
...               product behaves for everybody in that scope, so it must not be world-writable
...               — but every user has to be able to read and change their own preferences, so
...               a blanket denial is equally wrong. This suite pins both halves.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Permission Fixtures
Force Tags        settings    permissions
