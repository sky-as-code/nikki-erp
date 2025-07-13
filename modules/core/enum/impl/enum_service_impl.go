package impl

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/sky-as-code/nikki-erp/common/defense"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	val "github.com/sky-as-code/nikki-erp/common/validator"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/core/enum/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	i18n "github.com/sky-as-code/nikki-erp/modules/core/i18n/interfaces"
)

func NewEnumServiceImpl(
	cqrsBus cqrs.CqrsBus,
	enumRepo it.EnumRepository,
	eventBus event.EventBus,
) it.EnumService {
	return &EnumServiceImpl{
		cqrsBus:  cqrsBus,
		enumRepo: enumRepo,
		eventBus: eventBus,
	}
}

type EnumServiceImpl struct {
	cqrsBus  cqrs.CqrsBus
	enumRepo it.EnumRepository
	eventBus event.EventBus
}

func (this *EnumServiceImpl) CreateEnum(ctx context.Context, cmd it.CreateEnumCommand) (result *it.CreateEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to create %s", cmd.EntityName); e != nil {
			err = e
		}
	}()

	enum := cmd.ToEnum()
	enum.SetDefaults()

	var langCodes []model.LanguageCode
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = enum.Validate(false)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			langCodes, err = this.valSanitizeStep(ctx, enum, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			this.assertEnumUnique(ctx, enum, *enum.Type, langCodes, vErrs)
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.CreateEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	enum, err = this.enumRepo.Create(ctx, *enum)
	ft.PanicOnErr(err)

	return &it.CreateEnumResult{
		Data:    enum,
		HasData: true,
	}, err
}

func (this *EnumServiceImpl) UpdateEnum(ctx context.Context, cmd it.UpdateEnumCommand) (result *it.UpdateEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to update %s", cmd.EntityName); e != nil {
			err = e
		}
	}()

	enum := cmd.ToEnum()

	// vErrs := enum.Validate(true)
	var langCodes []model.LanguageCode
	flow := val.StartValidationFlow()
	vErrs, err := flow.
		Step(func(vErrs *ft.ValidationErrors) error {
			*vErrs = enum.Validate(true)
			return nil
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			return this.assertCorrectEnum(ctx, enum, cmd.EntityName, vErrs)
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			langCodes, err = this.valSanitizeStep(ctx, enum, vErrs)
			return err
		}).
		Step(func(vErrs *ft.ValidationErrors) error {
			if enum.Value != nil {
				this.assertEnumUnique(ctx, enum, *enum.Type, langCodes, vErrs)
			}
			return nil
		}).
		End()
	ft.PanicOnErr(err)

	if vErrs.Count() > 0 {
		return &it.UpdateEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	prevEtag := enum.Etag
	enum.Etag = model.NewEtag()
	enum, err = this.enumRepo.Update(ctx, *enum, *prevEtag)
	ft.PanicOnErr(err)

	return &it.UpdateEnumResult{
		Data:    enum,
		HasData: true,
	}, err
}

func (this *EnumServiceImpl) valSanitizeStep(ctx context.Context, enum *it.Enum, vErrs *ft.ValidationErrors) ([]model.LanguageCode, error) {
	langCodes, err := this.getEnabledLanguages(ctx)
	if err != nil {
		return nil, err
	}

	this.sanitizeEnum(enum, langCodes, vErrs)
	return langCodes, nil
}

func (this *EnumServiceImpl) sanitizeEnum(enum *it.Enum, langCodes []model.LanguageCode, vErrs *ft.ValidationErrors) {
	newLabel, fieldCount, err := enum.Label.SanitizeClone(langCodes, false)
	ft.PanicOnErr(err)

	if fieldCount == 0 {
		vErrs.Append("label", "no enabled language")
	}
	enum.Label = newLabel

	if enum.Value != nil {
		enum.Value = util.ToPtr(defense.SanitizePlainText(*enum.Value))
	}
	if enum.Type != nil {
		enum.Type = util.ToPtr(defense.SanitizePlainText(*enum.Type))
	}
}

func (this *EnumServiceImpl) assertEnumUnique(
	ctx context.Context,
	enum *it.Enum,
	enumType string,
	langCodes []model.LanguageCode,
	errors *ft.ValidationErrors,
) {
	dbEnums, err := this.enumRepo.List(ctx, it.ListParam{
		Page: util.ToPtr(0),
		Size: util.ToPtr(math.MaxInt32), // Get all, no pagination
		Type: util.ToPtr(enumType),
	})
	ft.PanicOnErr(err)

	existFound := false
	for _, dbEnum := range dbEnums.Items {
		for _, langCode := range langCodes {
			targetLabel := (*enum.Label)[langCode]
			hasTranslation := len(targetLabel) > 0
			isTransExists := (targetLabel == (*dbEnum.Label)[langCode])
			if hasTranslation && isTransExists {
				errors.Append(fmt.Sprintf("label.%s", langCode), "label already exists in this language")
				existFound = true
			}
		}
		if existFound {
			return
		}
	}
}

func (this *EnumServiceImpl) assertCorrectEnum(ctx context.Context, enum *it.Enum, entityName string, vErrs *ft.ValidationErrors) error {
	dbEnum, err := this.enumRepo.FindById(ctx, *enum.Id)
	if err != nil {
		return err
	}

	if dbEnum == nil {
		vErrs.Appendf("id", "%s not found", entityName)
		return nil
	} else if *dbEnum.Etag != *enum.Etag {
		vErrs.Appendf("etag", "%s has been modified by another process", entityName)
		return nil
	}
	return nil
}

func (this *EnumServiceImpl) getEnabledLanguages(ctx context.Context) ([]model.LanguageCode, error) {
	query := i18n.ListEnabledLangCodesQuery{}
	result := i18n.ListEnabledLangCodesResult{}
	err := this.cqrsBus.Request(ctx, query, &result)
	ft.PanicOnErr(err)

	return result.Data, nil
}

func (this *EnumServiceImpl) DeleteEnum(ctx context.Context, cmd it.DeleteEnumCommand) (result *it.DeleteEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to delete %s", cmd.EntityName); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &it.DeleteEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	var deletedCount int
	if cmd.Id != nil {
		this.deleteById(ctx, *cmd.Id, &vErrs)
	} else if cmd.Type != nil {
		this.deleteMultiByType(ctx, *cmd.Type, &vErrs)
	}

	if vErrs.Count() > 0 {
		return &it.DeleteEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &it.DeleteEnumResult{
		Data: &it.DeleteEnumResultData{
			DeletedAt:    time.Now(),
			DeletedCount: deletedCount,
		},
		HasData: true,
	}, nil
}

func (this *EnumServiceImpl) deleteById(ctx context.Context, id model.Id, vErrs *ft.ValidationErrors) {
	deletedCount, err := this.enumRepo.DeleteById(ctx, id)
	ft.PanicOnErr(err)

	if deletedCount == 0 {
		vErrs.Append("id", "enum not found")
		return
	}

	// err = this.publishEnumDeletedEvent(ctx, enum)
	// ft.PanicOnErr(err)
}

func (this *EnumServiceImpl) deleteMultiByType(ctx context.Context, enumType string, vErrs *ft.ValidationErrors) {
	deletedCount, err := this.enumRepo.DeleteByType(ctx, enumType)
	ft.PanicOnErr(err)

	if deletedCount == 0 {
		vErrs.Append("type", "enum not found")
		return
	}

	// err = this.publishEnumDeletedEvent(ctx, enum)
	// ft.PanicOnErr(err)
}

// func (this *EnumServiceImpl) publishEnumDeletedEvent(ctx context.Context, enum *domain.Enum) error {
// 	eventId, err := ulid.New()
// 	if err != nil {
// 		return err
// 	}

// 	enumDeletedEvent := &EnumDeletedEvent{
// 		ID:        *enum.Id,
// 		DeletedBy: "", // You might want to get this from context or pass it as parameter
// 		EventID:   eventId.String(),
// 	}

// 	return this.eventBus.PublishEvent(ctx, "core.enum.deleted.done", enumDeletedEvent)
// }

func (this *EnumServiceImpl) EnumExists(ctx context.Context, query it.EnumExistsQuery) (result *it.EnumExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to check if %s exists", query.EntityName); e != nil {
			err = e
		}
	}()

	isExisting, err := this.enumRepo.Exists(ctx, query.Id)
	ft.PanicOnErr(err)

	return &it.EnumExistsResult{
		Data:    isExisting,
		HasData: true,
	}, nil
}

func (this *EnumServiceImpl) EnumExistsMulti(ctx context.Context, query it.EnumExistsMultiQuery) (result *it.EnumExistsMultiResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to check if multiple %s exist", query.EntityName); e != nil {
			err = e
		}
	}()

	existing, notExisting, err := this.enumRepo.ExistsMulti(ctx, query.Ids)
	ft.PanicOnErr(err)

	return &it.EnumExistsMultiResult{
		Data: &it.ExistsMultiResultData{
			Existing:    existing,
			NotExisting: notExisting,
		},
		HasData: true,
	}, nil
}

func (this *EnumServiceImpl) GetEnum(ctx context.Context, query it.GetEnumQuery) (result *it.GetEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to get %s", query.EntityName); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &it.GetEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	var enum *it.Enum
	if query.Id != nil {
		enum, err = this.enumRepo.FindById(ctx, *query.Id)
	} else if query.Value != nil {
		enum, err = this.enumRepo.FindByValue(ctx, *query.Value, *query.Type)
	}
	ft.PanicOnErr(err)

	if enum == nil {
		return &it.GetEnumResult{
			HasData: false,
		}, nil
	}

	return &it.GetEnumResult{
		Data:    enum,
		HasData: true,
	}, nil
}

func (this *EnumServiceImpl) ListEnums(ctx context.Context, query it.ListEnumsQuery) (result *it.ListEnumsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to list %s", query.EntityName); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	if vErrsModel.Count() > 0 {
		return &it.ListEnumsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	query.SetDefaults()
	enums, err := this.enumRepo.List(ctx, query)
	ft.PanicOnErr(err)

	return &it.ListEnumsResult{
		Data:    enums,
		HasData: len(enums.Items) > 0,
	}, nil
}

func (this *EnumServiceImpl) SearchEnums(ctx context.Context, query it.SearchEnumsQuery) (result *it.SearchEnumsResult, err error) {
	defer func() {
		if e := ft.RecoverPanicf(recover(), "failed to search %s", query.EntityName); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	predicate, order, vErrsGraph := this.enumRepo.ParseSearchGraph(query.Graph)

	vErrsModel.Merge(vErrsGraph)

	if vErrsModel.Count() > 0 {
		return &it.SearchEnumsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}
	query.SetDefaults()

	enums, err := this.enumRepo.Search(ctx, it.SearchParam{
		Predicate:  predicate,
		Order:      order,
		Page:       *query.Page,
		Size:       *query.Size,
		TypePrefix: query.TypePrefix,
	})
	ft.PanicOnErr(err)

	return &it.SearchEnumsResult{
		Data:    enums,
		HasData: len(enums.Items) > 0,
	}, nil
}
