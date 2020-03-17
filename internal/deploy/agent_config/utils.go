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

package deployagentconfig

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type RouterMode string

const (
	EdgeRouter     RouterMode = "edge"
	InteriorRouter RouterMode = "interior"
	NoneRouter     RouterMode = "none"
)

func getRouterMode(config config.AgentConfiguration) RouterMode {
	if config.RouterConfig.RouterMode != nil {
		return RouterMode(*config.RouterConfig.RouterMode)
	} else {
		return EdgeRouter
	}
}

func Validate(config config.AgentConfiguration) error {
	routerMode := getRouterMode(config)

	if routerMode != EdgeRouter && routerMode != InteriorRouter && routerMode != NoneRouter {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. RouterMode has to be one of edge, interior, none. Default is: edge", config.Name))
	}
	if routerMode != NoneRouter && config.NetworkRouter != nil {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. Cannot have a network if routerMode is different from none. Current router mode is: %s", config.Name, routerMode))
	}
	if routerMode == NoneRouter && config.UpstreamRouters != nil && len(*config.UpstreamRouters) > 0 {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. Cannot have a upstreamRouters if routerMode is none", config.Name))
	}
	if routerMode != InteriorRouter && (config.RouterConfig.EdgeRouterPort != nil || config.RouterConfig.InterRouterPort != nil) {
		return util.NewInputError(fmt.Sprintf("Agent config %s validation failed. Cannot have a edgeRouterPort or InterRouterPort if routerMode is different from interior. Current router mode is: %s", config.Name, routerMode))
	}

	return nil
}

func findAgentUuidInList(list []client.AgentInfo, name string) (uuid string, err error) {
	if name == iofog.VanillaRouterAgentName {
		return name, nil
	}
	for _, agent := range list {
		if agent.Name == name {
			return agent.UUID, nil
		}
	}
	return "", util.NewNotFoundError(fmt.Sprintf("Could not find router: %s\n", name))
}

// Process update the config to translate agent names into uuids, and sets the host value if needed
func Process(agentConfig config.AgentConfiguration, name string, clt *client.Client) (config.AgentConfiguration, error) {
	// If local agent, set fixed config
	if agentConfig.Host != nil && util.IsLocalHost(*agentConfig.Host) {
		upstreamRouters := []string{}
		routerMode := "interior"
		edgeRouterPort := 56721
		interRouterPort := 56722
		agentConfig.UpstreamRouters = &upstreamRouters
		agentConfig.RouterConfig = client.RouterConfig{
			RouterMode:      &routerMode,
			EdgeRouterPort:  &edgeRouterPort,
			InterRouterPort: &interRouterPort,
		}
		return agentConfig, nil
	}

	routerMode := getRouterMode(agentConfig)
	agentList, err := clt.ListAgents()
	if err != nil {
		return agentConfig, err
	}

	// Try to find current agent based on name
	var agent *client.AgentInfo
	for idx := range agentList.Agents {
		if name == agentList.Agents[idx].Name {
			agent = &agentList.Agents[idx]
			break
		}
	}

	if agentConfig.UpstreamRouters != nil {
		upstreamRoutersUUID := []string{}
		for _, agentName := range *agentConfig.UpstreamRouters {
			uuid, err := findAgentUuidInList(agentList.Agents, agentName)
			if err != nil {
				return agentConfig, err
			}
			upstreamRoutersUUID = append(upstreamRoutersUUID, uuid)
		}
		agentConfig.UpstreamRouters = &upstreamRoutersUUID
	}

	if agentConfig.NetworkRouter != nil {
		uuid, err := findAgentUuidInList(agentList.Agents, *agentConfig.NetworkRouter)
		if err != nil {
			return agentConfig, err
		}
		agentConfig.NetworkRouter = &uuid
	}

	if routerMode != NoneRouter && agentConfig.Host == nil {
		if agent == nil {
			return agentConfig, util.NewInputError(fmt.Sprintf("Could not infere agent host for agent %s. Host is required because router mode is %s", name, routerMode))
		}
		agentConfig.Host = &agent.IPAddressExternal
	}

	return agentConfig, nil
}

func getAgentUpdateRequestFromAgentConfig(agentConfig config.AgentConfiguration) (request client.AgentUpdateRequest) {
	var fogTypePtr *int64
	if agentConfig.FogType != nil {
		fogType, found := config.FogTypeStringMap[*agentConfig.FogType]
		if !found {
			fogType = 0
		}
		fogTypePtr = &fogType
	}
	request.Location = agentConfig.Location
	request.Latitude = agentConfig.Latitude
	request.Longitude = agentConfig.Longitude
	request.Description = agentConfig.Description
	request.Name = agentConfig.Name
	request.FogType = fogTypePtr
	request.AgentConfiguration = agentConfig.AgentConfiguration
	return
}

func createAgentFromConfiguration(agentConfig config.AgentConfiguration, name string, clt *client.Client) (uuid string, err error) {
	updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(agentConfig)
	createAgentRequest := client.CreateAgentRequest{
		AgentUpdateRequest: updateAgentConfigRequest,
	}
	if createAgentRequest.AgentUpdateRequest.Name == "" {
		createAgentRequest.AgentUpdateRequest.Name = name
	}
	if createAgentRequest.AgentUpdateRequest.FogType == nil {
		fogType := int64(0)
		createAgentRequest.AgentUpdateRequest.FogType = &fogType
	}
	agent, err := clt.CreateAgent(createAgentRequest)
	if err != nil {
		return "", err
	}
	return agent.UUID, nil
}

func updateAgentConfiguration(agentConfig *config.AgentConfiguration, uuid string, clt *client.Client) (err error) {
	if agentConfig != nil {
		updateAgentConfigRequest := getAgentUpdateRequestFromAgentConfig(*agentConfig)
		updateAgentConfigRequest.UUID = uuid

		if _, err = clt.UpdateAgent(&updateAgentConfigRequest); err != nil {
			return
		}
	}
	return nil
}
