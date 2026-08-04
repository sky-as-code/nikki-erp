package mqttclient

import (
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const Qos = 2

type MqttClient interface {
	mqtt.Client
}

func NewVerneMq(cfg config.ConfigService, logger logging.LoggerService) MqttClient {
	uri := fmt.Sprintf("%s://%s:%s",
		cfg.GetStr(constants.PubSubVerneScheme, "tcp"),
		cfg.GetStr(constants.PubSubVerneHost, "localhost"),
		cfg.GetStr(constants.PubSubVernePort, "1883"),
	)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(uri)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logging.Logger().Error("[MQTT] Connection Lost", err)
	})
	opts.SetClientID(uuid.NewString())
	opts.SetUsername(cfg.GetStr(constants.PubSubVerneUsername))
	opts.SetPassword(cfg.GetStr(constants.PubSubVernePassword))
	opts.WillQos = Qos
	// opts.SetConnectRetry(true)
	// opts.SetConnectRetryInterval(2 * time.Second)
	// opts.SetConnectTimeout(10 * time.Second)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(time.Second * 10) {
		panic("mqtt connect time out")
	}

	if token.Error() != nil {
		panic(token.Error())
	}

	return client
}
