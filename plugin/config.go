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
	"path/filepath"

	"golang.zabbix.com/sdk/conf"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/plugin"
)

// Session struct holds individual options for PostgreSQL connection for each session.
type Session struct {
	// URI is a connection string consisting of a network scheme, a host address and a port or a path to a Unix-socket.
	URI string `conf:"name=Uri,optional"`

	// User of PostgreSQL server.
	User string `conf:"optional"`

	// Password to send to protected PostgreSQL server.
	Password string `conf:"optional"`

	// Database of PostgreSQL server.
	Database string `conf:"optional"`

	// Connection type of PostgreSQL server.
	TLSConnect string `conf:"name=TLSConnect,optional"`

	// Certificate Authority filepath for PostgreSQL server.
	TLSCAFile string `conf:"name=TLSCAFile,optional"`

	// Certificate filepath for PostgreSQL server.
	TLSCertFile string `conf:"name=TLSCertFile,optional"`

	// Key filepath for PostgreSQL server.
	TLSKeyFile string `conf:"name=TLSKeyFile,optional"`

	// CacheMode for PostgreSQL server.
	CacheMode string `conf:"name=CacheMode,optional"`
}

// PluginOptions are options for PostgreSQL connection.
type PluginOptions struct {
	System plugin.SystemOptions `conf:"optional"` //nolint:staticcheck
	// Timeout is the maximum time in seconds for waiting when a connection has to be established.
	// Default value equals to the global agent timeout.
	Timeout int `conf:"optional,range=1:30"`

	// CallTimeout is the maximum time in seconds for waiting when a request has to be done.
	// Default value equals to the global agent timeout.
	CallTimeout int `conf:"optional,range=1:30"`

	// KeepAlive is a time to wait before unused connections will be closed.
	KeepAlive int `conf:"optional,range=60:900,default=300"`

	// Sessions stores pre-defined named sets of connections settings.
	Sessions map[string]Session `conf:"optional"`

	// CustomQueriesPath is a full pathname of a directory containing *.sql files with custom queries.
	CustomQueriesPath string `conf:"optional"`

	// CustomQueriesEnabled disabled or enabled custom query functionality.
	CustomQueriesEnabled bool `conf:"optional,default=false"`

	// Default stores default connection parameter values from configuration file
	Default Session `conf:"optional"`
}

// Configure implements the Configurator interface.
// Initializes configuration structures.
func (p *Plugin) Configure(global *plugin.GlobalOptions, options any) {
	if err := conf.UnmarshalStrict(options, &p.options); err != nil {
		p.Errf("cannot unmarshal configuration options: %s", err)
	}

	p.options.setCustomQueriesPathDefault()

	if p.options.Timeout == 0 {
		p.options.Timeout = global.Timeout
	}

	if p.options.CallTimeout == 0 {
		p.options.CallTimeout = global.Timeout
	}
}

// Validate implements the Configurator interface.
// Returns an error if validation of a plugin's configuration is failed.
func (*Plugin) Validate(options any) error {
	var opts PluginOptions

	err := conf.UnmarshalStrict(options, &opts)
	if err != nil {
		return errs.Wrap(err, "failed to unmarshal configuration options")
	}

	if opts.CustomQueriesEnabled && opts.CustomQueriesPath != "" && !filepath.IsAbs(opts.CustomQueriesPath) {
		return errs.Errorf("opts.CustomQueriesDir path: '%s' must be absolute", opts.CustomQueriesPath)
	}

	return nil
}
