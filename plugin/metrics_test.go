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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"golang.zabbix.com/sdk/metric"
)

func Test_getParameters(t *testing.T) {
	type args struct {
		additional *additionalParam
	}

	tests := []struct {
		name string
		args args
		want []*metric.Param
	}{
		{
			"common parameters",
			args{nil},
			[]*metric.Param{
				paramURI,
				paramUsername,
				paramPassword,
				paramDatabase,
				paramTLSConnect,
				paramTLSCaFile,
				paramTLSCertFile,
				paramTLSKeyFile,
				paramCacheMode,
			},
		},
		{
			"empty additions map",
			args{&additionalParam{}},
			[]*metric.Param{
				paramURI,
				paramUsername,
				paramPassword,
				paramDatabase,
				paramTLSConnect,
				paramTLSCaFile,
				paramTLSCertFile,
				paramTLSKeyFile,
				paramCacheMode,
			},
		},
		{
			"with additional parameter",
			args{
				&additionalParam{
					param:    metric.NewParam("test", "Foo bar."),
					position: 4,
				},
			},
			[]*metric.Param{
				paramURI,
				paramUsername,
				paramPassword,
				paramDatabase,
				metric.NewParam("test", "Foo bar."),
				paramTLSConnect,
				paramTLSCaFile,
				paramTLSCertFile,
				paramTLSKeyFile,
				paramCacheMode,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getParameters(tt.args.additional); !reflect.DeepEqual(got, tt.want) {
				var gotString, wantString string

				for _, v := range got {
					gotString = fmt.Sprintf("%s %+v", gotString, v)
				}

				for _, v := range tt.want {
					wantString = fmt.Sprintf("%s %+v", wantString, v)
				}

				t.Errorf("getParameters() = %v,\n want %v", strings.TrimSpace(gotString), strings.TrimSpace(wantString))
			}
		})
	}
}
