*** Settings ***
Documentation     IAM Service resource suite: iam_service is a view over iam_users (rows of type
...               service) with its own CRUD, routes and permission resource. File order: CRUD ->
...               type isolation -> principal guards. Teardown removes every tracked fixture.
Resource          ../../resources/iam.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Permission Fixtures
