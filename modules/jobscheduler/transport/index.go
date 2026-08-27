package transport

import (
	"github.com/sky-as-code/nikki-erp/modules/jobscheduler/transport/restful"
)

// InitTransport mounts the REST surface.
//
// There are no CQRS handlers. The scheduler dispatches commands to other modules and never
// receives one: a module registers its jobs over HTTP like any other client, and a handler here
// would be a second door onto the same registration with its own permission story.
func InitTransport() error {
	return restful.InitRestfulHandlers()
}
