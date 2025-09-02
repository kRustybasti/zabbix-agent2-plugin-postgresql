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

// archiveHandler gets info about count and size of archive files and returns JSON if all is OK or nil otherwise.
func archiveHandler(ctx context.Context, conn PostgresClient,
	_ string, _ map[string]string, _ ...string) (any, error) {
	var archiveCountJSON, archiveSizeJSON string

	queryArchiveCount := `SELECT row_to_json(T)
							FROM (
									SELECT archived_count, failed_count
								   	  FROM pg_stat_archiver
								) T;`

	queryArchiveSize := `SELECT row_to_json(T)
							FROM (
								WITH values AS (
									SELECT
										4096/(ceil(pg_settings.setting::numeric/1024/1024))::int AS segment_parts_count,
										setting::bigint AS segment_size,
										('x' || substring(pg_stat_archiver.last_archived_wal from 9 for 8))::bit(32)::int AS last_wal_div,
										('x' || substring(pg_stat_archiver.last_archived_wal from 17 for 8))::bit(32)::int AS last_wal_mod,
										CASE WHEN pg_is_in_recovery() THEN NULL 
											ELSE ('x' || substring(pg_walfile_name(pg_current_wal_lsn()) from 9 for 8))::bit(32)::int END AS current_wal_div,
										CASE WHEN pg_is_in_recovery() THEN NULL 
											ELSE ('x' || substring(pg_walfile_name(pg_current_wal_lsn()) from 17 for 8))::bit(32)::int END AS current_wal_mod
									FROM pg_settings, pg_stat_archiver
									WHERE pg_settings.name = 'wal_segment_size')
								SELECT 
									greatest(coalesce((segment_parts_count - last_wal_mod) + ((current_wal_div - last_wal_div - 1) * segment_parts_count) + current_wal_mod - 1, 0), 0) AS count_files,
									greatest(coalesce(((segment_parts_count - last_wal_mod) + ((current_wal_div - last_wal_div - 1) * segment_parts_count) + current_wal_mod - 1) * segment_size, 0), 0) AS size_files
								FROM values
							) T;`

	row, err := conn.QueryRow(ctx, queryArchiveCount)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&archiveCountJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	row, err = conn.QueryRow(ctx, queryArchiveSize)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	err = row.Scan(&archiveSizeJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, zbxerr.ErrorEmptyResult.Wrap(err)
		}

		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	result := archiveCountJSON[:len(archiveCountJSON)-1] + "," + archiveSizeJSON[1:]

	return result, nil
}
