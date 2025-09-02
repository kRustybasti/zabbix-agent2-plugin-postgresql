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
	"encoding/json"
	"errors"
	"strings"

	"github.com/jackc/pgx/v4"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/zbxerr"
)

// customQueryHandler executes custom user queries from *.sql files.
func customQueryHandler(ctx context.Context, conn PostgresClient,
	_ string, params map[string]string, extraParams ...string) (any, error) {
	queryName := params["QueryName"]

	queryArgs := make([]any, 0, len(extraParams))
	for _, v := range extraParams {
		queryArgs = append(queryArgs, v)
	}

	rows, err := conn.QueryByName(ctx, queryName, queryArgs...)
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}
	defer rows.Close()

	// JSON marshaling
	var data []string

	columns, err := rows.Columns()
	if err != nil {
		return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
	}

	values := make([]any, len(columns))       //nolint:makezero
	valuePointers := make([]any, len(values)) //nolint:makezero

	for i := range values {
		valuePointers[i] = &values[i]
	}

	results := make(map[string]any)

	for rows.Next() {
		err = rows.Scan(valuePointers...)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, zbxerr.ErrorEmptyResult.Wrap(err)
			}

			return nil, zbxerr.ErrorCannotFetchData.Wrap(err)
		}

		setResult(results, values, columns)

		jsonRes, err := json.Marshal(results)
		if err != nil {
			return nil, errs.Wrap(err, "cannot marshal results")
		}

		data = append(data, strings.TrimSpace(string(jsonRes)))
	}

	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		return nil, errs.Wrap(err, "cannot fetch data")
	}

	return "[" + strings.Join(data, ",") + "]", nil
}

func setResult(results map[string]any, values []any, columns []string) {
	for i, value := range values {
		switch v := value.(type) {
		case []uint8:
			results[columns[i]] = string(v)
		default:
			results[columns[i]] = value
		}
	}
}
