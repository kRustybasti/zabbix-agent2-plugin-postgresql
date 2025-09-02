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

// cacheHandler finds cache hit percent and returns int64 if all is OK or nil otherwise.
func cacheHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (any, error) {
	var cache float64

	query := `SELECT round(sum(blks_hit)*100/sum(blks_hit+blks_read), 2) FROM pg_catalog.pg_stat_database;`

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&cache)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	return cache, nil
}
