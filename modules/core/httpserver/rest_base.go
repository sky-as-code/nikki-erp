package httpserver

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/safe"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type RestBase struct {
	ConfigSvc config.ConfigService
	Logger    logging.LoggerService
	CqrsBus   cqrs.CqrsBus
}

func JsonCreated(echoCtx echo.Context, data any) error {
	return echoCtx.JSON(http.StatusCreated, data)
}

func JsonOk(echoCtx echo.Context, data any) error {
	return echoCtx.JSON(http.StatusOK, data)
}

func JsonBadRequest(echoCtx echo.Context, err any) error {
	return echoCtx.JSON(http.StatusBadRequest, err)
}

type RestCreateResponse struct {
	Id        model.Id   `json:"id"`
	CreatedAt int64      `json:"createdAt"`
	Etag      model.Etag `json:"etag"`
}

func (this *RestCreateResponse) FromEntity(src createdEntity) {
	this.Id = *src.GetId()
	this.CreatedAt = *safe.Indirect(src.GetCreatedAt(), func(srcTime time.Time) *int64 {
		milli := srcTime.UnixMilli()
		return &milli
	})
	this.Etag = *src.GetEtag()
}

func (this *RestCreateResponse) FromNonEntity(src any) {
	model.MustCopy(src, this)
}

type RestUpdateResponse struct {
	Id        model.Id   `json:"id"`
	UpdatedAt int64      `json:"updatedAt"`
	Etag      model.Etag `json:"etag"`
}

func (this *RestUpdateResponse) FromEntity(src updatedEntity) {
	this.Id = *src.GetId()
	this.Etag = *src.GetEtag()
	this.UpdatedAt = *safe.Indirect(src.GetUpdatedAt(), func(srcTime time.Time) *int64 {
		milli := srcTime.UnixMilli()
		return &milli
	})
}

func (this *RestUpdateResponse) FromNonEntity(src any) {
	model.MustCopy(src, this)
}

type RestDeleteResponse struct {
	Id        model.Id `json:"id"`
	DeletedAt int64    `json:"deletedAt"`
}

func (this *RestDeleteResponse) FromEntity(src deletedEntity) {
	this.Id = *src.GetId()
	this.DeletedAt = *safe.Indirect(src.GetDeletedAt(), func(srcTime time.Time) *int64 {
		milli := srcTime.UnixMilli()
		return &milli
	})
}

func (this *RestDeleteResponse) FromNonEntity(src any) {
	model.MustCopy(src, this)
}

type RestSearchResponse[TItem any] struct {
	Items []TItem `json:"items"`
	Total int     `json:"total"`
	Page  int     `json:"page"`
	Size  int     `json:"size"`
}

type createdEntity interface {
	GetId() *model.Id
	GetCreatedAt() *time.Time
	GetEtag() *model.Etag
}

type updatedEntity interface {
	GetId() *model.Id
	GetUpdatedAt() *time.Time
	GetEtag() *model.Etag
}

type deletedEntity interface {
	GetId() *model.Id
	GetDeletedAt() *time.Time
}
