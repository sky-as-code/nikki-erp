*** Settings ***
Documentation     Essential Currency suite: the Odoo-style rounding that derives decimal_places, and
...               the per-organization activation (is_active) every currency request answers for.
Resource          resources/essential.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Org Under Test
Suite Teardown    Delete Currency Fixtures
