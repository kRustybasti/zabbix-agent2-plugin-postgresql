/*
** Copyright (C) 2001-2025 Zabbix SIA
**
** This program is free software: you can redistribute it and/or modify it under the terms of
** the GNU Affero General Public License as published by the Free Software Foundation, version 3.
**
** This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
** without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
** See the GNU Affero General Public License for more details.
**
** You should have received a copy of the GNU Affero General Public License along with this program.
** If not, see <https://www.gnu.org/licenses/>.
**/

package plugin

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func Test_versionHandler(t *testing.T) {
	type mock struct {
		row *sqlmock.Rows
		err error
	}

	tests := []struct {
		name    string
		mock    mock
		want    any
		wantErr bool
	}{
		{
			"+valid",
			mock{
				row: sqlmock.NewRows([]string{"version"}).AddRow("postgres 69"),
			},
			"postgres 69",
			false,
		},
		{
			"-queryErr",
			mock{
				row: sqlmock.NewRows([]string{"version"}).AddRow("postgres 69"),
				err: errors.New("query err"),
			},
			nil,
			true,
		},
		{
			"-noRows",
			mock{
				row: sqlmock.NewRows([]string{"version"}),
			},
			nil,
			true,
		},
		{
			"-scanErr",
			mock{row: sqlmock.NewRows([]string{"version"}).AddRow(nil)},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("failed to create sql mock: %s", err.Error())
			}

			defer db.Close()

			mock.ExpectQuery(`^SELECT version\(\);$`).
				WillReturnRows(tt.mock.row).
				WillReturnError(tt.mock.err)

			got, err := versionHandler(
				context.Background(), &PGConn{client: db}, "", nil,
			)
			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"versionHandler() error = %v, wantErr %v", err, tt.wantErr,
				)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("versionHandler() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf(
					"versionHandler() sql mock expectations where not met: %s",
					err.Error(),
				)
			}
		})
	}
}
