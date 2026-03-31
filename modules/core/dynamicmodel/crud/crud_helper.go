package crud

import (
	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	coredyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/baserepo"
)

func Create[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param dyn.CreateParam[TDomain, TDomainPtr],
) (*dyn.OpResult[TDomain], error) {

	dynamicRepo := param.BaseRepoGetter.GetBaseRepo()
	schema := dynamicRepo.GetSchema()
	fieldData := param.Data.GetFieldData()
	newModel := TDomainPtr(new(TDomain))
	newModel.SetFieldData(fieldData)

	flow := coredyn.StartValidationFlow()
	clientErrs, err := flow.
		Step(func(vErrs *ft.ClientErrors) error {
			if param.BeforeValidation == nil {
				return nil
			}
			result, err := param.BeforeValidation(ctx, newModel)
			if err == nil {
				fieldData = result.GetFieldData()
			}
			return err
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			result, clientErrs := schema.Validate(fieldData)
			if clientErrs != nil {
				*vErrs = clientErrs
			} else {
				fieldData = result
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			validateUniques(ctx, fieldData, dynamicRepo, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.ValidateExtra == nil {
				return nil
			}
			newModel.SetFieldData(fieldData)
			return param.ValidateExtra(ctx, newModel, vErrs)
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.AfterValidation != nil {
				result, err := param.AfterValidation(ctx, newModel)
				if err == nil {
					fieldData = result.GetFieldData()
				}
				return err
			}
			return nil
		}).
		End()

	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dyn.OpResult[TDomain]{
			ClientErrors: clientErrs,
		}, nil
	}

	newModel.SetFieldData(fieldData)
	insRes, err := baserepo.Insert(ctx, dynamicRepo, newModel)
	if err != nil {
		return nil, err
	}
	if insRes.ClientErrors.Count() > 0 {
		return &dyn.OpResult[TDomain]{ClientErrors: insRes.ClientErrors}, nil
	}

	return &dyn.OpResult[TDomain]{Data: *newModel, HasData: true}, nil
}

type DeleteOneParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Cmd          dyn.DeleteOneCommand
}

func DeleteOne(ctx corectx.Context, param DeleteOneParam) (result *dyn.OpResult[dyn.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	querySchema := deleteOneSchema()
	sanitizedFields, cErrs := querySchema.ValidateStruct(param.Cmd)

	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}

	cmd := *(sanitizedFields.(*dyn.DeleteOneCommand))
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	delResult, err := baserepo.DeleteOne(ctx, dynamicRepo, dmodel.DynamicFields{
		basemodel.FieldId: cmd.Id,
	})

	return delResult, err
}

func deleteOneSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.delete_one_query",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.DeleteOneQuerySchemaBuilder()
		},
	)
}

type ExistsParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Query        dyn.ExistsQuery
}

func Exists(ctx corectx.Context, param ExistsParam) (result *dyn.OpResult[dyn.ExistsResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	sanitized, cErrs := existsQuerySchema().ValidateStruct(param.Query)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.ExistsResultData]{ClientErrors: cErrs}, nil
	}
	cmd := *(sanitized.(*dyn.ExistsQuery))
	keys := existsQueryToKeyMaps(cmd)
	repoOut, err := baserepo.Exists(ctx, param.DbRepoGetter.GetBaseRepo(), keys)
	if err != nil {
		return nil, err
	}
	if len(repoOut.ClientErrors) > 0 {
		return &dyn.OpResult[dyn.ExistsResultData]{ClientErrors: repoOut.ClientErrors}, nil
	}
	data := existsRepoResultToData(repoOut.Data)
	return &dyn.OpResult[dyn.ExistsResultData]{Data: data, HasData: true}, nil
}

func existsQueryToKeyMaps(query dyn.ExistsQuery) []dmodel.DynamicFields {
	keys := make([]dmodel.DynamicFields, len(query.Ids))
	for i, id := range query.Ids {
		keys[i] = dmodel.DynamicFields{basemodel.FieldId: id}
	}
	return keys
}

func existsRepoResultToData(repo dyn.RepoExistsResult) dyn.ExistsResultData {
	out := dyn.ExistsResultData{
		Existing:    make([]model.Id, 0),
		NotExisting: make([]model.Id, 0),
	}
	for _, f := range repo.Existing {
		out.Existing = append(out.Existing, idFromExistsFieldMap(f))
	}
	for _, f := range repo.NotExisting {
		out.NotExisting = append(out.NotExisting, idFromExistsFieldMap(f))
	}
	return out
}

func idFromExistsFieldMap(f dmodel.DynamicFields) model.Id {
	v, ok := f[basemodel.FieldId]
	if !ok {
		return ""
	}
	if id, ok := v.(model.Id); ok {
		return id
	}
	if s, ok := v.(string); ok {
		return model.Id(s)
	}
	return model.Id("")
}

func existsQuerySchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.exists_query",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.ExistsQuerySchemaBuilder()
		},
	)
}

type GetOneParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Query        dyn.GetOneQuery
}

func GetOne[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context, param GetOneParam,
) (*dyn.OpResult[TDomain], error) {
	querySchema := getOneSchema()
	sanitized, cErrs := querySchema.ValidateStruct(param.Query)
	if cErrs.Count() > 0 {
		return &dyn.OpResult[TDomain]{ClientErrors: cErrs}, nil
	}
	sanitizedQuery := sanitized.(*dyn.GetOneQuery)

	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	return baserepo.GetOne[TDomain, TDomainPtr](ctx, dynamicRepo, coredyn.RepoGetOneParam{
		Filter: dmodel.DynamicFields{
			basemodel.FieldId: sanitizedQuery.Id,
		},
		Columns: sanitizedQuery.Columns,
	})
}

func getOneSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.get_one_query",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.GetOneQuerySchemaBuilder()
		},
	)
}

type SearchParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Query        dyn.SearchQuery
}

func Search[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context, param SearchParam,
) (*dyn.OpResult[dyn.PagedResultData[TDomain]], error) {
	querySchema := searchSchema()
	sanitized, cErrs := querySchema.ValidateStruct(param.Query)

	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.PagedResultData[TDomain]]{ClientErrors: cErrs}, nil
	}

	sanitizedQuery := *(sanitized.(*dyn.SearchQuery))
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	return baserepo.Search[TDomain, TDomainPtr](ctx, dynamicRepo, coredyn.RepoSearchParam{
		Graph:   sanitizedQuery.Graph,
		Columns: sanitizedQuery.Columns,
		Page:    sanitizedQuery.Page,
		Size:    sanitizedQuery.Size,
	})
}

func searchSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.search_query",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.SearchQuerySchemaBuilder()
		},
	)
}

func validateUniques(ctx corectx.Context, data dmodel.DynamicFields, dbRepo coredyn.BaseRepository, vErrs *ft.ClientErrors) error {
	collRes, err := dbRepo.CheckUniqueCollisions(ctx, data)
	if err != nil {
		return err
	}
	if len(collRes.ClientErrors) > 0 {
		for _, item := range collRes.ClientErrors {
			vErrs.Append(item)
		}
		return nil
	}
	collidingKeys := collRes.Data

	if len(collidingKeys) > 0 {
		field := collidingKeys[0][0]
		vErrs.Append(*ft.NewBusinessViolation(
			field,
			"common.err_unique_constraint_violated",
			"unique constraint violated {{.uniques}}",
			map[string]any{"uniques": collidingKeys},
		))
	}
	return nil
}

func SetIsArchived(
	ctx corectx.Context,
	dbRepoGetter coredyn.BaseRepoGetter,
	cmd dyn.SetIsArchivedCommand,
) (*dyn.OpResult[dyn.MutateResultData], error) {
	cmdSchema := setIsArchivedSchema()
	sanitizedCmd, cErrs := cmdSchema.ValidateStruct(cmd, true)

	if cErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: cErrs}, nil
	}

	cmd = *(sanitizedCmd.(*dyn.SetIsArchivedCommand))
	result, err := UpdateRegardless(ctx, UpdateRegardlessParam{
		Action:       "setIsArchived",
		DbRepoGetter: dbRepoGetter,
		Data: dmodel.DynamicFields{
			basemodel.FieldId:         cmd.Id,
			basemodel.FieldEtag:       cmd.Etag,
			basemodel.FieldIsArchived: cmd.IsArchived,
		},
	})

	return result, err
}

func setIsArchivedSchema() *dmodel.ModelSchema {
	return dmodel.GetOrRegisterSchema(
		"core.set_archived_command",
		func() *dmodel.ModelSchemaBuilder {
			return dyn.SetArchivedCommandSchemaBuilder()
		},
	)
}

type UpdateParam[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
] struct {
	Action           string
	DbRepoGetter     coredyn.BaseRepoGetter
	Data             dmodel.DynamicModelGetter
	BeforeValidation dyn.BeforeValidationFunc[TDomainPtr]
	AfterValidation  dyn.AfterValidationFunc[TDomainPtr]
	ValidateExtra    dyn.ValidateExtraFunc[TDomainPtr]
}

func Update[
	TDomain any,
	TDomainPtr coredyn.DynamicModelPtr[TDomain],
](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
) (*dyn.OpResult[dyn.MutateResultData], error) {
	model := TDomainPtr(new(TDomain))
	model.SetFieldData(param.Data.GetFieldData())

	isExisting, clientErrs, err := runUpdateValidationFlow(ctx, param, model)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: clientErrs}, nil
	}

	if !isExisting {
		return &dyn.OpResult[dyn.MutateResultData]{HasData: false}, nil
	}

	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	return baserepo.Update(ctx, dynamicRepo, model.GetFieldData())
}

type UpdateRegardlessParam struct {
	Action       string
	DbRepoGetter coredyn.BaseRepoGetter
	Data         dmodel.DynamicFields
}

// UpdateRegardless updates a record without validation, but it still checks for existence and etag matching.
func UpdateRegardless(
	ctx corectx.Context,
	param UpdateRegardlessParam,
) (_ *dyn.OpResult[dyn.MutateResultData], err error) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()

	isExisting, clientErrs, err := runUpdateRegardlessCheckingFlow(ctx, param)
	if err != nil {
		return nil, err
	}

	if clientErrs != nil && clientErrs.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: clientErrs}, nil
	}

	if !isExisting {
		return &dyn.OpResult[dyn.MutateResultData]{HasData: false}, nil
	}

	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	return baserepo.Update(ctx, dynamicRepo, param.Data)
}

func runUpdateValidationFlow[TDomain any, TDomainPtr coredyn.DynamicModelPtr[TDomain]](
	ctx corectx.Context,
	param UpdateParam[TDomain, TDomainPtr],
	model TDomainPtr,
) (bool, ft.ClientErrors, error) {
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	schema := dynamicRepo.GetSchema()

	isExisting := false
	cErr, err := coredyn.StartValidationFlow().
		Step(func(vErrs *ft.ClientErrors) error {
			if param.BeforeValidation == nil {
				return nil
			}
			result, err := param.BeforeValidation(ctx, model)
			if err == nil {
				model.SetFieldData(result.GetFieldData())
			}
			return err
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			result, clientErrs := schema.Validate(model.GetFieldData(), true)
			if clientErrs != nil {
				*vErrs = clientErrs
			} else {
				model.SetFieldData(result)
			}
			return nil
		}).
		StepS(func(vErrs *ft.ClientErrors, stopFlow func()) error {
			existing, err := checkExistenceAndEtag(ctx, schema, dynamicRepo, model.GetFieldData(), vErrs)
			if err != nil {
				return err
			}
			isExisting = existing
			if !existing {
				stopFlow()
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			validateUniques(ctx, model.GetFieldData(), dynamicRepo, vErrs)
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.ValidateExtra == nil {
				return nil
			}
			return param.ValidateExtra(ctx, model, vErrs)
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			if param.AfterValidation == nil {
				return nil
			}
			result, err := param.AfterValidation(ctx, model)
			if err == nil {
				model.SetFieldData(result.GetFieldData())
			}
			return err
		}).
		End()

	return isExisting, cErr, err
}

func runUpdateRegardlessCheckingFlow(
	ctx corectx.Context,
	param UpdateRegardlessParam,
) (bool, ft.ClientErrors, error) {
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	schema := dynamicRepo.GetSchema()

	isExisting := false
	cErr, err := coredyn.StartValidationFlow().
		StepS(func(vErrs *ft.ClientErrors, stopFlow func()) error {
			existing, err := checkExistenceAndEtag(ctx, schema, dynamicRepo, param.Data, vErrs)
			if err != nil {
				return err
			}
			isExisting = existing
			if !existing {
				stopFlow()
			}
			return nil
		}).
		Step(func(vErrs *ft.ClientErrors) error {
			validateUniques(ctx, param.Data, dynamicRepo, vErrs)
			return nil
		}).
		End()

	return isExisting, cErr, err
}

type ManageM2mParam struct {
	Action         string
	DbRepoGetter   coredyn.BaseRepoGetter
	DestSchemaName string
	Associations   []coredyn.RepoM2mAssociation
	Desociations   []coredyn.RepoM2mAssociation
}

func ManageM2m(ctx corectx.Context, param ManageM2mParam) (
	result *dyn.OpResult[dyn.MutateResultData], err error,
) {
	defer func() {
		if e := ft.RecoverPanicFailedTo(recover(), param.Action); e != nil {
			err = e
		}
	}()
	dynamicRepo := param.DbRepoGetter.GetBaseRepo()
	repoOut, err := dynamicRepo.ManageM2m(
		ctx, param.DestSchemaName, param.Associations, param.Desociations,
	)
	if err != nil {
		return nil, err
	}
	if repoOut.ClientErrors.Count() > 0 {
		return &dyn.OpResult[dyn.MutateResultData]{ClientErrors: repoOut.ClientErrors}, nil
	}
	data := dyn.MutateResultData{
		AffectedCount: repoOut.Data,
		AffectedAt:    model.NewModelDateTime(),
	}
	return &dyn.OpResult[dyn.MutateResultData]{Data: data, HasData: true}, nil
}

func checkExistenceAndEtag(
	ctx corectx.Context,
	schema *dmodel.ModelSchema,
	dynamicRepo coredyn.BaseRepository,
	fieldData dmodel.DynamicFields,
	vErrs *ft.ClientErrors,
) (bool, error) {
	primaryKeys := make(dmodel.DynamicFields)
	for _, key := range schema.KeyColumns() {
		primaryKeys[key] = fieldData[key]
	}

	dbRes, err := dynamicRepo.GetOne(ctx, coredyn.RepoGetOneParam{Filter: primaryKeys})
	if err != nil {
		return false, err
	}
	if len(dbRes.ClientErrors) > 0 {
		for _, item := range dbRes.ClientErrors {
			vErrs.Append(item)
		}
		return false, nil
	}
	if !dbRes.HasData {
		return false, nil
	}
	dbRecord := dbRes.Data

	dbEtag, hasEtag := dbRecord[basemodel.FieldEtag]
	etagMatched := dbEtag == fieldData[basemodel.FieldEtag]
	if hasEtag && !etagMatched {
		vErrs.Append(*ft.NewEtagMismatchedError())
	}
	return true, nil
}
