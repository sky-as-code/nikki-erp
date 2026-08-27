package services

import (
	"github.com/sky-as-code/nikki-erp/common/model"
)

// instanceId identifies this process for as long as it runs.
//
// It is generated at start-up rather than read from configuration because two containers
// started from the same image must never share one. A shared id would let one instance
// mistake another's live lease for its own and take work that is already running, which is
// the exact failure the lease exists to prevent.
var instanceId = mustNewInstanceId()

// InstanceId returns this process's scheduler identity, stamped onto every attempt it claims
// so that an abandoned attempt can be attributed to the instance that died holding it.
func InstanceId() string {
	return instanceId
}

func mustNewInstanceId() string {
	generated, err := model.NewId()
	if err != nil {
		// Reached only if the ULID source fails, which would also break every other id in the
		// process. There is no useful degraded mode: an instance with no identity cannot
		// safely claim work.
		panic("jobscheduler: failed to generate an instance id: " + err.Error())
	}
	return string(*generated)
}
