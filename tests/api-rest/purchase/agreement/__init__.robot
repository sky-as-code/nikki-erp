*** Settings ***
Documentation     Purchase Agreement suite: blanket orders and purchase templates, and the
...               orders drawn against them. AC-15 through AC-20.
Resource          resources/purchase.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Purchase Seed Data
