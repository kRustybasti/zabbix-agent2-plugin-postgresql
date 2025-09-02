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

// bgwriterHandler executes select  with statistics from pg_stat_bgwriter
// and returns JSON if all is OK or nil otherwise.
func bgwriterHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (any, error) {
	var bgwriterJSON string

	const queryV1 = `
		SELECT row_to_json(T)
		FROM (
			SELECT
				checkpoints_timed,
				checkpoints_req,
				checkpoint_write_time,
				checkpoint_sync_time,
				buffers_checkpoint,
				buffers_clean,
				maxwritten_clean,
				buffers_backend,
				buffers_backend_fsync,
				buffers_alloc
			FROM pg_catalog.pg_stat_bgwriter
		) T;
	`

	const queryV2 = `
		SELECT row_to_json(T)
		FROM (
			SELECT  
				psc.num_timed AS checkpoints_timed,
				psc.num_requested AS checkpoints_req,
				psc.write_time AS checkpoint_write_time,
				psc.sync_time AS checkpoint_sync_time,
				psc.buffers_written AS buffers_checkpoint,
				psb.buffers_clean AS buffers_clean,
				psb.maxwritten_clean AS maxwritten_clean,
				psb.buffers_alloc AS buffers_alloc
			FROM 
				pg_catalog.pg_stat_checkpointer AS psc, 
				pg_catalog.pg_stat_bgwriter AS psb
		) T;
	  `

	var query string

	version := conn.PostgresVersion()

	switch {
	// Postgres V17 and higher.
	case version >= 170000:
		query = queryV2
	default:
		query = queryV1
	}

	row, err := conn.QueryRow(ctx, query)
	if err != nil {
		return nil, errs.WrapConst(err, zbxerr.ErrorCannotFetchData) //nolint:wrapcheck
	}

	err = row.Scan(&bgwriterJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.WrapConst(err, zbxerr.ErrorEmptyResult) //nolint:wrapcheck
		}

		return nil, errs.WrapConst(err, zbxerr.ErrorCannotFetchData) //nolint:wrapcheck
	}

	return bgwriterJSON, nil
}
