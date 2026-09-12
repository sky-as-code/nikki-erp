*** Settings ***
Documentation     Notification Inbox suite: what a signed-in person calls about their own
...               notifications.
...
...               These endpoints are not dynamic-model CRUD, which is why they are a suite
...               of their own rather than more files under notification/. None of them is
...               addressed by a record id, and none of them accepts a user id: whose
...               notifications they answer is decided by the server from the request
...               context (BR 18, BR 29). A test that could ask for somebody else's inbox
...               would be testing a parameter that must never exist.
...
...               The stream is covered here too, but only at its contract edges — the
...               headers and the refusals. Reading the body is a long-lived connection and
...               belongs in the micro-app's own tests, which do exactly that against a
...               stubbed response.
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
