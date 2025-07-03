package test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type TestConfig struct{}

func (cfg *TestConfig) Init() error           { return nil }
func (cfg *TestConfig) GetAppVersion() string { return "test" }
func (cfg *TestConfig) GetStr(configName c.ConfigName, defaultVal ...any) string {
	switch configName {
	case c.EventBusRedisHost:
		return "localhost"
	case c.EventBusRedisPort:
		return "7379"
	case c.EventBusRedisPassword:
		return "nikki_password"
	default:
		if len(defaultVal) > 0 {
			if str, ok := defaultVal[0].(string); ok {
				return str
			}
		}
		return ""
	}
}
func (cfg *TestConfig) GetStrArr(configName c.ConfigName, defaultVal ...any) []string { return nil }
func (cfg *TestConfig) GetDuration(configName c.ConfigName, defaultVal ...any) time.Duration {
	return 0
}
func (cfg *TestConfig) GetBool(configName c.ConfigName, defaultVal ...any) bool     { return false }
func (cfg *TestConfig) GetUint(configName c.ConfigName, defaultVal ...any) uint     { return 0 }
func (cfg *TestConfig) GetUint64(configName c.ConfigName, defaultVal ...any) uint64 { return 0 }
func (cfg *TestConfig) GetInt(configName c.ConfigName, defaultVal ...any) int {
	switch configName {
	case c.EventBusRedisDB:
		return 0
	case c.EventRequestTimeoutSecs:
		return 5
	default:
		if len(defaultVal) > 0 {
			if val, ok := defaultVal[0].(int); ok {
				return val
			}
		}
		return 0
	}
}
func (cfg *TestConfig) GetInt32(configName c.ConfigName, defaultVal ...any) int32     { return 0 }
func (cfg *TestConfig) GetInt64(configName c.ConfigName, defaultVal ...any) int64     { return 0 }
func (cfg *TestConfig) GetFloat32(configName c.ConfigName, defaultVal ...any) float32 { return 0 }

type TestLogger struct {
	slogger *slog.Logger
}

func NewTestLogger() *TestLogger {
	// Create a slog logger that discards output for testing
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	return &TestLogger{
		slogger: logger,
	}
}

func (l *TestLogger) Level() logging.Level              { return logging.LevelInfo }
func (l *TestLogger) SetLevel(lvl logging.Level)        {}
func (l *TestLogger) InnerLogger() any                  { return l.slogger }
func (l *TestLogger) Debug(message string, data any)    {}
func (l *TestLogger) Debugf(format string, args ...any) {}
func (l *TestLogger) Info(message string, data any)     {}
func (l *TestLogger) Infof(format string, args ...any)  {}
func (l *TestLogger) Warn(message string, data any)     {}
func (l *TestLogger) Warnf(format string, args ...any)  {}
func (l *TestLogger) Error(message string, err error)   {}
func (l *TestLogger) Errorf(format string, args ...any) {}

type TestHandler struct {
	mu           sync.RWMutex
	messageCount int
	messages     []string
}

func (h *TestHandler) Handle(ctx context.Context, msg *message.Message) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	correlationId := msg.Metadata.Get("correlation_id")
	h.messageCount++
	h.messages = append(h.messages, correlationId)

	fmt.Printf("[%d] Received message: %s (at %s)\n",
		h.messageCount, correlationId, time.Now().Format("15:04:05"))
	return nil
}

func (h *TestHandler) NewEvent() any { return nil }

func (h *TestHandler) GetStats() (int, []string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return copy of messages
	messagesCopy := make([]string, len(h.messages))
	copy(messagesCopy, h.messages)

	return h.messageCount, messagesCopy
}

func TestRedisEventBusSubscribe(t *testing.T) {
	eventBus, err := event.NewRedisEventBus(event.EventBusParams{
		Config: &TestConfig{},
		Logger: NewTestLogger(),
	})
	if err != nil {
		t.Fatalf("failed to create event bus: %v", err)
	}
	defer eventBus.Close()

	ctx := context.Background()

	// Create a subscription request
	result := &message.Payload{}
	eventRequest := event.NewEventRequest(
		"",
		"user.deleted.done",
		"user.deleted.reply",
		nil,
	)

	requestChan, err := eventBus.SubscribeRequest(ctx, *eventRequest, result)
	if err != nil {
		t.Fatalf("failed to subscribe to event: %v", err)
	}

	// Listen for messages in a goroutine
	go func() {
		for {
			select {
			case request := <-requestChan:
				if request != nil {
					fmt.Printf("Received request: %v (at %s)\n",
						request, time.Now().Format("15:04:05"))
				}
			case <-time.After(10 * time.Minute): // Timeout after 10 minutes
				fmt.Println("Subscription timed out after 10 minutes")
				return
			}
		}
	}()

	// Wait for a few seconds to test subscription
	time.Sleep(20 * time.Minute)

	fmt.Printf("Subscription test completed\n")
}

func TestRedisEventBusPublish(t *testing.T) {
	eventBus, err := event.NewRedisEventBus(event.EventBusParams{
		Config: &TestConfig{},
		Logger: NewTestLogger(),
	})
	if err != nil {
		t.Fatalf("failed to create event bus: %v", err)
	}
	defer eventBus.Close()

	ctx := context.Background()

	// Publish a few test messages
	for messageCount := 1; messageCount <= 3; messageCount++ {
		correlationId := fmt.Sprintf("test-message-%d", messageCount)

		payload := []byte(fmt.Sprintf(`{"message": "Message payload for %d"}`, messageCount))
		eventRequest := event.NewEventRequest(
			correlationId,
			"user.deleted.done",
			"user.deleted.reply",
			&message.Message{
				Payload: payload,
			},
		)

		fmt.Printf("[%d] Publishing message: %s (at %s)\n",
			messageCount, correlationId, time.Now().Format("15:04:05"))

		err = eventBus.PublishRequest(ctx, *eventRequest)
		if err != nil {
			fmt.Printf("Failed to publish message %d: %v\n", messageCount, err)
		} else {
			fmt.Printf("[%d] Published successfully: %s\n", messageCount, correlationId)
		}

		time.Sleep(1 * time.Second) // Wait 1 second between messages
	}
}

// func TestRedisEventBusPublishWaitReply(t *testing.T) {
// 	eventBus, err := event.NewRedisEventBus(event.EventBusParams{
// 		Config: &TestConfig{},
// 		Logger: NewTestLogger(),
// 	})
// 	if err != nil {
// 		t.Fatalf("failed to create event bus: %v", err)
// 	}
// 	defer eventBus.Close()

// 	ctx := context.Background()

// 	// Publish a few messages with wait for reply
// 	for messageCount := 1; messageCount <= 3; messageCount++ {
// 		correlationId := fmt.Sprintf("wait-reply-message-%d", messageCount)

// 		payload := []byte(fmt.Sprintf(`{"message": "Wait reply message payload for %d"}`, messageCount))
// 		msg := message.NewMessage(correlationId, payload)

// 		eventRequest := event.NewEventRequest(
// 			correlationId,
// 			"identity.organization.deleted",
// 			"identity.organization.deleted.reply",
// 			msg,
// 		)

// 		fmt.Printf("[%d] Publishing message with wait reply: %s (at %s)\n",
// 			messageCount, correlationId, time.Now().Format("15:04:05"))

// 		// Create a timeout context for each request
// 		timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)

// 		var result interface{}
// 		reply, err := eventBus.PublishRequestWaitReply(timeoutCtx, *eventRequest, &result)

// 		if err != nil {
// 			fmt.Printf("Failed to publish wait reply message %d: %v\n", messageCount, err)
// 		} else {
// 			fmt.Printf("[%d] Published with reply successfully: %s, Result: %v\n",
// 				messageCount, correlationId, reply)
// 		}

// 		cancel()                    // Clean up timeout context
// 		time.Sleep(1 * time.Second) // Wait 1 second between messages
// 	}
// }
