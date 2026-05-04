package httpclient

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
//
// 	"github.com/sky-as-code/nikki-erp/modules/core/config"
// 	"github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
// 	"github.com/sky-as-code/nikki-erp/modules/core/logging"
// )
//
// func NewTestClient(cfg config.ConfigService, logger logging.LoggerService) {
// 	testConfig := &client.HttpClientConfig{
// 		Timeout: 10_000,
// 		TlsConfig: client.TlsConfig{
// 			InsecureSkipVerify:     false,
// 			IncludeSystemTrustedCa: true,
// 			CustomTrustedCas:       []string{},
// 		},
// 		ClientCertConfig: client.ClientCertConfig{
// 			Enabled:    true,
// 			Cert:       cfg.GetStr("CORE.HTTP_CLIENT.TEST_CLIENT_CERT"),
// 			PrivateKey: cfg.GetStr("CORE.HTTP_CLIENT.TEST_CLIENT_KEY"),
// 		},
// 	}
//
// 	client, err := client.NewHttpClient(testConfig)
// 	if err != nil {
// 		fmt.Println(err.Error())
// 	}
//
// 	caller := NewHttpCaller("https://localhost:4433", client, logger)
// 	_, err = caller.Do(context.Background(), &Request{
// 		Method: http.MethodGet,
// 		Path:   "/ping",
// 	})
// 	if err != nil {
// 		logger.Error("Test HTTP client mTLS", err)
// 	} else {
// 		logger.Infof("Test HTTP client mTLS SUCCESS")
// 	}
// }
