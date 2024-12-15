package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// TestPgContainer wraps a PostgreSQL test container and provides helper methods.
type TestPgContainer struct {
	instance *postgres.PostgresContainer
	conn     *sql.DB
	connOnce sync.Once
}

// GetDSN retrieves the PostgreSQL connection string for the container.
// It returns an error if the container is not initialized.
func (c *TestPgContainer) GetDSN(ctx context.Context) (string, error) {
	if c.instance == nil {
		return "", fmt.Errorf("container is not initialized")
	}

	connStr, err := c.instance.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", fmt.Errorf("failed to retrieve connection string: %w", err)
	}
	return connStr, nil
}

// GetConnection returns an active database connection to the PostgreSQL container.
// It ensures that the connection is established only once, even if called concurrently.
func (c *TestPgContainer) GetConnection(ctx context.Context) (*sql.DB, error) {
	var err error
	var conn *sql.DB

	c.connOnce.Do(func() {
		var connStr string
		connStr, err = c.GetDSN(ctx)
		if err != nil {
			return
		}

		conn, err = sql.Open("postgres", connStr)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get or open database connection: %w", err)
	}
	return conn, nil
}

// Stop terminates the PostgreSQL container and closes the connection if it's open.
func (c *TestPgContainer) Stop(ctx context.Context) error {
	if c.conn != nil {
		_ = c.conn.Close() // Close the database connection if it exists
	}
	if c.instance != nil {
		return c.instance.Terminate(ctx)
	}
	return nil
}

// PgContainerOpts contains the options for creating a PostgreSQL test container.
type PgContainerOpts struct {
	ContainerImage string
	DbName         string
	DbUser         string
	DbPass         string
	InitScripts    []string
	ConfigFile     string
}

// PgOptFunc is a function that modifies PgContainerOpts.
type PgOptFunc func(*PgContainerOpts)

// Option setters for PostgreSQL container options.

func WithPgContainerImage(image string) PgOptFunc {
	return func(o *PgContainerOpts) {
		o.ContainerImage = image
	}
}

func WithPgDbUser(user string) PgOptFunc {
	return func(o *PgContainerOpts) {
		o.DbUser = user
	}
}

func WithPgDbPassword(pass string) PgOptFunc {
	return func(o *PgContainerOpts) {
		o.DbPass = pass
	}
}

func WithPgDbName(name string) PgOptFunc {
	return func(o *PgContainerOpts) {
		o.DbName = name
	}
}

func WithPgInitScripts(scripts ...string) PgOptFunc {
	return func(o *PgContainerOpts) {
		o.InitScripts = scripts
	}
}

func WithPgConfigFile(file string) PgOptFunc {
	return func(o *PgContainerOpts) {
		o.ConfigFile = file
	}
}

// defaultPgContainerOpts returns the default PostgreSQL container options.
func defaultPgContainerOpts() PgContainerOpts {
	return PgContainerOpts{
		ContainerImage: "postgres:latest",
		DbName:         fmt.Sprintf("test_%d", time.Now().UnixNano()),
		DbUser:         "test",
		DbPass:         "test",
	}
}

// InitPgTestContainer initializes a new PostgreSQL test container with customizable options.
// It accepts a variadic list of PgOptFunc to modify the default options.
func InitPgTestContainer(ctx context.Context, opts ...PgOptFunc) (*TestPgContainer, error) {
	// Apply user-provided options to the default ones.
	containerOpts := defaultPgContainerOpts()
	for _, opt := range opts {
		opt(&containerOpts)
	}

	// Prepare the list of container options
	postgresOpts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase(containerOpts.DbName),
		postgres.WithUsername(containerOpts.DbUser),
		postgres.WithPassword(containerOpts.DbPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30 * time.Second)),
	}

	// Add optional configurations
	if len(containerOpts.InitScripts) > 0 {
		postgresOpts = append(postgresOpts, postgres.WithInitScripts(containerOpts.InitScripts...))
	}

	if containerOpts.ConfigFile != "" {
		postgresOpts = append(postgresOpts, postgres.WithConfigFile(containerOpts.ConfigFile))
	}

	// Start the PostgreSQL container
	postgresContainer, err := postgres.Run(
		ctx,
		containerOpts.ContainerImage,
		postgresOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize PostgreSQL container: %w", err)
	}

	return &TestPgContainer{
		instance: postgresContainer,
	}, nil
}
