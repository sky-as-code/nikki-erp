*** Settings ***
Documentation     The schema rows a feature module registers to declare what it can be
...               configured with. These are declarations, not values: they are the same for
...               every tenant and are written by the boot-time registration path, not by a
...               client. The suite therefore reads and searches; it does not create.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Force Tags        settings    settings_schema
