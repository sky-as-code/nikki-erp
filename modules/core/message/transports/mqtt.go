package transports

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttclient "github.com/sky-as-code/nikki-erp/modules/core/infra/mqtt"
	"go.uber.org/dig"
)

type MqttTransport struct {
	Transport *MessageTransport `name:"mqtt"`
	dig.Out
}

type MqttStream struct {
	client mqttclient.MqttClient
}

func NewMqttTransport(mqttClient mqttclient.MqttClient) (MqttTransport, error) {
	stream := &MqttStream{
		client: mqttClient,
	}

	return MqttTransport{
		Transport: &MessageTransport{
			Publisher:  stream,
			Subscriber: stream,
		},
	}, nil
}

type Message struct {
	UUID     string
	Metadata map[string]string
	Payload  json.RawMessage
}

func (this *MqttStream) Publish(topic string, messages ...*message.Message) error {
	publishError := make(chan error, 1)
	for _, msg := range messages {
		go func() {
			data := Message{
				UUID:     msg.UUID,
				Metadata: msg.Metadata,
				Payload:  json.RawMessage(msg.Payload),
			}

			b, err := json.Marshal(data)
			if err != nil {
				publishError <- err
			}

			publishError <- this.client.Publish(topic, 2, false, b).Error()
		}()

		select {
		case <-msg.Context().Done():
			return msg.Context().Err()
		case err := <-publishError:
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (this *MqttStream) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	ch := make(chan *message.Message, 10000)
	err := this.client.Subscribe(topic, 2, func(_ mqtt.Client, m mqtt.Message) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		msg := &Message{}
		json.Unmarshal(m.Payload(), msg)

		watermillMsg := message.NewMessage(msg.UUID, []byte(msg.Payload))
		watermillMsg.Metadata = msg.Metadata

		select {
		case <-ctx.Done():
			return
		case ch <- watermillMsg:
		}
	}).Error()
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		this.client.Unsubscribe(topic)
		close(ch)
	}()

	return ch, nil
}

func (this *MqttStream) Close() error {
	return nil
}
