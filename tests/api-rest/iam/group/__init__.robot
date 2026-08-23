*** Settings ***
Documentation     IAM group suite. iam_group is the representative model for LangJson
...               behaviour: both "name" and "description" are jsonb columns holding one
...               object of BCP47 language code -> text, and "name" is unique.
...
...               File order encodes the flow: MODEL SCHEMA -> CREATE -> SEARCH, with
...               DELETE last as the cleanup step.
...
...               The search file is the one that matters. It proves that filtering and
...               ordering on a LangJson column dive into the acting user's own language,
...               resolved from that user's stored settings rather than sent by the client.
Resource          resources/iam.resource
Suite Setup       Create Authorized API Session
Suite Teardown    Run Keywords    Delete Locale Group Fixtures
...               AND    Delete Seed Data
Force Tags        iam    group
