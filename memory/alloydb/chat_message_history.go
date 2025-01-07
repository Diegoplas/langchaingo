package alloydb

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tmc/langchaingo/internal/alloydbutil"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
)

type ChatMessageHistory struct {
	engine     alloydbutil.PostgresEngine
	sessionID  string
	tableName  string
	schemaName string
	overwrite  bool
}

var _ schema.ChatMessageHistory = &ChatMessageHistory{}

// NewChatMessageHistory creates a new NewChatMessageHistory with options.
func NewChatMessageHistory(ctx context.Context, engine alloydbutil.PostgresEngine, tableName string, opts ...ChatMessageHistoryStoresOption) (ChatMessageHistory, error) {
	cmh, err := ApplyChatMessageHistoryOptions(engine, tableName, opts...)
	if err != nil {
		return ChatMessageHistory{}, err
	}
	err = cmh.validateTable(ctx)
	if err != nil {
		return ChatMessageHistory{}, err
	}
	return cmh, nil
}

// validateTable validates if a table with a specific schema exist and it
// contains the required columns.
func (c *ChatMessageHistory) validateTable(ctx context.Context) error {
	tableExistsQuery := fmt.Sprintf(`SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = '%s' AND table_name = '%s');`, // TODO :: Are table/schema names which require case-sensitive?
		c.schemaName, c.tableName)
	var exists bool
	err := c.engine.Pool.QueryRow(ctx, tableExistsQuery).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error validating table %s: %w", c.tableName, err)
	}
	if !exists {
		return fmt.Errorf("table '%s' does not exist in schema '%s'", c.tableName, c.schemaName)
	}

	requiredColumns := []string{"id", "session_id", "data", "type"} // TODO :: Should the field be data or content?
	for _, reqColumn := range requiredColumns {
		columnExistsQuery := fmt.Sprintf(`SELECT EXISTS (
			SELECT 1 FROM information_schema.columns 
			WHERE table_schema = '%s' AND table_name = '%s' AND column_name = '%s'
		);`, c.schemaName, c.tableName, reqColumn)
		var columnExists bool
		err := c.engine.Pool.QueryRow(ctx, columnExistsQuery).Scan(&columnExists)
		if err != nil {
			return fmt.Errorf("error scanning columns from table %s: %w", c.tableName, err)
		}
		if !columnExists {
			return fmt.Errorf("column '%s' is missing in table '%s'. Expected columns: %v", reqColumn, c.tableName, requiredColumns)
		}
	}
	return nil
}

// addMessage inserts a new message into AlloyDB.
func (c *ChatMessageHistory) addMessage(_ context.Context, content string, messageType llms.ChatMessageType) error {
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (session_id, data, type) VALUES ($1, $2, $3)`,
		c.schemaName, c.tableName)

	_, err := c.engine.Pool.Exec(context.Background(), query, c.sessionID, content, messageType)
	if err != nil {
		return fmt.Errorf("failed to add message to database: %w", err)
	}
	return nil
}

// AddMessage adds a message to the chat message history.
func (c *ChatMessageHistory) AddMessage(ctx context.Context, message llms.ChatMessage) error {
	return c.addMessage(ctx, message.GetContent(), message.GetType())
}

// AddAIMessage adds an AIMessage to the chat message history.
func (c *ChatMessageHistory) AddAIMessage(ctx context.Context, content string) error {
	return c.addMessage(ctx, content, llms.ChatMessageTypeAI)
}

// AddUserMessage adds a user to the chat message history.
func (c *ChatMessageHistory) AddUserMessage(ctx context.Context, content string) error {
	return c.addMessage(ctx, content, llms.ChatMessageTypeHuman)
}

// Clear resets messages.
func (c *ChatMessageHistory) Clear(_ context.Context) error {
	if !c.overwrite {
		return nil
	}
	query := fmt.Sprintf(`DELETE FROM "%s"."%s" WHERE session_id = $1`,
		c.schemaName, c.tableName)

	_, err := c.engine.Pool.Exec(context.Background(), query, c.sessionID)
	if err != nil {
		return fmt.Errorf("failed to clear session %s: %w", c.sessionID, err)
	}
	return err
}

// AddMessages inserts new messages into AlloyDB.
func (c *ChatMessageHistory) AddMessages(ctx context.Context, messages []llms.ChatMessage) error {
	b := &pgx.Batch{}
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (session_id, data, type) VALUES ($1, $2, $3)`,
		c.schemaName, c.tableName)

	for _, message := range messages {
		b.Queue(query, c.sessionID, message.GetContent(), message.GetType())
	}
	return c.engine.Pool.SendBatch(ctx, b).Close()
}

// Messages returns all messages for a given session_id from AlloyDB.
func (c *ChatMessageHistory) Messages(_ context.Context) ([]llms.ChatMessage, error) {
	query := fmt.Sprintf(`SELECT id, session_id, data, type, timestamp FROM "%s"."%s" WHERE session_id = $1 ORDER BY id`,
		c.schemaName, c.tableName)

	rows, err := c.engine.Pool.Query(context.Background(), query, c.sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}
	defer rows.Close()

	var messages []llms.ChatMessage
	for rows.Next() {
		var id int
		var sessionID, data, messageType string
		var timestamp time.Time
		if err := rows.Scan(&id, &sessionID, &data, &messageType, &timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		switch messageType {
		case string(llms.ChatMessageTypeAI):
			messages = append(messages, llms.AIChatMessage{Content: data}) // TODO :: Should types be added here too?
		case string(llms.ChatMessageTypeHuman):
			messages = append(messages, llms.HumanChatMessage{Content: data})
		case string(llms.ChatMessageTypeSystem):
			messages = append(messages, llms.SystemChatMessage{Content: data})
		default:
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// SetMessages resets chat history and bulk insert new messages into it.
func (c *ChatMessageHistory) SetMessages(ctx context.Context, messages []llms.ChatMessage) error {
	if !c.overwrite {
		return nil
	}
	err := c.Clear(ctx)
	if err != nil {
		return err
	}

	b := &pgx.Batch{}
	query := fmt.Sprintf(`INSERT INTO "%s"."%s" (session_id, data, type) VALUES ($1, $2, $3)`,
		c.schemaName, c.tableName)

	for _, message := range messages {
		b.Queue(query, c.sessionID, message.GetContent(), message.GetType())
	}
	return c.engine.Pool.SendBatch(ctx, b).Close()
}
