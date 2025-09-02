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

	"github.com/jackc/pgx/v4"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/zbxerr"
)

// versionHandler queries the version of the PostgreSQL server returns string
// response.
func versionHandler(
	ctx context.Context,
	conn PostgresClient,
	_ string, _ map[string]string, _ ...string,
) (any, error) {
	var version string

	row, err := conn.QueryRow(ctx, `SELECT version();`)
	if err != nil {
		return nil, errs.Wrap(zbxerr.ErrorCannotFetchData, err.Error())
	}

	err = row.Scan(&version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.Wrap(zbxerr.ErrorEmptyResult, err.Error())
		}

		return nil, errs.Wrap(zbxerr.ErrorCannotFetchData, err.Error())
	}

	return version, nil
}
