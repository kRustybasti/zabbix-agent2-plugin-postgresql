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

package main

import (
	"errors"
	"fmt"
	"os"

	"golang.zabbix.com/plugin/postgresql/plugin"
	"golang.zabbix.com/sdk/errs"
	sdkplugin "golang.zabbix.com/sdk/plugin"
	"golang.zabbix.com/sdk/plugin/container"
	"golang.zabbix.com/sdk/plugin/flag"
)

const COPYRIGHT_MESSAGE = //
`Copyright (C) 2025 Zabbix SIA
License AGPLv3: GNU Affero General Public License version 3 <https://www.gnu.org/licenses/>.
This is free software: you are free to change and redistribute it according to
the license. There is NO WARRANTY, to the extent permitted by law.`

const (
	PLUGIN_VERSION_MAJOR = 8
	PLUGIN_VERSION_MINOR = 0
	PLUGIN_VERSION_PATCH = 0
	PLUGIN_VERSION_RC    = "alpha1"
)

func main() {
	args, err := flag.HandleFlags()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to handle flags: "))
	}

	pluginInfo := &sdkplugin.Info{
		Name:             plugin.Name,
		BinName:          os.Args[0],
		CopyrightMessage: COPYRIGHT_MESSAGE,
		MajorVersion:     PLUGIN_VERSION_MAJOR,
		MinorVersion:     PLUGIN_VERSION_MINOR,
		PatchVersion:     PLUGIN_VERSION_PATCH,
		Alphatag:         PLUGIN_VERSION_RC,
	}

	err = flag.DecideActionFromFlags(args, &plugin.Impl, pluginInfo, nil)
	if err != nil {
		if errors.Is(err, errs.ErrExitGracefully) {
			// exit gracefully if parameter supposed to exit after execution
			exitGracefully()
		}

		exitWithError(errs.Wrap(err, "failed to execute plugin functions: "))
	}

	h, err := container.NewHandler(plugin.Impl.Name())
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to create plugin handler: "))
	}

	plugin.Impl.Logger = h

	err = h.Execute()
	if err != nil {
		exitWithError(errs.Wrap(err, "failed to execute plugin handler: "))
	}
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	os.Exit(1)
}

func exitGracefully() {
	os.Exit(0)
}
