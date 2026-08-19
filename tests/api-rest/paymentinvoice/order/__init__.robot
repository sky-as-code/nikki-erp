*** Settings ***
Documentation     Payment & Invoice order suite.
...
...               This one departs from the usual CREATE -> ... -> DELETE lifecycle, and the
...               reason is worth stating. An order is not authored through the engine's plain
...               create: `create_payment` mints one, and it talks to a payment gateway. Every
...               gateway defaults to ENABLED: false, so against a stock deployment that
...               action is expected to fail as a *client* error — which is itself the
...               contract worth pinning, and is all of it that belongs in CI. Driving real
...               gateway traffic from a test suite is not CI material: it would take real
...               money from a real merchant account.
...
...               So the numbered files cover what can be exercised without a gateway: the
...               schema, the money actions' refusal path, the refund guard rails, and read
...               access over whatever orders the deployment holds.
Resource          resources/paymentinvoice.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Delete Payment Invoice Seed Data
