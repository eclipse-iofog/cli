/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package apps

import (
	"fmt"
	"net/url"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

type applicationExecutor struct {
	controller         IofogController
	app                *Application
	microserviceByName map[string]*client.MicroserviceInfo
	client             *client.Client
	flowInfo           *client.FlowInfo
	applicationInfo    *client.ApplicationInfo
	agentsByName       map[string]*client.AgentInfo
	catalogByID        map[int]*client.CatalogItemInfo
	catalogByName      map[string]*client.CatalogItemInfo
	registryByID       map[int]*client.RegistryInfo
}

func microserviceArrayToClientMap(a []Microservice) (result map[string]*client.MicroserviceInfo) {
	result = make(map[string]*client.MicroserviceInfo)
	for i := 0; i < len(a); i++ {
		// No need to fill information, we only need to know if the name exists
		result[a[i].Name] = &client.MicroserviceInfo{}
	}
	return
}

func newApplicationExecutor(controller IofogController, app *Application) *applicationExecutor {
	exe := &applicationExecutor{
		controller:         controller,
		app:                app,
		microserviceByName: microserviceArrayToClientMap(app.Microservices),
	}

	return exe
}

func (exe *applicationExecutor) execute() (err error) {
	// Init remote resources
	if err = exe.init(); err != nil {
		return
	}

	// Try application API
	// Look for exisiting application
	exe.applicationInfo, err = exe.client.GetApplicationByName(exe.app.Name)

	// If not notfound error, return error
	if _, ok := err.(*client.NotFoundError); err != nil && !ok {
		return err
	}

	// Deploy application
	if err := exe.deploy(); err != nil {
		if _, ok := err.(*client.NotFoundError); ok {
			// If notfound error, try legacy
			return exe.deployLegacy()
		}
		return err
	}
	return nil
}

func (exe *applicationExecutor) init() (err error) {
	baseURL, err := url.Parse(exe.controller.Endpoint)
	if err != nil {
		return fmt.Errorf(errParseControllerURL, err.Error())
	}
	if exe.controller.Token != "" {
		exe.client, err = client.NewWithToken(client.Options{BaseURL: *baseURL}, exe.controller.Token)
	} else {
		exe.client, err = client.NewAndLogin(client.Options{BaseURL: *baseURL}, exe.controller.Email, exe.controller.Password)
	}
	if err != nil {
		return
	}
	listAgents, err := exe.client.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return
	}

	exe.agentsByName = make(map[string]*client.AgentInfo)
	for i := 0; i < len(listAgents.Agents); i++ {
		exe.agentsByName[listAgents.Agents[i].Name] = &listAgents.Agents[i]
	}

	return
}

func (exe *applicationExecutor) create() (err error) {
	microservices, err := mapMicroservicesToClientMicroserviceRequests(exe.app.Microservices)
	if err != nil {
		return err
	}
	if microservices == nil {
		microservices = []client.MicroserviceCreateRequest{}
	}
	routes := mapRoutesToClientRouteRequests(exe.app.Routes)
	if routes == nil {
		routes = []client.ApplicationRouteCreateRequest{}
	}
	template := mapTemplateToClientTemplate(exe.app.Template)
	request := &client.ApplicationCreateRequest{
		Name:          exe.app.Name,
		Microservices: microservices,
		Routes:        &routes,
		Template:      template,
	}

	if _, err = exe.client.CreateApplication(request); err != nil {
		return err
	}
	return nil
}

func (exe *applicationExecutor) update() (err error) {
	// Convert Microservices and Routes
	microservices, err := mapMicroservicesToClientMicroserviceRequests(exe.app.Microservices)
	if err != nil {
		return err
	}
	routes := mapRoutesToClientRouteRequests(exe.app.Routes)
	// Convert Template
	template := mapTemplateToClientTemplate(exe.app.Template)

	request := &client.ApplicationUpdateRequest{
		Name:          &exe.app.Name,
		Routes:        &routes,
		Microservices: &microservices,
		Template:      template,
	}

	if _, err = exe.client.UpdateApplication(exe.app.Name, request); err != nil {
		return err
	}
	return nil
}

func (exe *applicationExecutor) deploy() (err error) {
	// Existing app info retrieved in init
	if exe.applicationInfo == nil {
		if err := exe.create(); err != nil {
			return err
		}
	} else {
		if err := exe.update(); err != nil {
			return err
		}
	}

	// Start application
	if _, err = exe.client.StartApplication(exe.app.Name); err != nil {
		return err
	}
	return nil
}
