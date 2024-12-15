// Package migrator provides an abstraction for managing database migrations.
// It uses Goose under the hood but can be swapped out for other migration tools.
// Example usage:
//  // package main
//  // Get a database connection
//  dbConn, err := sql.Open("postgres", "user=postgres password=postgres dbname=postgres sslmode=disable")
//
// 	// Create a new Goose migrator
//	m, err := migrator.NewGooseMigrator(
//		migrator.WithGooseDB(dbConn),
//		migrator.WithGooseMigrationsDir("migrations"),
//		migrator.WithGooseDialect("postgres"),
//		migrator.WithGooseMigrationsType("sql"),
//	)

//	if err != nil {
//		return err
//	}
//
// // Run the migrations
//
//	if err := m.Up(ctx); err != nil {
//		return err
//	}
package migrator
