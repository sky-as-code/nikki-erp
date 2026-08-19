package uom

import (
	corectx "github.com/sky-as-code/nikki-erp/modules/core/context"
)

// The usage probe registry behind BR-UOM-ESS-020.
//
// A UoM already used by a transaction may not have its factor, type or category changed: doing so
// would reinterpret quantities already recorded, so a document written last year would silently
// come to mean a different amount of goods. The supported remedy is to archive the unit and create
// a replacement.
//
// Essential cannot answer "is this in use" on its own — the transactions live in the consuming
// modules, and Essential importing them would invert the dependency the ports exist to keep one
// way. So consumers register a probe instead: Essential asks, each consumer answers about its own
// data, and no import crosses the wrong direction.

// UomUsageProbe reports whether one module's transactions reference a UoM.
//
// A probe that cannot answer should return an ERROR rather than false. False means "nothing here
// uses it", which permits an edit that could reinterpret history; a probe that failed knows only
// that it does not know, and the two must not be confused.
type UomUsageProbe interface {
	// ModuleName identifies the probe in diagnostics, so a refusal can say which module still
	// references the unit rather than only that something does.
	ModuleName() string

	// IsUomInUse reports whether this module holds any record referencing the UoM.
	IsUomInUse(ctx corectx.Context, uomId string) (bool, error)
}

// probes is the registered set. Registration happens during module Init, which is single-threaded,
// and reads happen per request afterwards — so no lock is needed for the same reason the schema
// registry needs none.
var probes []UomUsageProbe

// RegisterUomUsageProbe adds a consumer's probe. Called from the consuming module's Init.
func RegisterUomUsageProbe(probe UomUsageProbe) {
	if probe == nil {
		return
	}
	probes = append(probes, probe)
}

// UomUsageProbes returns the registered probes, for Essential's own guard to consult.
func UomUsageProbes() []UomUsageProbe {
	return probes
}

// ResetUomUsageProbesForTest clears the registry between tests. Exported because the guard's tests
// live in another package, and a probe left over from one test would silently change the next.
func ResetUomUsageProbesForTest() {
	probes = nil
}
