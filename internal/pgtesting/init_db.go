// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package pgtesting

import (
	"database/sql"
	"embed"
	"testing"

	_ "github.com/lib/pq" // register the "postgres" driver
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
)

const driverName = "postgres"

// NewDB is a helper that returns an open connection to a unique and isolated
// test database, fully migrated and ready to query.
func NewDB(t *testing.T, migrationsFS embed.FS, migrationDir string) *sql.DB {
	t.Parallel() // each test has its own isolated database
	t.Helper()
	conf := pgtestdb.Config{ // must match the deployment in `compose.yml`
		DriverName: driverName,
		User:       "postgres",
		Password:   "password",
		Host:       "localhost",
		Port:       "5532",
		Options:    "sslmode=disable",
	}

	migrator := golangmigrator.New(
		migrationDir,
		golangmigrator.WithFS(migrationsFS),
	)
	db := pgtestdb.New(t, conf, migrator)

	return db
}
