package alloydb

import (
	"context"
	"fmt"
	"time"
)

type chatMessageHistory struct {
	engine     PostgresEngine
	sessionID  string
	tableName  string
	schemaName string
}

type baseMessage interface {
	getData() string
	getType() string
}

// humanMessage represents a message sent by a human.
type humanMessage struct {
	data string
}

// aIMessage represents a message sent by an AI.
type aIMessage struct {
	data string
}

// getData gets the data of a HumanMessage.
func (h humanMessage) getData() string {
	return h.data
}

// getData gets the data of an AIMessage.
func (a aIMessage) getData() string {
	return a.data
}

// GetType returns "human" for HumanMessage
func (h humanMessage) getType() string {
	return "human"
}

// GetType returns "ai" for AIMessage
func (a aIMessage) getType() string {
	return "ai"
}

// NewchatMessageHistory creates a new NewchatMessageHistory with options.
func NewchatMessageHistory(ctx context.Context, opts ...ChatMessageHistoryStoresOption) (chatMessageHistory, error) {
	cmh, err := ApplyChatMessageHistoryOptions(opts...)
	if err != nil {
		return chatMessageHistory{}, err
	}
	err = cmh.initChatHistoryTable(ctx)
	if err != nil {
		return chatMessageHistory{}, err
	}
	return cmh, nil
}

// initChatHistoryTable creates a Cloud SQL table to store chat history
// if it do not exist.
func (c *chatMessageHistory) initChatHistoryTable(ctx context.Context) error {
	// Create schema if necessary
	createSchemaQuery := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s";`, c.schemaName)
	_, err := c.engine.pool.Exec(ctx, createSchemaQuery)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	// Create table if necessary
	createTableQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"."%s" (
			id SERIAL PRIMARY KEY,
			session_id TEXT NOT NULL,
			data TEXT NOT NULL,
			type TEXT NOT NULL,
			timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
		);`, c.schemaName, c.tableName)
	_, err = c.engine.pool.Exec(ctx, createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create the table: %w", err)
	}
	return nil
}

// AddMessage inserts a new message into AlloyDB.
func (c *chatMessageHistory) AddMessage(message baseMessage) error {
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (session_id, data, type) VALUES ($1, $2, $3)`,
		c.schemaName, c.tableName)

	_, err := c.engine.pool.Exec(context.Background(), query, c.sessionID, message.getData(), message.getType())
	if err != nil {
		return fmt.Errorf("failed to add message to database: %w", err)
	}
	return nil
}

// AddMessages inserts new messages into AlloyDB.
func (c *chatMessageHistory) AddMessages(messages []baseMessage) error {
	for _, message := range messages {
		err := c.AddMessage(message)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetMessages retrieves all messages for a given session_id from AlloyDB.
func (c *chatMessageHistory) GetMessages() ([]baseMessage, error) {
	query := fmt.Sprintf(`SELECT id, session_id, data, type, timestamp FROM "%s"."%s" WHERE session_id = $1 ORDER BY id`,
		c.schemaName, c.tableName)

	rows, err := c.engine.pool.Query(context.Background(), query, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}
	defer rows.Close()

	var messages []baseMessage
	for rows.Next() {
		var id int
		var sessionID, data, messageType string
		var timestamp time.Time
		if err := rows.Scan(&id, &sessionID, &data, &messageType, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
	}

	return messages, nil
}

// Clear session memory from AlloyDB.
func (c *chatMessageHistory) Clear() error {
	query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE session_id = $1`,
		c.schemaName, c.tableName)

	_, err := c.engine.pool.Exec(context.Background(), query, c.sessionID)
	if err != nil {
		return fmt.Errorf("failed to clear session %s: %w", c.sessionID, err)
	}
	return nil
}
