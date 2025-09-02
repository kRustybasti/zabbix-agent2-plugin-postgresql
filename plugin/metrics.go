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
	"regexp"
	"strings"

	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/log"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/uri"
)

const (
	keyArchiveSize                     = "pgsql.archive"
	keyAutovacuum                      = "pgsql.autovacuum.count"
	keyBgwriter                        = "pgsql.bgwriter"
	keyCache                           = "pgsql.cache.hit"
	keyConnections                     = "pgsql.connections"
	keyCustomQuery                     = "pgsql.custom.query"
	keyDBStat                          = "pgsql.dbstat"
	keyDBStatSum                       = "pgsql.dbstat.sum"
	keyDatabaseAge                     = "pgsql.db.age"
	keyDatabasesBloating               = "pgsql.db.bloating_tables"
	keyDatabasesDiscovery              = "pgsql.db.discovery"
	keyDatabaseSize                    = "pgsql.db.size"
	keyLocks                           = "pgsql.locks"
	keyOldestXid                       = "pgsql.oldest.xid"
	keyPing                            = "pgsql.ping"
	keyQueries                         = "pgsql.queries"
	keyReplicationCount                = "pgsql.replication.count"
	keyReplicationLagB                 = "pgsql.replication.lag.b"
	keyReplicationLagSec               = "pgsql.replication.lag.sec"
	keyReplicationProcessInfo          = "pgsql.replication.process"
	keyReplicationProcessNameDiscovery = "pgsql.replication.process.discovery"
	keyReplicationRecoveryRole         = "pgsql.replication.recovery_role"
	keyReplicationStatus               = "pgsql.replication.status"
	keyUptime                          = "pgsql.uptime"
	keyVersion                         = "pgsql.version"
	keyWal                             = "pgsql.wal.stat"

	uriParam        = "URI"
	tcpParam        = "tcp"
	userParam       = "User"
	databaseParam   = "Database"
	passwordParam   = "Password"
	tlsConnectParam = "TLSConnect"
	tlsCAParam      = "TLSCAFile"
	tlsCertParam    = "TLSCertFile"
	tlsKeyParam     = "TLSKeyFile"
	cacheModeParam  = "CacheMode"
)

var uriDefaults = &uri.Defaults{Scheme: "tcp", Port: "5432"}

var (
	minDBNameLen = 1
	maxDBNameLen = 63
	maxPassLen   = 512
)

var reSocketPath = regexp.MustCompile(`^.*\.s\.PGSQL\.\d{1,5}$`)

var (
	paramURI = metric.NewConnParam(uriParam, "URI to connect or session name.").
			WithDefault(uriDefaults.Scheme + "://localhost:" + uriDefaults.Port).WithSession().
			WithValidator(PostgresURIValidator{
			Defaults:       uriDefaults,
			AllowedSchemes: []string{tcpParam, "postgresql", "unix"},
		})
	paramUsername = metric.NewConnParam(userParam, "PostgreSQL user.").WithDefault("postgres")
	paramPassword = metric.NewConnParam(passwordParam, "User's password.").
			WithDefault("").
			WithValidator(metric.LenValidator{Max: &maxPassLen})
	paramDatabase = metric.NewConnParam(databaseParam, "Database name to be used for connection.").
			WithDefault("postgres").
			WithValidator(metric.LenValidator{Min: &minDBNameLen, Max: &maxDBNameLen})
	paramTLSConnect  = metric.NewSessionOnlyParam(tlsConnectParam, "DB connection encryption type.").WithDefault("")
	paramTLSCaFile   = metric.NewSessionOnlyParam(tlsCAParam, "TLS ca file path.").WithDefault("")
	paramTLSCertFile = metric.NewSessionOnlyParam(tlsCertParam, "TLS cert file path.").WithDefault("")
	paramTLSKeyFile  = metric.NewSessionOnlyParam(tlsKeyParam, "TLS key file path.").WithDefault("")
	paramCacheMode   = metric.NewSessionOnlyParam(cacheModeParam, "Cache mode for postgresql connections.").
				WithDefault("prepare").
				WithValidator(metric.SetValidator{Set: []string{"prepare", "describe"}, CaseInsensitive: false})
	paramQueryName = metric.NewParam(
		"QueryName", "Name of a custom query (must be equal to a name of an SQL file without an extension).",
	).SetRequired()
	paramTimePeriod = metric.NewParam("TimePeriod", "Execution time limit for count of slow queries.").SetRequired()
)

var metrics = metric.MetricSet{
	keyArchiveSize: metric.New(
		"Returns info about size of archive files.", getParameters(nil), false,
	),
	keyAutovacuum: metric.New(
		"Returns count of autovacuum workers.", getParameters(nil), false,
	),
	keyBgwriter: metric.New(
		"Returns JSON for sum of each type of bgwriter statistic.", getParameters(nil), false,
	),
	keyCache: metric.New(
		"Returns cache hit percent.", getParameters(nil), false,
	),
	keyConnections: metric.New(
		"Returns JSON for sum of each type of connection.", getParameters(nil), false,
	),
	keyCustomQuery: metric.New(
		"Returns result of a custom query.", getParameters(&additionalParam{paramQueryName, 4}), true,
	),
	keyDBStat: metric.New(
		"Returns JSON for sum of each type of statistic.", getParameters(nil), false,
	),
	keyDBStatSum: metric.New(
		"Returns JSON for sum of each type of statistic for all database.", getParameters(nil), false,
	),
	keyDatabaseAge: metric.New(
		"Returns age for specific database.", getParameters(nil), false,
	),
	keyDatabasesBloating: metric.New(
		"Returns percent of bloating tables for each database.", getParameters(nil), false,
	),
	keyDatabasesDiscovery: metric.New(
		"Returns JSON discovery rule with names of databases.", getParameters(nil), false,
	),
	keyDatabaseSize: metric.New(
		"Returns size in bytes for specific database.", getParameters(nil), false,
	),
	keyLocks: metric.New(
		"Returns collect all metrics from pg_locks.", getParameters(nil), false,
	),
	keyOldestXid: metric.New(
		"Returns age of oldest xid.", getParameters(nil), false,
	),
	keyPing: metric.New(
		"Tests if connection is alive or not.", getParameters(nil), false,
	),
	keyQueries: metric.New(
		"Returns queries statistic.", getParameters(&additionalParam{paramTimePeriod, 4}), false,
	),
	keyReplicationCount: metric.New(
		"Returns number of standby servers.", getParameters(nil), false,
	),
	keyReplicationLagB: metric.New(
		"Returns replication lag with Master in byte.", getParameters(nil), false,
	),
	keyReplicationLagSec: metric.New(
		"Returns replication lag with Master in seconds.", getParameters(nil), false,
	),
	keyReplicationProcessNameDiscovery: metric.New(
		"Returns JSON with application name from pg_stat_replication.", getParameters(nil), false,
	),
	keyReplicationProcessInfo: metric.New(
		"Returns flush lag, write lag and replay lag per each sender process.", getParameters(nil), false,
	),
	keyReplicationRecoveryRole: metric.New(
		"Returns postgreSQL recovery role.", getParameters(nil), false,
	),
	keyReplicationStatus: metric.New(
		"Returns postgreSQL replication status.", getParameters(nil), false,
	),
	keyUptime: metric.New(
		"Returns uptime.", getParameters(nil), false,
	),
	keyVersion: metric.New(
		"Returns PostgreSQL version.", getParameters(nil), false,
	),
	keyWal: metric.New(
		"Returns JSON wal by type.", getParameters(nil), false,
	),
}

func init() { //todo remove init and global variable Impl
	err := log.Open(log.Console, log.Info, "", 0)
	if err != nil {
		panic(errs.Wrap(err, "failed to open log"))
	}

	Impl.Logger = log.New(Name)
	err = plugin.RegisterMetrics(&Impl, Name, metrics.List()...)
	if err != nil {
		panic(err)
	}
}

type PostgresURIValidator struct {
	Defaults       *uri.Defaults
	AllowedSchemes []string
}

// handlerFunc defines an interface must be implemented by handlers.
type handlerFunc func(ctx context.Context, conn PostgresClient, key string,
	params map[string]string, extraParams ...string) (res any, err error)

type additionalParam struct {
	param    *metric.Param
	position int
}

// getHandlerFunc returns a handlerFunc related to a given key.
func getHandlerFunc(key string) handlerFunc {
	switch key {
	case keyArchiveSize:
		return archiveHandler
	case keyAutovacuum:
		return autovacuumHandler
	case keyBgwriter:
		return bgwriterHandler
	case keyCache:
		return cacheHandler
	case keyConnections:
		return connectionsHandler
	case keyCustomQuery:
		return customQueryHandler
	case keyDBStat, keyDBStatSum:
		return dbStatHandler
	case keyDatabaseAge:
		return databaseAgeHandler
	case keyDatabasesBloating:
		return databasesBloatingHandler
	case keyDatabasesDiscovery:
		return databasesDiscoveryHandler
	case keyDatabaseSize:
		return databaseSizeHandler
	case keyLocks:
		return locksHandler
	case keyOldestXid:
		return oldestXIDHandler
	case keyPing:
		return pingHandler
	case keyQueries:
		return queriesHandler
	case keyReplicationCount,
		keyReplicationLagB,
		keyReplicationLagSec,
		keyReplicationProcessInfo,
		keyReplicationRecoveryRole,
		keyReplicationStatus:
		return replicationHandler
	case keyReplicationProcessNameDiscovery:
		return processNameDiscoveryHandler
	case keyUptime:
		return uptimeHandler
	case keyVersion:
		return versionHandler
	case keyWal:
		return walHandler
	default:
		return nil
	}
}

func (v PostgresURIValidator) Validate(value *string) error {
	if value == nil {
		return nil
	}

	u, err := uri.New(*value, v.Defaults)
	if err != nil {
		return errs.Wrap(err, "cannot create URI validator")
	}

	isValidScheme := false

	if v.AllowedSchemes != nil {
		for _, s := range v.AllowedSchemes {
			if u.Scheme() == s {
				isValidScheme = true
				break
			}
		}

		if !isValidScheme {
			return fmt.Errorf("allowed schemes: %s", strings.Join(v.AllowedSchemes, ", "))
		}
	}

	if u.Scheme() == "unix" && !reSocketPath.MatchString(*value) {
		return errors.New(
			`socket file must satisfy the format: "/path/.s.PGSQL.nnnn" where nnnn is the server's port number`)
	}

	return nil
}

func getParameters(add *additionalParam) []*metric.Param {
	m := []*metric.Param{
		paramURI,
		paramUsername,
		paramPassword,
		paramDatabase,
		paramTLSConnect,
		paramTLSCaFile,
		paramTLSCertFile,
		paramTLSKeyFile,
		paramCacheMode,
	}

	if add != nil && add.param != nil {
		m = append(m[:add.position+1], m[add.position:]...)
		m[add.position] = add.param
	}

	return m
}
