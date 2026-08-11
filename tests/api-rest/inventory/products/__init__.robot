*** Settings ***
Documentation     Cross-resource Inventory Products suites — the capabilities that span the
...               template/variant pair rather than belonging to either alone: variant
...               generation (BR §8.2), selection resolution (BR §14.4), and the permission
...               posture of the module's entitlement seeds. They run after the per-resource
...               suites because they assume the CRUD contract already holds.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
