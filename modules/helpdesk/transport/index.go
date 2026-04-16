package transport

import "github.com/sky-as-code/nikki-erp/modules/helpdesk/transport/restful"

func InitTransport() error {
	return restful.InitRestfulHandlers()
}
