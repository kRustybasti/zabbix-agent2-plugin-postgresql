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
	"net/http"
	"time"

	"github.com/omeid/go-yarn"
	"golang.zabbix.com/sdk/errs"
	"golang.zabbix.com/sdk/metric"
	"golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/zbxerr"
)

const (
	Name       = "PostgreSQL"
	sqlExt     = ".sql"
	hkInterval = 10
)

var (
	_ plugin.Runner       = (*Plugin)(nil)
	_ plugin.Configurator = (*Plugin)(nil)
	_ plugin.Exporter     = (*Plugin)(nil)
	_ plugin.Accessor     = (*Plugin)(nil)
)

// Plugin inherits plugin.Base and store plugin-specific data.
type Plugin struct {
	plugin.Base
	connMgr *ConnManager
	options PluginOptions
}

// Impl is the pointer to the plugin implementation.
var Impl Plugin

// Export implements the Exporter interface.
//
//nolint:gocyclo,cyclop
func (p *Plugin) Export(key string, rawParams []string, pluginCtx plugin.ContextProvider) (any, error) {
	if key == keyCustomQuery && !p.options.CustomQueriesEnabled {
		return nil, errs.Errorf("key %q is disabled", keyCustomQuery)
	}

	m, ok := metrics[key]
	if !ok {
		return nil, errs.Wrapf(zbxerr.ErrorUnsupportedMetric, "unknown metric %q", key)
	}

	params, extraParams, hc, err := m.EvalParams(rawParams, p.options.Sessions)
	if err != nil {
		return nil, err
	}

	err = metric.SetDefaults(params, hc, p.options.Default)
	if err != nil {
		return nil, err
	}

	connID, err := createConnID(params)
	if err != nil {
		return nil, err
	}

	handleMetric := getHandlerFunc(key)
	if handleMetric == nil {
		return nil, zbxerr.ErrorUnsupportedMetric
	}

	conn, err := p.connMgr.GetConnection(connID, params)
	if err != nil {
		// Special logic of processing connection errors should be used if pgsql.ping is requested
		// because it must return pingFailed if any error occurred.
		if key == keyPing {
			return pingFailed, nil
		}

		p.Errf(err.Error())

		return nil, err
	}

	timeout := conn.callTimeout

	if pluginCtx != nil && timeout < time.Second*time.Duration(pluginCtx.Timeout()) {
		timeout = time.Second * time.Duration(pluginCtx.Timeout())
	}

	handlerCtx, cancel := context.WithTimeout(conn.ctx, timeout)
	defer cancel()

	result, err := handleMetric(handlerCtx, conn, key, params, extraParams...)
	if err != nil {
		ctxErr := handlerCtx.Err()
		if ctxErr != nil && errors.Is(ctxErr, context.DeadlineExceeded) {
			p.Errf(
				"failed to handle metric: query execution timeout %s exceeded: %s",
				timeout.String(),
				err.Error(),
			)

			return nil, errs.New("query execution timeout exceeded")
		}

		p.Errf("failed to handle metric %q: %s", key, err.Error())

		return nil, err
	}

	return result, err
}

// Start implements the Runner interface and performs initialization when plugin is activated.
func (p *Plugin) Start() {
	p.connMgr = NewConnManager(
		time.Duration(p.options.KeepAlive)*time.Second,
		time.Duration(p.options.Timeout)*time.Second,
		time.Duration(p.options.CallTimeout)*time.Second,
		hkInterval*time.Second,
		p.setCustomQuery(),
	)
}

func (p *Plugin) setCustomQuery() yarn.Yarn {
	if p.options.CustomQueriesPath == "" {
		return yarn.NewFromMap(map[string]string{})
	}

	queryStorage, err := yarn.New(http.Dir(p.options.CustomQueriesPath), "*"+sqlExt)
	if err != nil {
		p.Errf(err.Error())
		// create empty storage if error occurred
		return yarn.NewFromMap(map[string]string{})
	}

	return queryStorage
}

// Stop implements the Runner interface and frees resources when plugin is deactivated.
func (p *Plugin) Stop() {
	p.connMgr.Destroy()
	p.connMgr = nil
}
