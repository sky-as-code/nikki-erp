*** Settings ***
Documentation     Notification resource suite.
...
...               Four files rather than the usual eight, and that is the contract rather
...               than an omission: a notification is raised by the module that has
...               something to say, through SendNotification or the command bus. Neither
...               has an HTTP route (BR 29), and 1008002_notification_iam.sql seeds no
...               create, update, delete or set_archived action. A create, update, delete
...               or archive suite would be testing endpoints that by design do not exist.
...
...               The suite therefore reads notifications it did not create. Where the
...               database has none, tests skip rather than invent a fixture: skipping says
...               the contract was not exercised, where a hand-written row would imply it
...               passed.
Resource          resources/notification.resource
Suite Setup       Create Authorized API Session
