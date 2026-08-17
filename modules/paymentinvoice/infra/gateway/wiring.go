// Package gateway builds the gateway registry this deployment will actually serve payments
// through, from configuration.
//
// It is the only place that knows all three adapters exist. The domain layer selects one by
// adapter code and the adapters themselves know nothing of each other, so this file is where
// adding a fourth gateway takes its one edit outside its own package.
package gateway

import (
	"strings"

	"go.bryk.io/pkg/errors"

	"github.com/sky-as-code/nikki-erp/modules/core/config"
	core "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/httpclient"
	httpclientclient "github.com/sky-as-code/nikki-erp/modules/core/httpclient/client"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/constants"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/momo"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/mpos"
	"github.com/sky-as-code/nikki-erp/modules/paymentinvoice/infra/gateway/vietqr"
	itGateway "github.com/sky-as-code/nikki-erp/modules/paymentinvoice/interfaces/gateway"
)

// BuildRegistry constructs the adapters this deployment has enabled and configured.
//
// A gateway is registered only when its ENABLED flag is set *and* it has an endpoint to call, so
// a deployment that uses one gateway needs only that one's credentials. An enabled gateway with
// no endpoint is a configuration mistake and is refused here rather than at the first payment,
// where it would present as a network failure at the worst possible moment.
//
// Each adapter gets its own HttpCaller, because the caller is constructed around a base URL and
// the three gateways are three different hosts.
func BuildRegistry(
	cfg config.ConfigService,
	httpClient *httpclientclient.HttpClient,
	logger logging.LoggerService,
) (*itGateway.Registry, error) {
	registry := itGateway.NewRegistry()

	if err := registerMomo(registry, cfg, httpClient, logger); err != nil {
		return nil, err
	}
	if err := registerMpos(registry, cfg, httpClient, logger); err != nil {
		return nil, err
	}
	if err := registerVietQr(registry, cfg, httpClient, logger); err != nil {
		return nil, err
	}

	if len(registry.Codes()) == 0 {
		// Not an error: a deployment that takes no card or wallet payments is legitimate, and
		// every attempt to pay through a method will report the gateway as unavailable.
		logger.Warnf("paymentinvoice: no payment gateway is enabled; no payment can be taken")
	}
	return registry, nil
}

func registerMomo(
	registry *itGateway.Registry,
	cfg config.ConfigService,
	httpClient *httpclientclient.HttpClient,
	logger logging.LoggerService,
) error {
	if !cfg.GetBool(constants.MomoEnabled, false) {
		return nil
	}

	endpoint, err := requireEndpoint(cfg, constants.MomoApiEndpoint, "MOMO")
	if err != nil {
		return err
	}

	adapter := momo.NewAdapter(momo.Config{
		PartnerCode: cfg.GetStr(constants.MomoPartnerCode, ""),
		AccessKey:   cfg.GetStr(constants.MomoAccessKey, ""),
		SecretKey:   cfg.GetStr(constants.MomoSecretKey, ""),
		ApiEndpoint: endpoint,
		IpnUrl:      cfg.GetStr(constants.MomoIpnUrl, ""),
		RedirectUrl: cfg.GetStr(constants.MomoRedirectUrl, ""),
	}, httpclient.NewHttpCaller(endpoint, httpClient, logger))

	return registry.Register(adapter)
}

func registerMpos(
	registry *itGateway.Registry,
	cfg config.ConfigService,
	httpClient *httpclientclient.HttpClient,
	logger logging.LoggerService,
) error {
	if !cfg.GetBool(constants.MposEnabled, false) {
		return nil
	}

	endpoint, err := requireEndpoint(cfg, constants.MposApiEndpoint, "MPOS")
	if err != nil {
		return err
	}

	secretKey := cfg.GetStr(constants.MposSecretKey, "")
	// The secret is used directly as an AES-128 key, so a wrong length is not a mis-typed
	// credential that fails to authenticate — it panics inside the cipher on the first payment.
	// The length is checked here, where it can still be reported as a configuration problem.
	if len(secretKey) != mpos.SecretKeyLength {
		return errors.Errorf(
			"paymentinvoice: %s must be exactly %d characters, got %d",
			constants.MposSecretKey, mpos.SecretKeyLength, len(secretKey))
	}

	adapter := mpos.NewAdapter(mpos.Config{
		MerchantId: cfg.GetStr(constants.MposMerchantId, ""),
		SecretKey:  secretKey,
	}, httpclient.NewHttpCaller(endpoint, httpClient, logger))

	return registry.Register(adapter)
}

func registerVietQr(
	registry *itGateway.Registry,
	cfg config.ConfigService,
	httpClient *httpclientclient.HttpClient,
	logger logging.LoggerService,
) error {
	if !cfg.GetBool(constants.VietQrEnabled, false) {
		return nil
	}

	endpoint, err := requireEndpoint(cfg, constants.VietQrApiEndpoint, "VIETQR")
	if err != nil {
		return err
	}

	// Only the outbound credentials belong to the adapter. VIETQR.INBOUND_* authenticate the
	// bank when it calls us, and are read by the webhook layer instead; conflating the two pairs
	// would have this deployment present the bank's own credentials back to it.
	adapter := vietqr.NewAdapter(vietqr.Config{
		Username:   cfg.GetStr(constants.VietQrUsername, ""),
		Password:   cfg.GetStr(constants.VietQrPassword, ""),
		SecretKey:  cfg.GetStr(constants.VietQrSecretKey, ""),
		BankCode:   cfg.GetStr(constants.VietQrBankCode, ""),
		BankNumber: cfg.GetStr(constants.VietQrBankNumber, ""),
		BankName:   cfg.GetStr(constants.VietQrBankName, ""),
	}, httpclient.NewHttpCaller(endpoint, httpClient, logger))

	return registry.Register(adapter)
}

// requireEndpoint reads a gateway's base URL, refusing an enabled gateway that has none.
//
// NewHttpCaller panics on an empty or malformed base URL, which during Init would take the whole
// application down with a message about a URL rather than about the gateway that lacks one.
func requireEndpoint(cfg config.ConfigService, name core.ConfigName, gateway string) (string, error) {
	endpoint := strings.TrimSpace(cfg.GetStr(name, ""))
	if endpoint == "" {
		return "", errors.Errorf(
			"paymentinvoice: gateway %s is enabled but %s is not set", gateway, name)
	}
	return endpoint, nil
}
