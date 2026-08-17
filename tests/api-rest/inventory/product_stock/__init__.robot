*** Settings ***
Documentation     Product ↔ Stock integration. Product becomes a place to see stock and reach
...               the operations that change it, while owning none of it: no quantity is stored
...               against a product, and every number here is read from Stock on request.
...
...               File order (NN_ prefixes) runs the reads before the guards, because the guards
...               archive things the reads would otherwise still be looking at.
Resource          resources/inventory.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Inventory Seed Data
