/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package get

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
)

type routeExecutor struct {
	namespace string
}

func newRouteExecutor(namespace string) *routeExecutor {
	return &routeExecutor{
		namespace: namespace,
	}
}

func (exe *routeExecutor) GetName() string {
	return ""
}

func (exe *routeExecutor) Execute() error {
	return generateRouteOutput(exe.namespace, true)
}

func generateRouteOutput(namespace string, printNS bool) error {
	_, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Connect to Controller
	clt, err := iutil.NewControllerClient(namespace)
	if err != nil && !rsc.IsNoControlPlaneError(err) {
		return err
	}

	routes := make([]client.Route, 0)
	if err == nil {
		// Populate table
		listResponse, err := clt.ListRoutes()
		if err != nil {
			return err
		}
		routes = listResponse.Routes
	}

	if printNS {
		printNamespace(namespace)
	}

	return tabulateRoutes(namespace, routes)
}

func tabulateRoutes(namespace string, routes []client.Route) error {
	// Generate table and headers
	table := make([][]string, len(routes)+1)
	headers := []string{"ROUTE", "SOURCE MSVC", "DEST MSVC"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, route := range routes {
		// Convert route details
		from, err := iutil.GetMicroserviceName(namespace, route.SourceMicroserviceUUID)
		if err != nil {
			return err
		}
		to, err := iutil.GetMicroserviceName(namespace, route.DestMicroserviceUUID)
		if err != nil {
			return err
		}
		// Store values
		row := []string{
			route.Name,
			from,
			to,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	// Print table
	if err := print(table); err != nil {
		return err
	}
	return nil
}
