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

// autovacuumHandler returns count of autovacuum workers if all is OK or nil otherwise.
func autovacuumHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (any, error) {
	var countAutovacuumWorkers int64

	query := `SELECT count(*)
				FROM pg_catalog.pg_stat_activity
				WHERE backend_type = 'autovacuum worker'
				 AND state <> 'idle'
				 AND pid <> pg_catalog.pg_backend_pid()`

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&countAutovacuumWorkers)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	return countAutovacuumWorkers, nil
}
