package crud

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	modelLib "github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

// Function signature types for CRUD operations
type SetDefaultFunc[TDomain any] func(model TDomain)
type SanitizeFunc[TDomain any] func(model TDomain)
type AssertBusinessRulesForCreateFunc[TDomain ValidatableForEdit] func(ctx Context, model TDomain, vErrs *ft.ValidationErrors) error
type AssertBusinessRulesForUpdateFunc[TDomain ValidatableForEdit] func(ctx Context, domainModel TDomain, modelFromDb TDomain, vErrs *ft.ValidationErrors) error
type AssertBusinessRulesForDeleteFunc[TDomain any, TCommand DeleteCommander[TDomain]] func(ctx Context, command TCommand, modelFromDb TDomain, vErrs *ft.ValidationErrors) error
type AssertExistsFunc[TDomain any] func(ctx Context, domainModel TDomain, vErrs *ft.ValidationErrors) (TDomain, error)
type RepoCreateFunc[TDomain any] func(ctx Context, model TDomain) (TDomain, error)
type RepoCreateBulkFunc[TDomain any] func(ctx Context, models []TDomain) ([]TDomain, error)
type RepoUpdateFunc[TDomain any] func(ctx Context, domainModel TDomain, prevEtag modelLib.Etag) (TDomain, error)
type RepoDeleteFunc[TDomain any] func(ctx Context, model TDomain) (int, error)
type RepoExistsOneFunc[TQuery any] func(ctx Context, query TQuery, vErrs *ft.ValidationErrors) (bool, error)
type RepoFindOneFunc[TDomain any, TQuery any] func(ctx Context, query TQuery, vErrs *ft.ValidationErrors) (TDomain, error)
type RepoListAllFunc[TDomain any, TQuery any] func(ctx Context, query TQuery, vErrs *ft.ValidationErrors) ([]TDomain, error)
type RepoSearchFunc[TDomain any, TQuery any] func(ctx Context, query TQuery, predicate *orm.Predicate, order []orm.OrderOption) (*PagedResult[TDomain], error)
type ParseSearchGraphFunc func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
type SetQueryDefaultsFunc[TQuery any] func(query *TQuery)
type ToFailureResultFunc[TResult any] func(vErrs *ft.ValidationErrors) *TResult
type ToSuccessResultFunc[TDomain any, TResult any] func(model TDomain) *TResult
type ToSuccessResultBulkFunc[TDomain any, TResult any] func(models []TDomain) *TResult
type ToSuccessResultWithCountFunc[TDomain any, TResult any] func(model TDomain, deletedCount int) *TResult
type ToSuccessResultBoolFunc[TResult any] func(existing bool) *TResult
type ToSuccessResultPagedFunc[TDomain any, TResult any] func(pagedResult *PagedResult[TDomain]) *TResult

type CreateParam[
	TDomain ValidatableForEdit,
	TCommand DomainModelProducer[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules AssertBusinessRulesForCreateFunc[TDomain]
	RepoCreate          RepoCreateFunc[TDomain]
	SetDefault          SetDefaultFunc[TDomain]
	Sanitize            SanitizeFunc[TDomain]
	ToFailureResult     ToFailureResultFunc[TResult]
	ToSuccessResult     ToSuccessResultFunc[TDomain, TResult]
}

func Create[
	TDomain ValidatableForEdit,
	TCommand DomainModelProducer[TDomain],
	TResult any,
](ctx Context, param CreateParam[TDomain, TCommand, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	modelToCreate := param.Command.ToDomainModel()
	vErrs, err := validateForCreate(ctx, modelToCreate, param.SetDefault, param.Sanitize, param.AssertBusinessRules)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(vErrs), nil
	}

	createdModel, err := param.RepoCreate(ctx, modelToCreate)
	ft.PanicOnErr(err)

	return param.ToSuccessResult(createdModel), nil
}

type CreateBulkParam[
	TDomain ValidatableForEdit,
	TCommand DomainModelBulkProducer[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules AssertBusinessRulesForCreateFunc[TDomain]
	RepoCreateBulk      RepoCreateBulkFunc[TDomain]
	SetDefault          SetDefaultFunc[TDomain]
	Sanitize            SanitizeFunc[TDomain]
	ToFailureResult     ToFailureResultFunc[TResult]
	ToSuccessResult     ToSuccessResultBulkFunc[TDomain, TResult]
}

func CreateBulk[
	TDomain ValidatableForEdit,
	TCommand DomainModelBulkProducer[TDomain],
	TResult any,
](ctx Context, param CreateBulkParam[TDomain, TCommand, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	vErrs := ft.NewValidationErrors()
	modelsToCreate := param.Command.ToDomainModels()

	for _, model := range modelsToCreate {
		modelVErrs, err := validateForCreate(ctx, model, param.SetDefault, param.Sanitize, param.AssertBusinessRules)
		ft.PanicOnErr(err)
		vErrs.Merge(*modelVErrs)
	}

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	createdModels, err := param.RepoCreateBulk(ctx, modelsToCreate)
	ft.PanicOnErr(err)

	return param.ToSuccessResult(createdModels), nil
}

func validateForCreate[
	TDomain ValidatableForEdit,
](ctx Context, model TDomain, setDefault SetDefaultFunc[TDomain], sanitize SanitizeFunc[TDomain], assertBusinessRules AssertBusinessRulesForCreateFunc[TDomain]) (*ft.ValidationErrors, error) {
	setDefault(model)

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = model.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			sanitize(model)
			if assertBusinessRules != nil {
				return assertBusinessRules(ctx, model, vErrs)
			}
			return nil
		}).
		End()

	return &vErrs, err
}

type UpdateParam[
	TDomain ValidatableForEdit,
	TCommand DomainModelProducer[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules AssertBusinessRulesForUpdateFunc[TDomain]
	AssertExists        AssertExistsFunc[TDomain]
	RepoUpdate          RepoUpdateFunc[TDomain]
	Sanitize            SanitizeFunc[TDomain]
	ToFailureResult     ToFailureResultFunc[TResult]
	ToSuccessResult     ToSuccessResultFunc[TDomain, TResult]
}

func Update[
	TDomain ValidatableForEdit,
	TCommand DomainModelProducer[TDomain],
	TResult any,
](ctx Context, param UpdateParam[TDomain, TCommand, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	modelToUpdate := param.Command.ToDomainModel()
	etagger, hasEtag := any(modelToUpdate).(Etagger)

	vErrs, err := validateForUpdate(ctx, modelToUpdate, param.AssertExists, param.Sanitize, param.AssertBusinessRules)
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(vErrs), nil
	}

	var prevEtag *modelLib.Etag
	if hasEtag {
		prevEtag = etagger.GetEtag()
		etagger.SetEtag(*modelLib.NewEtag())
	}
	// TODO: Passing nuillable prevTag
	updatedModel, err := param.RepoUpdate(ctx, modelToUpdate, *prevEtag)
	ft.PanicOnErr(err)

	// if updatedModel == nil {
	// 	vErrs.Appendf(param.Resource, "not found")
	// 	return nil, errors.New("model was not found during update execution")
	// }

	return param.ToSuccessResult(updatedModel), nil
}

type UpdateBulkParam[
	TDomain ValidatableForEdit,
	TCommand DomainModelBulkProducer[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules AssertBusinessRulesForUpdateFunc[TDomain]
	AssertExists        AssertExistsFunc[TDomain]
	RepoUpdate          RepoUpdateFunc[TDomain]
	Sanitize            SanitizeFunc[TDomain]
	ToFailureResult     ToFailureResultFunc[TResult]
	ToSuccessResult     ToSuccessResultBulkFunc[TDomain, TResult]
}

func UpdateBulk[
	TDomain ValidatableForEdit,
	TCommand DomainModelBulkProducer[TDomain],
	TResult any,
](ctx Context, param UpdateBulkParam[TDomain, TCommand, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	var vErrs ft.ValidationErrors
	modelsToUpdate := param.Command.ToDomainModels()

	for _, model := range modelsToUpdate {
		modelVErrs, err := validateForUpdate(ctx, model, param.AssertExists, param.Sanitize, param.AssertBusinessRules)
		ft.PanicOnErr(err)
		vErrs.Merge(*modelVErrs)
	}

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	var updatedModels []TDomain
	var prevEtag *modelLib.Etag
	for _, modelToUpdate := range modelsToUpdate {
		etagger, hasEtag := any(modelToUpdate).(Etagger)
		if hasEtag {
			prevEtag = etagger.GetEtag()
			etagger.SetEtag(*modelLib.NewEtag())
		}
		// TODO: Passing nuillable prevTag
		updated, err := param.RepoUpdate(ctx, modelToUpdate, *prevEtag)
		updatedModels = append(updatedModels, updated)
		ft.PanicOnErr(err)
	}

	// if updatedModel == nil {
	// 	vErrs.Appendf(param.Resource, "not found")
	// 	return nil, errors.New("model was not found during update execution")
	// }

	return param.ToSuccessResult(updatedModels), nil
}

func validateForUpdate[
	TDomain ValidatableForEdit,
](
	ctx Context, model TDomain,
	assertExists AssertExistsFunc[TDomain],
	sanitize SanitizeFunc[TDomain],
	assertBusinessRules AssertBusinessRulesForUpdateFunc[TDomain],
) (*ft.ValidationErrors, error) {
	etagger, hasEtag := any(model).(Etagger)

	var modelFromDb *TDomain
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = model.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			// var err error
			found, err := assertExists(ctx, model, vErrs)
			modelFromDb = &found
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if hasEtag {
				if *etagger.GetEtag() != *any(*modelFromDb).(Etagger).GetEtag() {
					vErrs.AppendEtagMismatched()
				}
			}
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			sanitize(model)
			if assertBusinessRules != nil {
				return assertBusinessRules(ctx, model, *modelFromDb, vErrs)
			}
			return nil
		}).
		End()

	return &vErrs, err
}

type DeleteHardParam[
	TDomain any,
	TCommand DeleteCommander[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules AssertBusinessRulesForDeleteFunc[TDomain, TCommand]
	AssertExists        AssertExistsFunc[TDomain]
	RepoDelete          RepoDeleteFunc[TDomain]
	ToFailureResult     ToFailureResultFunc[TResult]
	ToSuccessResult     ToSuccessResultWithCountFunc[TDomain, TResult]
}

func DeleteHard[
	TDomain any,
	TCommand DeleteCommander[TDomain],
	TResult any,
](ctx Context, param DeleteHardParam[TDomain, TCommand, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	modelToDel := param.Command.ToDomainModel()
	var modelFromDb TDomain
	flow := val.StartValidationFlow(param.Command)
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			if param.AssertExists != nil {
				modelFromDb, err = param.AssertExists(ctx, modelToDel, vErrs)
				return err
			}
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if param.AssertBusinessRules != nil {
				return param.AssertBusinessRules(ctx, param.Command, modelFromDb, vErrs)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	deletedCount, err := param.RepoDelete(ctx, modelFromDb)
	ft.PanicOnErr(err)

	return param.ToSuccessResult(modelFromDb, deletedCount), nil
}

type ExistsOneParam[
	TDomain any,
	TQuery Validatable,
	TResult any,
] struct {
	Action string
	Query  TQuery

	RepoExistsOne   RepoExistsOneFunc[TQuery]
	ToFailureResult ToFailureResultFunc[TResult]
	ToSuccessResult ToSuccessResultBoolFunc[TResult]
}

func ExistsOne[
	TDomain any,
	TQuery Validatable,
	TResult any,
](ctx Context, param ExistsOneParam[TDomain, TQuery, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	var existing bool
	flow := val.StartValidationFlow(param.Query)
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			existing, err = param.RepoExistsOne(ctx, param.Query, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	return param.ToSuccessResult(existing), nil
}

type GetOneParam[
	TDomain any,
	TQuery Validatable,
	TResult any,
] struct {
	Action string
	Query  TQuery

	RepoFindOne     RepoFindOneFunc[TDomain, TQuery]
	ToFailureResult ToFailureResultFunc[TResult]
	ToSuccessResult ToSuccessResultFunc[TDomain, TResult]
}

func GetOne[
	TDomain any,
	TQuery Validatable,
	TResult any,
](ctx Context, param GetOneParam[TDomain, TQuery, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	var modelFromDb TDomain
	flow := val.StartValidationFlow(param.Query)
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			modelFromDb, err = param.RepoFindOne(ctx, param.Query, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	return param.ToSuccessResult(modelFromDb), nil
}

type ListAllParam[
	TDomain any,
	TQuery Validatable,
	TResult any,
] struct {
	Action string
	Query  TQuery

	RepoListAll     RepoListAllFunc[TDomain, TQuery]
	ToFailureResult ToFailureResultFunc[TResult]
	ToSuccessResult ToSuccessResultBulkFunc[TDomain, TResult]
}

func ListAll[
	TDomain any,
	TQuery Validatable,
	TResult any,
](ctx Context, param ListAllParam[TDomain, TQuery, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	var modelsFromDb []TDomain
	flow := val.StartValidationFlow(param.Query)
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			modelsFromDb, err = param.RepoListAll(ctx, param.Query, vErrs)
			return err
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	return param.ToSuccessResult(modelsFromDb), nil
}

type SearchParam[
	TDomain any,
	TQuery Searchable,
	TResult any,
] struct {
	Action string
	Query  TQuery

	ParseSearchGraph ParseSearchGraphFunc
	RepoSearch       RepoSearchFunc[TDomain, TQuery]
	SetQueryDefaults SetQueryDefaultsFunc[TQuery]
	ToFailureResult  ToFailureResultFunc[TResult]
	ToSuccessResult  ToSuccessResultPagedFunc[TDomain, TResult]
}

func Search[
	TDomain any,
	TQuery Searchable,
	TResult any,
](ctx Context, param SearchParam[TDomain, TQuery, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	var predicate *orm.Predicate
	var order []orm.OrderOption
	var vErrsGraph ft.ValidationErrors

	param.SetQueryDefaults(&param.Query)
	vErrsModel := param.Query.Validate()
	if graph := param.Query.GetGraph(); graph != nil {
		predicate, order, vErrsGraph = param.ParseSearchGraph(graph)
		vErrsModel.Merge(vErrsGraph) // TODO: check if this is correct
	}

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return param.ToFailureResult(&vErrsModel), nil
	}

	pagedResult, err := param.RepoSearch(ctx, param.Query, predicate, order)
	ft.PanicOnErr(err)

	return param.ToSuccessResult(pagedResult), nil
}
