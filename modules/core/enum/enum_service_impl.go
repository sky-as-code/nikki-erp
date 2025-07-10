package enum

import (
	"context"
	"time"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
)

func NewEnumServiceImpl(
	enumRepo EnumRepository,
	eventBus event.EventBus,
) EnumService {
	return &EnumServiceImpl{
		enumRepo: enumRepo,
		eventBus: eventBus,
	}
}

type EnumServiceImpl struct {
	enumRepo EnumRepository
	eventBus event.EventBus
}

func (this *EnumServiceImpl) CreateEnum(ctx context.Context, cmd CreateEnumCommand) (result *CreateEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to create enum"); e != nil {
			err = e
		}
	}()

	enum := cmd.ToEnum()

	vErrs := enum.Validate(false)
	this.assertEnumUnique(ctx, enum, &vErrs)
	if vErrs.Count() > 0 {
		return &CreateEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	enum, err = this.enumRepo.Create(ctx, *enum)
	ft.PanicOnErr(err)

	return &CreateEnumResult{
		Data:    enum,
		HasData: true,
	}, err
}

func (this *EnumServiceImpl) assertEnumUnique(ctx context.Context, enum *Enum, errors *ft.ValidationErrors) {
	if errors.Has("value") {
		return
	}
	dbEnum, err := this.enumRepo.FindByValue(ctx, *enum.Value, *enum.Type)
	ft.PanicOnErr(err)

	if dbEnum != nil {
		errors.Append("value", "value already exists for this enum type")
	}
}

func (this *EnumServiceImpl) UpdateEnum(ctx context.Context, cmd UpdateEnumCommand) (result *UpdateEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to update enum"); e != nil {
			err = e
		}
	}()

	enum := cmd.ToEnum()

	vErrs := enum.Validate(true)
	if enum.Value != nil {
		this.assertEnumUnique(ctx, enum, &vErrs)
	}
	if vErrs.Count() > 0 {
		return &UpdateEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	dbEnum, err := this.enumRepo.FindById(ctx, *enum.Id)
	ft.PanicOnErr(err)

	if dbEnum == nil {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("id", "enum not found")

		return &UpdateEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil

	} else if *dbEnum.Etag != *enum.Etag {
		vErrs = ft.NewValidationErrors()
		vErrs.Append("etag", "enum has been modified by another process")

		return &UpdateEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	enum.Etag = model.NewEtag()
	enum, err = this.enumRepo.Update(ctx, *enum)
	ft.PanicOnErr(err)

	return &UpdateEnumResult{
		Data:    enum,
		HasData: true,
	}, err
}

func (this *EnumServiceImpl) DeleteEnum(ctx context.Context, cmd DeleteEnumCommand) (result *DeleteEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to delete enum"); e != nil {
			err = e
		}
	}()

	vErrs := cmd.Validate()
	if vErrs.Count() > 0 {
		return &DeleteEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	var deletedCount int
	if cmd.Id != nil {
		this.deleteById(ctx, *cmd.Id, &vErrs)
	} else if cmd.EnumType != nil {
		this.deleteMultiByType(ctx, *cmd.EnumType, &vErrs)
	}

	if vErrs.Count() > 0 {
		return &DeleteEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	return &DeleteEnumResult{
		Data: &DeleteEnumResultData{
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

func (this *EnumServiceImpl) Exists(ctx context.Context, cmd EnumExistsCommand) (result *EnumExistsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to check if enum exists"); e != nil {
			err = e
		}
	}()

	isExisting, err := this.enumRepo.Exists(ctx, cmd.Id)
	ft.PanicOnErr(err)

	return &EnumExistsResult{
		Data:    isExisting,
		HasData: true,
	}, nil
}

func (this *EnumServiceImpl) ExistsMulti(ctx context.Context, cmd EnumExistsMultiCommand) (result *EnumExistsMultiResult, err error) {
	existing, notExisting, err := this.enumRepo.ExistsMulti(ctx, cmd.Ids)
	ft.PanicOnErr(err)

	return &EnumExistsMultiResult{
		Data: &ExistsMultiResultData{
			Existing:    existing,
			NotExisting: notExisting,
		},
		HasData: true,
	}, nil
}

func (this *EnumServiceImpl) GetEnum(ctx context.Context, query GetEnumQuery) (result *GetEnumResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to get enum"); e != nil {
			err = e
		}
	}()

	vErrs := query.Validate()
	if vErrs.Count() > 0 {
		return &GetEnumResult{
			ClientError: vErrs.ToClientError(),
		}, nil
	}

	var enum *Enum
	if query.Id != nil {
		enum, err = this.enumRepo.FindById(ctx, *query.Id)
	} else if query.Value != nil {
		enum, err = this.enumRepo.FindByValue(ctx, *query.Value, *query.EnumType)
	}
	ft.PanicOnErr(err)

	if enum == nil {
		return &GetEnumResult{
			HasData: false,
		}, nil
	}

	return &GetEnumResult{
		Data:    enum,
		HasData: true,
	}, nil
}

func (this *EnumServiceImpl) ListEnums(ctx context.Context, query ListEnumsQuery) (result *ListEnumsResult, err error) {
	defer func() {
		if e := ft.RecoverPanic(recover(), "failed to list enums"); e != nil {
			err = e
		}
	}()

	vErrsModel := query.Validate()
	if vErrsModel.Count() > 0 {
		return &ListEnumsResult{
			ClientError: vErrsModel.ToClientError(),
		}, nil
	}

	query.SetDefaults()
	enums, err := this.enumRepo.List(ctx, query)
	ft.PanicOnErr(err)

	return &ListEnumsResult{
		Data:    enums,
		HasData: len(enums.Items) > 0,
	}, nil
}
