package alloydb

import (
	"errors"
)

const (
	defaultSchemaName = "public"
)

// ChatMessageHistoryStoresOption is a function for creating chat message
// history with other than the default values.
type ChatMessageHistoryStoresOption func(c *chatMessageHistory)

// WithCMHEngine sets the Engine field for the chatMessageHistory.
func WithCMHEngine(engine PostgresEngine) ChatMessageHistoryStoresOption {
	return func(c *chatMessageHistory) {
		c.engine = engine
	}
}

// WithSessionID sets the sessionID field for the chatMessageHistory.
func WithSessionID(sessionID string) ChatMessageHistoryStoresOption {
	return func(c *chatMessageHistory) {
		c.sessionID = sessionID
	}
}

// WithTableName sets the tableName field for the chatMessageHistory.
func WithTableName(tableName string) ChatMessageHistoryStoresOption {
	return func(c *chatMessageHistory) {
		c.tableName = tableName
	}
}

// WithSchemaName sets the schemaName field for the chatMessageHistory.
func WithSchemaName(schemaName string) ChatMessageHistoryStoresOption {
	return func(c *chatMessageHistory) {
		c.schemaName = schemaName
	}
}

// ApplyChatMessageHistoryOptions applies the given options to the
// chatMessageHistory.
func ApplyChatMessageHistoryOptions(opts ...ChatMessageHistoryStoresOption) (chatMessageHistory, error) {
	cmh := &chatMessageHistory{
		schemaName: defaultSchemaName,
	}
	for _, opt := range opts {
		opt(cmh)
	}
	if cmh.engine.pool == nil {
		return chatMessageHistory{}, errors.New("missing chat message history engine")
	}
	if cmh.tableName == "" {
		return chatMessageHistory{}, errors.New("table name must be provided")
	}
	if cmh.sessionID == "" {
		return chatMessageHistory{}, errors.New("session ID must be provided")
	}
	return *cmh, nil
}
