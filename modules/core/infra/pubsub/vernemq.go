package pubsub

import (
	"context"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/sky-as-code/nikki-erp/modules/core/config"
	"github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

const Qos = 2

type mqttImpl struct {
	logger logging.LoggerService
	client mqtt.Client
}

func NewVerneMq(cfg config.ConfigService, logger logging.LoggerService) PubSubOut {
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
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetConnectTimeout(30 * time.Second)
	client := mqtt.NewClient(opts)
	if err := client.Connect().Error(); err != nil {
		panic(err)
	}

	impl := &mqttImpl{
		client: client,
		logger: logger,
	}

	return PubSubOut{
		PubSub:    impl,
		Publisher: impl,
		Subcriber: impl,
	}
}

func (this *mqttImpl) Publish(ctx context.Context, topic string, message []byte) error {
	this.logger.Debug("Publishing message to topic", logging.Attr{"topic": topic, "message": message})
	return this.client.Publish(topic, Qos, false, message).Error()
}

func (this *mqttImpl) Subscribe(ctx context.Context, topic string) (<-chan []byte, error) {
	ch := make(chan []byte, 1000)
	token := this.client.Subscribe(topic, Qos, func(c mqtt.Client, m mqtt.Message) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case ch <- m.Payload():
		default:
		}
	})

	if err := token.Error(); err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		this.client.Unsubscribe(topic)
		close(ch)
	}()

	return ch, nil
}
