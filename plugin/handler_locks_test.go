//go:build postgresql_tests
// +build postgresql_tests

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
	"fmt"
	"testing"
)

func TestPlugin_locksHandler(t *testing.T) {

	// create pool or acquire conn from old pool for test
	sharedPool, err := getConnPool()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		ctx         context.Context
		conn        *PGConn
		key         string
		params      map[string]string
		extraParams []string
	}
	tests := []struct {
		name    string
		p       *Plugin
		args    args
		wantErr bool
	}{
		{
			fmt.Sprintf("Plugin.locksHandler() should return ptr to Pool for Plugin.locksHandler()"),
			&Impl,
			args{context.Background(), sharedPool, keyLocks, nil, []string{}},

			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := locksHandler(tt.args.ctx, tt.args.conn, tt.args.key, tt.args.params, tt.args.extraParams...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Plugin.locksHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got.(string)) == 0 && err != errors.New("cannot parse data") {
				t.Errorf("Plugin.locksHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
