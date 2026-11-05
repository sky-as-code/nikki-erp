package services

import (
	"regexp"

	"go.bryk.io/pkg/errors"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	ft "github.com/sky-as-code/nikki-erp/common/fault"
	"github.com/sky-as-code/nikki-erp/common/model"
	"github.com/sky-as-code/nikki-erp/common/util"
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
	dyn "github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel"
	"github.com/sky-as-code/nikki-erp/modules/core/dynamicmodel/basemodel"

	"github.com/sky-as-code/nikki-erp/modules/commerce/sales/domain/models"
)

// Generating voucher codes and issuing them (BR-V03, BR-V06, BR-V09). These are the entry points
// the order flow reuses: Confirm issues a next-order coupon through IssueNextOrderCoupon, and
// apply_voucher looks a code up through FindVoucherByCode.

const (
	voucherCodeCharset      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	voucherRandomLength     = 8
	voucherCodeMaxLength    = 32
	voucherGenerateAttempts = 5
)

var voucherPrefixPattern = regexp.MustCompile(`^[A-Z0-9-]{0,24}$`)

// GenerateVoucherCode draws prefix + 8 random characters of A-Z and 0-9 from a secure source
// (BR-V03). The prefix is normalised and must leave room for the random part.
func GenerateVoucherCode(prefix string) (string, error) {
	normalized := NormalizeVoucherCode(prefix)
	if !voucherPrefixPattern.MatchString(normalized) || len(normalized)+voucherRandomLength > voucherCodeMaxLength {
		return "", errors.New("GenerateVoucherCode: the prefix is not a valid code prefix")
	}
	random, err := util.GenerateRandomStr(voucherCodeCharset, voucherRandomLength)
	if err != nil {
		return "", errors.Wrap(err, "GenerateVoucherCode")
	}
	return normalized + random, nil
}

// ValidVoucherPrefix reports whether a batch prefix leaves room for the random part.
func ValidVoucherPrefix(prefix string) bool {
	normalized := NormalizeVoucherCode(prefix)
	return voucherPrefixPattern.MatchString(normalized) && len(normalized)+voucherRandomLength <= voucherCodeMaxLength
}

// FindVoucherByCode answers the org's voucher with this code, archived and revoked ones included,
// or nil. The comparison is on the normalised code (BR-V03).
func FindVoucherByCode(ctx corectx.Context, orgId, code string) (dmodel.DynamicFields, error) {
	repo, err := repoFor(models.SalesVoucherSchemaName)
	if err != nil {
		return nil, err
	}
	graph := &dmodel.SearchGraph{}
	// One And call: SearchGraph.And replaces the conditions of an earlier call, which dropped the
	// org filter and let another org's code match (AC-21).
	graph.And(
		*dmodel.NewSearchNode().NewCondition(basemodel.FieldOrgId, dmodel.Equals, orgId),
		*dmodel.NewSearchNode().NewCondition(models.SalesVoucherFieldNormalizedCode, dmodel.Equals,
			NormalizeVoucherCode(code)),
	)
	found, err := repo.Search(ctx, dyn.RepoSearchParam{Graph: graph, Page: 0, Size: 1, IncludeArchived: util.ToPtr(true)})
	if err != nil {
		return nil, errors.Wrap(err, "FindVoucherByCode")
	}
	if found == nil || !found.HasData || len(found.Data.Items) == 0 {
		return nil, nil
	}
	return found.Data.Items[0], nil
}

// freshVoucherCode draws codes until one is free in the org. Checking first rather than catching
// the unique violation keeps the caller's transaction usable: a failed INSERT aborts it.
func freshVoucherCode(ctx corectx.Context, orgId, prefix string) (string, error) {
	for attempt := 0; attempt < voucherGenerateAttempts; attempt++ {
		code, err := GenerateVoucherCode(prefix)
		if err != nil {
			return "", err
		}
		taken, err := FindVoucherByCode(ctx, orgId, code)
		if err != nil {
			return "", err
		}
		if taken == nil {
			return code, nil
		}
	}
	return "", errors.New("freshVoucherCode: no free code after several attempts")
}

// IssueCouponParams describes the next-order coupon a confirmed source order earns (BR-V09, V10).
type IssueCouponParams struct {
	OrgId         string
	PromotionId   string
	SourceOrderId string
	CustomerId    string
	ValidFrom     model.ModelDateTime
	ExpiresAt     *model.ModelDateTime
}

// IssuedVoucher is a voucher just created: its id and the code to hand to the customer.
type IssuedVoucher struct {
	Id   string `json:"id"`
	Code string `json:"code"`
}

// IssueNextOrderCoupon creates the single-use coupon of a next_order_coupon promotion, bound to the
// source order's customer, inside ctx's transaction. It never consumes the promotion's usage limit
// (BR-V11). The caller decides whether to issue at all and records the sales_voucher_issuance.
func IssueNextOrderCoupon(
	ctx corectx.Context, params IssueCouponParams,
) (*IssuedVoucher, *ft.ClientErrors, error) {
	code, err := freshVoucherCode(ctx, params.OrgId, "")
	if err != nil {
		return nil, nil, err
	}
	cmd := dmodel.DynamicFields{
		basemodel.FieldOrgId:                  params.OrgId,
		models.SalesVoucherFieldPromotionId:   params.PromotionId,
		models.SalesVoucherFieldCode:          code,
		models.SalesVoucherFieldSourceOrderId: params.SourceOrderId,
		models.SalesVoucherFieldCustomerId:    params.CustomerId,
		models.SalesVoucherFieldValidFrom:     params.ValidFrom,
		models.SalesVoucherFieldIssuedAt:      params.ValidFrom,
	}
	if params.ExpiresAt != nil {
		cmd[models.SalesVoucherFieldExpiresAt] = *params.ExpiresAt
	}
	return createVoucher(ctx, cmd)
}

// createVoucher runs the voucher's own create pipeline (normalising, type, limit, defaults).
func createVoucher(ctx corectx.Context, cmd dmodel.DynamicFields) (*IssuedVoucher, *ft.ClientErrors, error) {
	domSvc, err := domainServiceFor(models.SalesVoucherSchemaName)
	if err != nil {
		return nil, nil, err
	}
	created, err := domSvc.Create(ctx, cmd)
	if vErrs, err := clientErrorsOf(created, err); err != nil || vErrs != nil {
		return nil, vErrs, err
	}
	return &IssuedVoucher{
		Id: stringOf(created.Data, basemodel.FieldId), Code: stringOf(cmd, models.SalesVoucherFieldCode),
	}, nil, nil
}
