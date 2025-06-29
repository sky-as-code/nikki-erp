package test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	c "github.com/sky-as-code/nikki-erp/modules/core/constants"
	"github.com/sky-as-code/nikki-erp/modules/core/event"
	"github.com/sky-as-code/nikki-erp/modules/core/logging"
)

type TestConfig struct{}

func (cfg *TestConfig) Init() error           { return nil }
func (cfg *TestConfig) GetAppVersion() string { return "test" }
func (cfg *TestConfig) GetStr(configName c.ConfigName, defaultVal ...any) string {
	switch configName {
	case "EVENT_BUS_REDIS_HOST":
		return "localhost"
	case "EVENT_BUS_REDIS_PORT":
		return "7379"
	case "EVENT_BUS_REDIS_PASSWORD":
		return "nikki_password"
	default:
		return ""
	}
}
func (cfg *TestConfig) GetStrArr(configName c.ConfigName, defaultVal ...any) []string { return nil }
func (cfg *TestConfig) GetDuration(configName c.ConfigName, defaultVal ...any) time.Duration {
	return 0
}
func (cfg *TestConfig) GetBool(configName c.ConfigName, defaultVal ...any) bool       { return false }
func (cfg *TestConfig) GetUint(configName c.ConfigName, defaultVal ...any) uint       { return 0 }
func (cfg *TestConfig) GetUint64(configName c.ConfigName, defaultVal ...any) uint64   { return 0 }
func (cfg *TestConfig) GetInt(configName c.ConfigName, defaultVal ...any) int         { return 0 }
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

func (h *TestHandler) Handle(ctx context.Context, packet *event.EventPacket) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	correlationId := packet.CorrelationId()
	h.messageCount++
	h.messages = append(h.messages, correlationId)

	fmt.Printf("📨 [%d] Received message: %s (at %s)\n",
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
	handler := &TestHandler{}

	eventBus, err := event.NewRedisEventBus(event.EventBusParams{
		Config: &TestConfig{},
		Logger: NewTestLogger(),
	})
	if err != nil {
		t.Fatalf("failed to create event bus: %v", err)
	}
	defer eventBus.Close()

	err = eventBus.Subscribe(context.Background(), "user.deleted.done", handler)
	if err != nil {
		t.Fatalf("failed to subscribe to event: %v", err)
	}

	currentCountMessages := 0
	for {
		count, messages := handler.GetStats()
		if count != currentCountMessages {
			currentCountMessages = count
			fmt.Printf("Current message count: %d, Messages: %v\n", count, messages)
		}
	}
}
