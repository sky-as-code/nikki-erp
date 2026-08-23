package dynamicengines

import (
	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	drif "github.com/sky-as-code/nikki-erp/modules/dynamicresource/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/models"
	"github.com/sky-as-code/nikki-erp/modules/inventory/domain/services"
	itProduct "github.com/sky-as-code/nikki-erp/modules/inventory/interfaces/product"
)

func productTemplateEngineSpec() engineSpec {
	return engineSpec{
		SchemaName:    models.ProductTemplateSchemaName,
		DefineActions: defineProductTemplateActions,
	}
}

// defineProductTemplateActions adds the template's custom actions and validation.
//
// Archiving is absent on purpose: AC-PROD-019's cascade to the variants is a method of the
// Products service (app.ProductAppServiceImpl.SetArchived), which overrides the built-in. It
// writes, and a write belongs to the service layer. See docs/wiki/07. ERP backend module.md §6.7.
func defineProductTemplateActions(engine drif.DynamicResourceEngine) error {
	// BR-PROD-TPL-005: a template with history must be archived, not deleted.
	err := engine.ModifyAction(drif.DynamicActionDelta{
		ActionName:    drif.ActionDelete,
		KeysToFetch:   templateKeysToFetch,
		ValidateExtra: validateTemplateDelete(engine),
	})
	if err != nil {
		return errors.Wrap(err, "failed to attach product template delete guard")
	}

	// BR §8.2: bring the template's variants in step with its attribute configuration.
	err = engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  "generate_variants",
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    ":id/generate_variants",
		Permission:  drif.PermissionUpdate,
		KeysToFetch: templateKeysToFetch,
		MainProcess: processGenerateVariants,
	})
	if err != nil {
		return errors.Wrap(err, "failed to define generate_variants")
	}

	// BR §14.4: turn a template plus chosen attribute values into the concrete variant a
	// transaction line must reference.
	err = engine.DefineAction(drif.DynamicActionDefinition{
		ActionName:  "resolve_selection",
		ActionType:  drif.ActionTypeGeneric,
		RestPath:    "resolve_selection",
		Permission:  drif.PermissionRead,
		MainProcess: processResolveSelection,
	})
	return errors.Wrap(err, "failed to define resolve_selection")
}

// processGenerateVariants runs the generation capability through the derived resource service.
//
// The service is the one installed by SetResourceService during Init, so the assertion holds for
// every request; a failure here means the wiring was skipped, which is a programming error
// rather than something a caller can fix.
func processGenerateVariants(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	productSvc, err := asProductService(input)
	if err != nil {
		return nil, err
	}

	// The pipeline already fetched the template for KeysToFetch; its absence is a missing
	// record, not a malformed request.
	if input.FoundModel == nil {
		return &drif.ActionResult{HasData: false}, nil
	}

	templateId := derefId(models.NewProductTemplateFrom(*input.FoundModel).GetId())
	result, err := productSvc.GenerateVariants(ctx, itProduct.GenerateVariantsQuery{TemplateId: templateId})
	if err != nil {
		return nil, errors.Wrap(err, "processGenerateVariants")
	}
	if result.ClientErrors.Count() > 0 {
		return &drif.ActionResult{ClientErrors: result.ClientErrors}, nil
	}

	// HasData is always true: "nothing to generate" is a valid answer with a payload, and a
	// false here would be reported to the caller as a 404.
	return &drif.ActionResult{
		Data:    itProduct.NewGenerateVariantsView(result.Data),
		HasData: true,
	}, nil
}

// processResolveSelection resolves a chosen attribute combination to a variant.
//
// It validates the payload itself rather than declaring a ParamSchema: the selections are a
// nested array, and without validation a malformed body decodes to an empty selection list,
// which would resolve to the empty combination — a different, existing variant — and answer 200.
func processResolveSelection(
	ctx corectx.Context, input drif.ProcessInput,
) (*drif.ActionResult, error) {
	productSvc, err := asProductService(input)
	if err != nil {
		return nil, err
	}

	query, vErrs := buildResolveSelectionQuery(input.Params)
	if vErrs.Count() > 0 {
		return &drif.ActionResult{ClientErrors: *vErrs}, nil
	}

	result, err := productSvc.ResolveProductSelection(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "processResolveSelection")
	}
	if result.ClientErrors.Count() > 0 {
		return &drif.ActionResult{ClientErrors: result.ClientErrors}, nil
	}

	// The service reports HasData=false when the template does not exist, which the REST layer
	// turns into a missing record. Forcing it true here would answer 200 with the combination key
	// of a product line that cannot be resolved against.
	if !result.HasData {
		return &drif.ActionResult{HasData: false}, nil
	}

	return &drif.ActionResult{
		Data:    itProduct.NewResolveProductSelectionView(result.Data),
		HasData: true,
	}, nil
}

func templateKeysToFetch(params dmodel.DynamicFields) dmodel.DynamicFields {
	return dmodel.DynamicFields{models.ProductTemplateFieldId: params[models.ProductTemplateFieldId]}
}

// validateTemplateDelete adapts the engine's validation callback to the delete rule.
//
// The rule counts the template's variants, so it reads through the Product Variant engine's
// repository rather than the template engine passed in here: product_template_id is a field of the
// variant schema, and searching the template's own repository for it fails as undefined.
func validateTemplateDelete(_ drif.DynamicResourceEngine) drif.ActionValidateExtraFn {
	return func(
		ctx corectx.Context, params dmodel.DynamicFields, _ *dmodel.DynamicFields, vErrs *ft.ClientErrors,
	) error {
		templateId := derefId(models.NewProductTemplateFrom(params).GetId())
		variantEngine, err := services.EngineFor(models.ProductVariantSchemaName)
		if err != nil {
			return err
		}
		return services.AssertTemplateDeletable(ctx, variantEngine.ResourceRepository(), templateId, vErrs)
	}
}
