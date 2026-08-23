*** Settings ***
Documentation     The stored setting values, one row per setting per owner. The row grain is
...               what makes last-write-wins safe: two actors editing different settings touch
...               different rows and cannot collide at all.
Resource          resources/settings.resource
Suite Setup       Create Authorized API Session
Force Tags        settings    settings_record
