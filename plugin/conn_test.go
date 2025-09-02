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
	"strings"
	"testing"

	"golang.zabbix.com/sdk/tlsconfig"
)

func Test_createDNS(t *testing.T) {
	type args struct {
		host     string
		port     string
		dbname   string
		user     string
		password string
		mode     string
		details  tlsconfig.Details
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"default",
			args{host: "127.0.0.1", port: "123", dbname: "postgres", user: "foo"},
			[]string{"host=127.0.0.1", "port=123", "dbname=postgres", "user=foo"},
		},
		{
			"with_password",
			args{host: "127.0.0.1", port: "123", dbname: "postgres", user: "foo", password: "bar"},
			[]string{"host=127.0.0.1", "port=123", "dbname=postgres", "user=foo", "password=bar"},
		},
		{
			"tls_connect_require",
			args{
				host:    "127.0.0.1",
				port:    "123",
				dbname:  "postgres",
				user:    "foo",
				details: tlsconfig.Details{TlsConnect: "require"}},
			[]string{"host=127.0.0.1", "port=123", "dbname=postgres", "user=foo", "sslmode=require"},
		},
		{
			"tls_connect_verify_ca",
			args{
				host:    "127.0.0.1",
				port:    "123",
				dbname:  "postgres",
				user:    "foo",
				details: tlsconfig.Details{TlsConnect: "verify-ca", TlsCaFile: "path/to/ca"}},
			[]string{
				"host=127.0.0.1",
				"port=123",
				"dbname=postgres",
				"user=foo",
				"sslmode=verify-ca",
				"sslrootcert=path/to/ca",
			},
		},
		{
			"tls_full",
			args{
				host:   "127.0.0.1",
				port:   "123",
				dbname: "postgres",
				user:   "foo",
				details: tlsconfig.Details{
					TlsConnect:  "verify-full",
					TlsCaFile:   "path/to/ca",
					TlsCertFile: "path/to/cert",
					TlsKeyFile:  "path/to/key",
				}},
			[]string{
				"host=127.0.0.1", "port=123",
				"dbname=postgres",
				"user=foo",
				"sslmode=verify-full",
				"sslrootcert=path/to/ca",
				"sslcert=path/to/cert",
				"sslkey=path/to/key",
			},
		},
		{
			"mode describe",
			args{
				host:   "127.0.0.1",
				port:   "123",
				dbname: "postgres",
				user:   "foo",
				mode:   "describe",
			},
			[]string{
				"host=127.0.0.1", "port=123",
				"dbname=postgres",
				"user=foo",
				"statement_cache_mode=describe",
			},
		}, {
			"mode prepare",
			args{
				host:   "127.0.0.1",
				port:   "123",
				dbname: "postgres",
				user:   "foo",
				mode:   "prepare",
			},
			[]string{
				"host=127.0.0.1", "port=123",
				"dbname=postgres",
				"user=foo",
				"statement_cache_mode=prepare",
			},
		},
		{
			"full",
			args{
				host:   "127.0.0.1",
				port:   "123",
				dbname: "postgres",
				user:   "foo",
				mode:   "prepare",
				details: tlsconfig.Details{
					TlsConnect:  "verify-full",
					TlsCaFile:   "path/to/ca",
					TlsCertFile: "path/to/cert",
					TlsKeyFile:  "path/to/key",
				}},
			[]string{
				"host=127.0.0.1", "port=123",
				"dbname=postgres",
				"user=foo",
				"statement_cache_mode=prepare",
				"sslmode=verify-full",
				"sslrootcert=path/to/ca",
				"sslcert=path/to/cert",
				"sslkey=path/to/key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := createDNS(
				tt.args.host,
				tt.args.port,
				tt.args.dbname,
				tt.args.user,
				tt.args.password,
				tt.args.mode,
				tt.args.details,
			)

			if !sameValues(strings.Split(tmp, " "), tt.want) {
				t.Errorf(
					"createDNS() = %v, want %v, test checks for values and not value order",
					tmp,
					strings.Join(tt.want, " "),
				)
			}
		})
	}
}

func Test_renameTLS(t *testing.T) {
	type args struct {
		in string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{"rqeuired", args{"required"}, "require"},
		{"verify_ca", args{"verify_ca"}, "verify-ca"},
		{"verify_full", args{"verify_full"}, "verify-full"},
		{"any_other_string", args{"foobar"}, "foobar"},
		{"empty", args{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := renameTLS(tt.args.in); got != tt.want {
				t.Errorf("renameTLS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func sameValues(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}

	dif := make(map[string]int)
	for _, v := range x {
		dif[v]++
	}

	for _, v := range y {
		if _, ok := dif[v]; !ok {
			return false
		}

		dif[v]--

		if dif[v] == 0 {
			delete(dif, v)
		}
	}

	return len(dif) == 0
}
