*** Settings ***
Documentation     Template attribute configuration and variant reconcile (CR-INV-PRODUCT-VARIANTS +
...               ADDENDUM-02, PROD-008..PROD-022). The configuration is the FULL desired state sent
...               as `attribute_lines`/`exclusions` in the template's own PATCH, applied in the same
...               transaction as the template fields, guarded by the template etag and previewed by
...               `preview_attribute_configuration`. Every file builds its own templates and
...               attributes; they stay in the database (configured attributes keep history).
Resource          resources/inventory_configuration.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
