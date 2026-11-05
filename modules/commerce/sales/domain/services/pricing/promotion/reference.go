package promotion

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// ReferenceInput is what the calculation reference covers (D4). GrandTotal is the order total after
// tax, which only the caller knows.
type ReferenceInput struct {
	OrderId       string
	StageVersion  int32
	CurrencyScale int32
	Programs      []ProgramResult
	GrandTotal    decimal.Decimal
}

// CalculationReference is a stateless sha256 (hex) over the order id and stage version, the applied
// programs sorted by id with their config_version, reward, amounts and gift quantity, and the grand
// total. Only applied programs count: a candidate that changes nothing the customer pays or receives
// must not force a requote (BR-T02). Money is rendered at the currency scale and quantities without
// trailing zeros, so equal values always hash equal.
func CalculationReference(input ReferenceInput) string {
	applied := []ProgramResult{}
	for _, program := range input.Programs {
		if program.Outcome == OutcomeApplied {
			applied = append(applied, program)
		}
	}
	sort.SliceStable(applied, func(i, j int) bool { return applied[i].PromotionId < applied[j].PromotionId })

	var builder strings.Builder
	writeField(&builder, "order", input.OrderId, strconv.FormatInt(int64(input.StageVersion), 10))
	for _, program := range applied {
		writeProgram(&builder, program, input.CurrencyScale)
	}
	writeField(&builder, "grand_total", input.GrandTotal.StringFixed(input.CurrencyScale))

	digest := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(digest[:])
}

func writeProgram(builder *strings.Builder, program ProgramResult, scale int32) {
	giftVariant, giftQuantity := "", ""
	if program.Gift != nil {
		giftVariant = program.Gift.ProductVariantId
		giftQuantity = program.Gift.Quantity.String()
	}
	writeField(builder, "promotion",
		program.PromotionId,
		strconv.FormatInt(int64(program.ConfigVersion), 10),
		program.RewardId,
		program.DiscountUntaxedAmount.StringFixed(scale),
		program.TaxReductionAmount.StringFixed(scale),
		giftVariant,
		giftQuantity,
	)
}

// writeField length-prefixes every value so no concatenation of values can collide with another.
func writeField(builder *strings.Builder, tag string, values ...string) {
	builder.WriteString(tag)
	for _, value := range values {
		builder.WriteString("|")
		builder.WriteString(strconv.Itoa(len(value)))
		builder.WriteString(":")
		builder.WriteString(value)
	}
	builder.WriteString("\n")
}
