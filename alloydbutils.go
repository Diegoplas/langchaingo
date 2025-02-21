package main

import (
	"context"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/internal/alloydbutil"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory/alloydb"
)

// Creates alloyDB engine, tests the connection and defers
func NewAlloyDBEngine(ctx context.Context) (*alloydbutil.PostgresEngine, error) {
	// Call NewPostgresEngine to initialize the database connection
	pgEngine, err := alloydbutil.NewPostgresEngine(ctx,
		alloydbutil.WithUser(""),
		alloydbutil.WithPassword(""),
		alloydbutil.WithDatabase(""),
		alloydbutil.WithAlloyDBInstance("", "", "", ""),
	)
	if err != nil {
		return nil, fmt.Errorf("Error creating PostgresEngine: %s", err)
	}

	// Test the connection by pinging the database (this can be any query or check)
	if err := testConnection(ctx, pgEngine.Pool); err != nil {
		log.Fatalf("Connection test failed: %v", err)
	} else {
		fmt.Println("Successfully connected to the database!")
	}
	return pgEngine, nil
}

func testAlloyDBCMHMethods(ctx context.Context, pgEngine *alloydbutil.PostgresEngine) error {
	fmt.Println("<::: TESTING ALLOYDB CHAT MESSAGE HISTORY METHODS :::>")
	// Call NewChatMessageHistory to initialize a chat message history
	cmh, err := alloydb.NewChatMessageHistory(ctx, *pgEngine, "testtable", "testSessionID",
		alloydb.WithSchemaName("cmh"),
		//WithOverwrite(),
	)
	if err != nil {
		return fmt.Errorf("Error creating chat message history: %s", err)
	}

	fmt.Println(" :: CMH :: ", cmh)

	err = cmh.Clear(ctx)
	if err != nil {
		return fmt.Errorf("Error clearing messages: %s", err)
	}

	msgs, err := cmh.Messages(ctx)
	if err != nil {
		return fmt.Errorf("Error getting messages: %s", err)
	}
	for _, msg := range msgs {
		fmt.Println(msg)
	}
	fmt.Println("--------------------------------------------")

	aiMessage1 := llms.AIChatMessage{Content: "first AI message from single addMessage"}
	hMessage1 := llms.AIChatMessage{Content: "first HUMASN message from single addMessage"}

	err = cmh.AddUserMessage(ctx, string(hMessage1.GetContent()))
	if err != nil {
		return fmt.Errorf("Error addMessage: %s", err)
	}

	err = cmh.AddUserMessage(ctx, string(aiMessage1.GetContent()))
	if err != nil {
		return fmt.Errorf("Error addMessage: %s", err)
	}

	msgs, err = cmh.Messages(ctx)
	if err != nil {
		return fmt.Errorf("Error getting messages 2: %s", err)
	}
	for _, msg := range msgs {
		fmt.Println(msg)
	}
	fmt.Println("--------------------------------------------")

	manyMessages := []llms.ChatMessage{
		llms.AIChatMessage{Content: "first AI message from single addMessage"},
		llms.AIChatMessage{Content: "second AI message from single addMessage"},
		llms.AIChatMessage{Content: "first HUMASN message from single addMessage"},
	}

	err = cmh.AddMessages(ctx, manyMessages)
	if err != nil {
		return fmt.Errorf("Error adding Multiple Messages: %s", err)
	}
	msgs, err = cmh.Messages(ctx)
	if err != nil {
		return fmt.Errorf("Error getting many messages: %s", err)
	}
	for _, msg := range msgs {
		fmt.Println(msg)
	}
	fmt.Println("--------------------------------------------")

	lastMessage := []llms.ChatMessage{
		llms.AIChatMessage{Content: "last message warning!"},
	}

	err = cmh.SetMessages(ctx, lastMessage)
	if err != nil {
		return fmt.Errorf("Error setting Message: %s", err)
	}

	msgs, err = cmh.Messages(ctx)
	if err != nil {
		return fmt.Errorf("Error getting many messages: %s", err)
	}
	for _, msg := range msgs {
		fmt.Println(msg)
	}
	fmt.Println("--------------------------------------------")
	return nil
}
