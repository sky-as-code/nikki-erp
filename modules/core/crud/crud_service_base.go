package crud

import (
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/orm"
	val "github.com/sky-as-code/nikki-erp/common/validator"
)

type CreateParam[
	TDomain ValidatableForEdit,
	TCommand DomainModelProducer[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules func(ctx Context, model TDomain, vErrs *ft.ValidationErrors) error
	RepoCreate          func(ctx Context, model TDomain) (TDomain, error)
	SetDefault          func(model TDomain)
	Sanitize            func(model TDomain)
	ToFailureResult     func(vErrs *ft.ValidationErrors) *TResult
	ToSuccessResult     func(model TDomain) *TResult
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
	param.SetDefault(modelToCreate)

	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = modelToCreate.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			param.Sanitize(modelToCreate)
			if param.AssertBusinessRules != nil {
				return param.AssertBusinessRules(ctx, modelToCreate, vErrs)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	createdModel, err := param.RepoCreate(ctx, modelToCreate)
	ft.PanicOnErr(err)

	return param.ToSuccessResult(createdModel), nil
}

type UpdateParam[
	TDomain Updateable,
	TCommand DomainModelProducer[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules func(ctx Context, domainModel TDomain, modelFromDb TDomain, vErrs *ft.ValidationErrors) error
	AssertExists        func(ctx Context, domainModel TDomain, vErrs *ft.ValidationErrors) (TDomain, error)
	RepoUpdate          func(ctx Context, domainModel TDomain, prevEtag model.Etag) (TDomain, error)
	Sanitize            func(model TDomain)
	ToFailureResult     func(vErrs *ft.ValidationErrors) *TResult
	ToSuccessResult     func(model TDomain) *TResult
}

func Update[
	TDomain Updateable,
	TCommand DomainModelProducer[TDomain],
	TResult any,
](ctx Context, param UpdateParam[TDomain, TCommand, TResult]) (result *TResult, err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	modelToUpdate := param.Command.ToDomainModel()

	var modelFromDb TDomain
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = modelToUpdate.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			modelFromDb, err = param.AssertExists(ctx, modelToUpdate, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if *modelToUpdate.GetEtag() != *modelFromDb.GetEtag() {
				vErrs.AppendEtagMismatched()
			}
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			param.Sanitize(modelToUpdate)
			if param.AssertBusinessRules != nil {
				return param.AssertBusinessRules(ctx, modelToUpdate, modelFromDb, vErrs)
			}
			return nil
			// return param.AssertUnique(ctx, modelToUpdate, modelFromDb, vErrs)
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return param.ToFailureResult(&vErrs), nil
	}

	prevEtag := modelToUpdate.GetEtag()
	modelToUpdate.SetEtag(*model.NewEtag())
	updatedModel, err := param.RepoUpdate(ctx, modelToUpdate, *prevEtag)
	ft.PanicOnErr(err)

	// if updatedModel == nil {
	// 	vErrs.Appendf(param.Resource, "not found")
	// 	return nil, errors.New("model was not found during update execution")
	// }

	return param.ToSuccessResult(updatedModel), nil
}

type DeleteHardParam[
	TDomain Updateable,
	TCommand Deletable[TDomain],
	TResult any,
] struct {
	Action  string
	Command TCommand

	AssertBusinessRules func(ctx Context, command TCommand, modelFromDb TDomain, vErrs *ft.ValidationErrors) error
	AssertExists        func(ctx Context, domainModel TDomain, vErrs *ft.ValidationErrors) (TDomain, error)
	//AssertExists        func(ctx Context, command TCommand, vErrs *ft.ValidationErrors) (*TDomain, error)
	RepoDelete      func(ctx Context, model TDomain) (int, error)
	ToFailureResult func(vErrs *ft.ValidationErrors) *TResult
	ToSuccessResult func(model TDomain, deletedCount int) *TResult
}

func DeleteHard[
	TDomain Updateable,
	TCommand Deletable[TDomain],
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

type GetOneParam[
	TDomain Updateable,
	TQuery Validatable,
	TResult any,
] struct {
	Action string
	Query  TQuery

	RepoFindOne     func(ctx Context, query TQuery, vErrs *ft.ValidationErrors) (TDomain, error)
	ToFailureResult func(vErrs *ft.ValidationErrors) *TResult
	ToSuccessResult func(model TDomain) *TResult
}

func GetOne[
	TDomain Updateable,
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

type SearchParam[
	TDomain any,
	TQuery Searchable,
	TResult any,
] struct {
	Action string
	Query  TQuery

	ParseSearchGraph func(criteria *string) (*orm.Predicate, []orm.OrderOption, ft.ValidationErrors)
	RepoSearch       func(ctx Context, query TQuery, predicate *orm.Predicate, order []orm.OrderOption) (*PagedResult[TDomain], error)
	SetQueryDefaults func(query *TQuery)
	ToFailureResult  func(vErrs *ft.ValidationErrors) *TResult
	ToSuccessResult  func(pagedResult *PagedResult[TDomain]) *TResult
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
