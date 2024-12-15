package checks

import (
	"context"
	"database/sql"
	"time"

	"github.com/brpaz/lib-go/health"
)

// DBCheck is a health check that verifies the availability of a database connection
type DBCheck struct {
	name string
	db   *sql.DB
}

// NewDBCheck creates a new DBCheck instance with the provided parameters
func NewDBCheck(name string, db *sql.DB) *DBCheck {
	return &DBCheck{
		name: name,
		db:   db,
	}
}

// GetName returns the name of the check
func (d *DBCheck) GetName() string {
	return d.name
}

// Check verifies the availability of the database connection
func (d *DBCheck) Check(ctx context.Context) health.CheckResult {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := d.db.ExecContext(ctx, "SELECT 1")
	if err != nil {
		return health.CheckResult{
			Status: health.StatusFail,
			Error:  err,
		}
	}

	return health.CheckResult{
		Status: health.StatusPass,
	}
}
