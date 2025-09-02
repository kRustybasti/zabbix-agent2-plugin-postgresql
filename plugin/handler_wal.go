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

// walHandler executes select from directory which contains wal files and returns JSON if all is OK or nil otherwise.
func walHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (any, error) {
	var walJSON string

	query := `SELECT row_to_json(T)
			    FROM (
					SELECT
						CASE
							WHEN pg_is_in_recovery() THEN 0
							ELSE pg_wal_lsn_diff(pg_current_wal_lsn(),'0/00000000')
						END AS WRITE,
						CASE 
							WHEN NOT pg_is_in_recovery() THEN 0
							ELSE pg_wal_lsn_diff(pg_last_wal_receive_lsn(),'0/00000000')
						END AS RECEIVE,
						count(*)
						FROM pg_ls_waldir() AS COUNT
					) T;`

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, errs.Wrap(zbxerr.ErrorCannotFetchData, err.Error())
	}

	err = row.Scan(&walJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.Wrap(zbxerr.ErrorEmptyResult, err.Error())
		}

		return nil, errs.Wrap(zbxerr.ErrorCannotFetchData, err.Error())
	}

	return walJSON, nil
}
