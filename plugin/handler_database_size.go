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
	"golang.zabbix.com/sdk/zbxerr"
)

// databaseSizeHandler gets info about count and size of archive files and returns JSON if all is OK or nil otherwise.
func databaseSizeHandler(ctx context.Context, conn PostgresClient,
	_ string, params map[string]string, _ ...string) (any, error) {
	var countSize int64

	query := `SELECT pg_database_size(datname::text)
		FROM pg_catalog.pg_database
   		WHERE datistemplate = false
			 AND datname = $1;`

	row, err := conn.QueryRow(ctx, query, params["Database"])
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&countSize)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	return countSize, nil
}
