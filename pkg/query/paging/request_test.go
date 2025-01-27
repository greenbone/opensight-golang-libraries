// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package paging

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestAddPagingRequest(t *testing.T) {
	var (
		err     error
		sqlMock sqlmock.Sqlmock
		sqlDB   *sql.DB
		gormDB  *gorm.DB
	)

	setup := func(t *testing.T) {
		sqlDB, sqlMock, err = sqlmock.New()
		require.NoError(t, err)

		gormDB, err = gorm.Open(postgres.New(postgres.Config{
			Conn: sqlDB,
		}), &gorm.Config{})
		require.NoError(t, err)
	}

	exactMatchRegexp := func(sqlString string) string {
		return "^" + regexp.QuoteMeta(sqlString) + "$"
	}

	type testObject struct {
		TheString  string
		TheInteger int
	}

	t.Run("shouldBuildRequestWithoutSorting", func(t *testing.T) {
		setup(t)
		request := &Request{
			PageIndex: 2,
			PageSize:  10,
		}

		sqlMock.ExpectQuery(exactMatchRegexp(`SELECT * FROM "test_objects" WHERE "TheString" = 123 LIMIT $1 OFFSET $2`)).WithArgs(
			10, 20).WillReturnRows(sqlmock.NewRows([]string{}))

		gormDB = gormDB.Where(`"TheString" = 123`)
		gormDB = AddRequest(gormDB, request)

		gormDB.Find(&testObject{})
		require.NoError(t, sqlMock.ExpectationsWereMet())
	})
}
