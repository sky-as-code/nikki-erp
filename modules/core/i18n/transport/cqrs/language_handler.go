package cqrs

import (
	"context"

	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/modules/core/cqrs"
	it "github.com/sky-as-code/nikki-erp/modules/core/i18n/interfaces"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

func NewLanguageHandler(languageSvc it.LanguageService, logger logging.LoggerService) *LanguageHandler {
	return &LanguageHandler{
		LanguageSvc: languageSvc,
	}
}

type LanguageHandler struct {
	LanguageSvc it.LanguageService
}

func (this *LanguageHandler) GetCurrentLangCode(ctx context.Context, packet *cqrs.RequestPacket[it.GetCurrentLangCodeQuery]) (*cqrs.Reply[it.GetCurrentLangCodeResult], error) {
	result, err := this.LanguageSvc.GetCurrentLangCode(ctx)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.GetCurrentLangCodeResult]{
		Result: *result,
	}
	return reply, nil
}

func (this *LanguageHandler) ListEnabledLangCodes(ctx context.Context, packet *cqrs.RequestPacket[it.ListEnabledLangCodesQuery]) (*cqrs.Reply[it.ListEnabledLangCodesResult], error) {
	query := packet.Request()
	result, err := this.LanguageSvc.ListEnabledLangCodes(ctx, *query)
	ft.PanicOnErr(err)

	reply := &cqrs.Reply[it.ListEnabledLangCodesResult]{
		Result: *result,
	}
	return reply, nil
}
