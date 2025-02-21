package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO :: https://github.com/GoogleCloudPlatform/alloydb-go-connector
// connect with sql
// gcloud auth application-default login
// replace github.com/tmc/langchaingo => github.com/averikitsch/langchaingo v0.0.0-d0cc607989935cf3450965d1adfd40ae539e2be5

// testConnection is a simple function to ping the database to verify the connection
func testConnection(ctx context.Context, conn *pgxpool.Pool) error {
	// You can perform a simple query to test the connection
	var result string
	err := conn.QueryRow(ctx, "SELECT current_database()").Scan(&result)
	if err != nil {
		return fmt.Errorf("error testing connection: %w", err)
	}
	fmt.Println("Query result:", result)
	return nil
}

func main() {
	// Set up context
	ctx := context.Background()

	// Call NewPostgresEngine to initialize the database connection
	pgEngine, err := NewAlloyDBEngine(ctx)
	if err != nil {
		fmt.Println("error creating engine: %w", err)
		return
	}
	// Make sure to close the connection when done
	defer func() {
		pgEngine.Close()
	}()
	err = testAlloyDBCMHMethods(ctx, pgEngine)
	if err != nil {
		fmt.Println("error creating engine: %w", err)
		return
	}

	// Call cloudSQLEngine to initialize the database connection
	cloudSQLEngine, err := NewCloudSQLEngine(ctx)
	if err != nil {
		fmt.Println("error creating AlloyDB CMH: %w", err)
		return
	}
	err = testCloudSQLCMHMethods(ctx, cloudSQLEngine)
	if err != nil {
		fmt.Println("error creating CloudSQL CMH: %w", err)
		return
	}
}

//https://medium.com/google-cloud/using-alloydb-connector-for-automatic-iam-authentication-service-account-ec29c4ee5d2b
//https://cloud.google.com/alloydb/docs/connect-language-connectors#go-pgx
