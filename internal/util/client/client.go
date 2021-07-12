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

package client

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v3/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

// clientCacheRoutine handles concurrent requests for a cached Controller client
func clientCacheRoutine() {
	for {
		request := <-pkg.clientCacheRequestChan
		// Invalidate cache
		if request.namespace == "" {
			pkg.clientCache = make(map[string]*client.Client)
			continue
		}
		result := &clientCacheResult{}
		// From cache
		if cachedClient, exists := pkg.clientCache[request.namespace]; exists {
			result.client = cachedClient
			request.resultChan <- result
			continue
		}
		// Create new client
		ioClient, err := newControllerClient(request.namespace)
		// Failure
		if err != nil {
			result.err = err
			request.resultChan <- result
			continue
		}
		// Save to cache and return new client
		pkg.clientCache[request.namespace] = ioClient
		result.client = ioClient
		request.resultChan <- result
	}
}

// agentCacheRoutine handles concurrent requests for a cached list of Agents
func agentCacheRoutine() {
	for {
		request := <-pkg.agentCacheRequestChan
		if request.namespace == "" {
			// Invalidate cache
			pkg.agentCache = make(map[string][]client.AgentInfo)
			continue
		}
		result := &agentCacheResult{}
		// From cache
		if cachedAgents, exist := pkg.agentCache[request.namespace]; exist {
			result.agents = cachedAgents
			request.resultChan <- result
			continue
		}
		// Client to get agents
		ioClient, err := NewControllerClient(request.namespace)
		if err != nil {
			result.err = err
			request.resultChan <- result
			continue
		}
		// Get agents
		agents, err := getBackendAgents(request.namespace, ioClient)
		if err != nil {
			result.err = err
			request.resultChan <- result
			continue
		}
		// Save to cache and return new agents
		pkg.agentCache[request.namespace] = agents
		result.agents = agents
		request.resultChan <- result
	}
}

func agentSyncRoutine() {
	complete := false
	for {
		request := <-pkg.agentSyncRequestChan
		if complete {
			request.resultChan <- nil
			continue
		}
		if err := syncAgentInfo(request.namespace); err != nil {
			request.resultChan <- err
			continue
		}
		complete = true
		request.resultChan <- nil
	}
}

func syncAgentInfo(namespace string) error {
	// Get local cache Agents
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}
	// Check the Control Plane type
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return err
	}
	if _, ok := controlPlane.(*rsc.LocalControlPlane); ok {
		// Do not update local Agents
		return nil
	}
	// Generate map of config Agents
	agentsMap := make(map[string]*rsc.RemoteAgent)
	var localAgent *rsc.LocalAgent
	for _, baseAgent := range ns.GetAgents() {
		if v, ok := baseAgent.(*rsc.LocalAgent); ok {
			localAgent = v
		} else {
			agentsMap[baseAgent.GetName()] = baseAgent.(*rsc.RemoteAgent)
		}
	}

	// Get backend Agents
	backendAgents, err := GetBackendAgents(namespace)
	if err != nil {
		return err
	}

	// Generate cache types
	agents := make([]rsc.RemoteAgent, len(backendAgents))
	for idx := range backendAgents {
		backendAgent := &backendAgents[idx]
		if localAgent != nil && backendAgent.Name == localAgent.Name {
			localAgent.UUID = backendAgent.UUID
			continue
		}

		agent := rsc.RemoteAgent{
			Name: backendAgent.Name,
			UUID: backendAgent.UUID,
			Host: backendAgent.Host,
		}
		// Update additional info if local cache contains it
		if cachedAgent, exists := agentsMap[backendAgent.Name]; exists {
			agent.Created = cachedAgent.GetCreatedTime()
			agent.SSH = cachedAgent.SSH
		}

		agents[idx] = agent
	}

	// Overwrite the Agents
	ns.DeleteAgents()
	for idx := range agents {
		if err := ns.AddAgent(&agents[idx]); err != nil {
			return err
		}
	}

	if localAgent != nil {
		if err := ns.AddAgent(localAgent); err != nil {
			return err
		}
	}

	return config.Flush()
}

func newControllerClient(namespace string) (*client.Client, error) {
	// Get endpoint
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return nil, err
	}
	controlPlane, err := ns.GetControlPlane()
	if err != nil {
		return nil, err
	}
	endpoint, err := controlPlane.GetEndpoint()
	if err != nil {
		return nil, err
	}

	user := controlPlane.GetUser()
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return nil, err
	}
	cachedClient, err := client.NewAndLogin(client.Options{BaseURL: *baseURL}, user.Email, user.GetRawPassword())
	if err != nil {
		return nil, err
	}
	pkg.clientCache[namespace] = cachedClient

	return cachedClient, nil
}

func getBackendAgents(namespace string, ioClient *client.Client) ([]client.AgentInfo, error) {
	agentList, err := ioClient.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return nil, err
	}
	pkg.agentCache[namespace] = agentList.Agents
	return agentList.Agents, nil
}

func getAgentNameFromUUID(agentMapByUUID map[string]client.AgentInfo, uuid string) (name string) {
	if uuid == iofog.VanillaRouterAgentName {
		return uuid
	}
	agent, found := agentMapByUUID[uuid]
	if !found {
		util.PrintNotify(fmt.Sprintf("Could not find Router: %s\n", uuid))
		name = "UNKNOWN ROUTER: " + uuid
	} else {
		name = agent.Name
	}
	return
}
