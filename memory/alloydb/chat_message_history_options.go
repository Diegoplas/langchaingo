package alloydb

import (
	"errors"

	"github.com/tmc/langchaingo/internal/alloydbutil"
)

const (
	defaultSchemaName = "public"
)

// ChatMessageHistoryStoresOption is a function for creating chat message
// history with other than the default values.
type ChatMessageHistoryStoresOption func(c *ChatMessageHistory)

// WithSessionID sets the sessionID field for the ChatMessageHistory.
func WithSessionID(sessionID string) ChatMessageHistoryStoresOption {
	return func(c *ChatMessageHistory) {
		c.sessionID = sessionID
	}
}

// WithSchemaName sets the schemaName field for the ChatMessageHistory.
func WithSchemaName(schemaName string) ChatMessageHistoryStoresOption {
	return func(c *ChatMessageHistory) {
		c.schemaName = schemaName
	}
}

// WithOverwrite is an option for NewChatMessageHistory for
// allowing dangerous operations like SetMessages or Clear.
func WithOverwrite() ChatMessageHistoryStoresOption {
	return func(c *ChatMessageHistory) {
		c.overwrite = true
	}
}

// ApplyChatMessageHistoryOptions applies the given options to the
// ChatMessageHistory.
func ApplyChatMessageHistoryOptions(engine alloydbutil.PostgresEngine, tableName string, opts ...ChatMessageHistoryStoresOption) (ChatMessageHistory, error) {
	// Check for required values.
	if engine.Pool == nil {
		return ChatMessageHistory{}, errors.New("missing chat message history engine")
	}
	if tableName == "" {
		return ChatMessageHistory{}, errors.New("table name must be provided")
	}

	cmh := &ChatMessageHistory{
		engine:     engine,
		tableName:  tableName,
		schemaName: defaultSchemaName,
	}
	// Check for optional values.
	for _, opt := range opts {
		opt(cmh)
	}
	if cmh.sessionID == "" {
		return ChatMessageHistory{}, errors.New("session ID must be provided")
	}
	return *cmh, nil
}
