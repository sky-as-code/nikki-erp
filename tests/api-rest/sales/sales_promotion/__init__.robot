*** Settings ***
Documentation     Sales Promotion & Voucher (BR_Sales_Promotion_Voucher). Each suite builds the
...               promotions it needs through the API and archives them in its teardown, so the
...               suites run repeatedly against the same database and never depend on seeds.
Resource          resources/sales_promotion.resource
Suite Setup       Run Keywords    Create Authorized API Session    AND    Ensure Promotion Inputs
