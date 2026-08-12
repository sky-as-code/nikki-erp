# HISTORICAL — do not add code here

This folder holds the **superseded** Products implementation. It is kept only as a reference while
the Odoo-like rewrite settles, and is scheduled for **manual deletion**.

## Status

- **Disconnected** from the module graph. `nikkierp/modules/inventory/index.go` no longer imports
  this package: `product.Init()` is not called and none of its eight schemas are registered, so it
  serves no routes and owns no live resources.
- **Nothing imports it.** The last external consumers —
  `coremart/modules/vending_machine/interfaces/external/inventory_product.go` and
  `inventory_variant.go` — were rewritten against the new `interfaces/product` port.
- It still compiles, which is the only reason it can sit here harmlessly.

## Rules

1. **Do not add new code to this folder.** New Products code belongs in
   `nikkierp/modules/inventory/{domain,dynamicengines,domain/services,app,interfaces,transport}/`.
2. **Do not import this package** from anywhere. A `grep -rn "modules/inventory/product"` across
   the repo must return hits only inside this folder itself.
3. **Do not "fix" anything here.** If something looks wrong, it is superseded — fix it in the new
   implementation.

## What replaced it

| Old (here) | New |
|---|---|
| `domain/product_entity.go` | `domain/models/product_template.go` + `product_template.json` |
| `domain/variant_entity.go` | `domain/models/product_variant.go` + `product_variant.json` |
| `domain/product_category_entity.go` | `domain/models/product_category.go` + `.json` |
| `domain/attribute*_entity.go` | `domain/models/product_attribute*.go` + `.json` |
| `app/*_service_impl.go`, `infra/repository/*_ent_repository.go`, `transport/restful/v1/*_rest.go` | the dynamic resource engine — see `dynamicengines/` |
| `interfaces/{product,variant,...}/` | `interfaces/product/` (one cross-module port) |

The model changed as well as the layering: a flat `inventory_product` became the
Template/Variant pair `inventory_product_template` / `inventory_product_variant`. See
`docs/requirements/inventory/product-br-draft-2.md` for the business requirement and
`frontend-coremart26-react-mono/ai-prompts/inventory/01-plans.md` for the migration plan.
