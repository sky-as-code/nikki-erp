package services

import (
	"testing"

	dmodel "github.com/sky-as-code/nikki-erp/common/dynamicmodel/model"
	"github.com/sky-as-code/nikki-erp/modules/sales/domain/models"
)

// The parts of the rule-set read that decide something without touching a repository: which
// pricelists could rank at a place, and how an id list becomes a search condition.

const (
	thisPoint    = "SP1"
	otherPoint   = "SP2"
	thisChannel  = "CH1"
	otherChannel = "CH2"
)

func pricelistScoped(pointId, channelId string) dmodel.DynamicFields {
	record := dmodel.DynamicFields{models.SalesPricelistFieldId: "PL1"}
	if pointId != "" {
		record[models.SalesPricelistFieldSalesPointId] = pointId
	}
	if channelId != "" {
		record[models.SalesPricelistFieldSalesChannelId] = channelId
	}
	return record
}

// A list naming this point applies here, whatever channel it also names: the point is the most
// specific scope there is, and nothing below it can overrule it.
func TestPointScopedListAppliesAtItsOwnPoint(t *testing.T) {
	if !pricelistAppliesAt(pricelistScoped(thisPoint, ""), thisPoint, thisChannel) {
		t.Fatal("a list scoped to this point must travel")
	}
	if !pricelistAppliesAt(pricelistScoped(thisPoint, otherChannel), thisPoint, thisChannel) {
		t.Fatal("the point wins over the channel, so a foreign channel must not exclude it")
	}
}

// A list belonging to another machine can never price this one. Shipping it would hand the caller
// a rule it must never choose and no way to know that.
func TestPointScopedListDoesNotTravelToAnotherPoint(t *testing.T) {
	if pricelistAppliesAt(pricelistScoped(otherPoint, ""), thisPoint, thisChannel) {
		t.Fatal("another point's list must not travel")
	}
}

// A list scoped to a channel applies to every point on that channel and to no other.
func TestChannelScopedListFollowsItsChannel(t *testing.T) {
	if !pricelistAppliesAt(pricelistScoped("", thisChannel), thisPoint, thisChannel) {
		t.Fatal("a list scoped to this kiosk's channel must travel")
	}
	if pricelistAppliesAt(pricelistScoped("", otherChannel), thisPoint, thisChannel) {
		t.Fatal("another channel's list must not travel")
	}
}

// Naming neither scope is how a global list is expressed — an absent value, not a missing one.
func TestUnscopedListAppliesEverywhere(t *testing.T) {
	if !pricelistAppliesAt(pricelistScoped("", ""), thisPoint, thisChannel) {
		t.Fatal("a list naming no scope is the global one and must travel")
	}
}

// A kiosk whose point sits on no channel still gets the global lists. Excluding them would leave a
// misconfigured machine with no prices at all rather than with its fallback ones.
func TestUnscopedListSurvivesAPointWithNoChannel(t *testing.T) {
	if !pricelistAppliesAt(pricelistScoped("", ""), thisPoint, "") {
		t.Fatal("a global list must survive a point that names no channel")
	}
	if pricelistAppliesAt(pricelistScoped("", thisChannel), thisPoint, "") {
		t.Fatal("a channel-scoped list must not match a point that is on no channel")
	}
}

// An empty IN matches nothing, so building one would be a query asked to answer nothing. Every
// caller short-circuits instead, and this is what lets them.
func TestInConditionRefusesAnEmptyList(t *testing.T) {
	if inCondition(models.SalesPricelistItemFieldSalesPricelistId, nil) != nil {
		t.Fatal("an empty id list must produce no condition")
	}
	if inCondition(models.SalesPricelistItemFieldSalesPricelistId, []string{"PL1"}) == nil {
		t.Fatal("a non-empty id list must produce a condition")
	}
}
