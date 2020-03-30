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

package get

import (
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type agentExecutor struct {
	namespace    string
	showDetached bool
}

func newAgentExecutor(namespace string, showDetached bool) *agentExecutor {
	a := &agentExecutor{}
	a.namespace = namespace
	a.showDetached = showDetached
	return a
}

func (exe *agentExecutor) GetName() string {
	return ""
}

func (exe *agentExecutor) Execute() error {
	if exe.showDetached {
		printDetached()
		if err := generateDetachedAgentOutput(); err != nil {
			return err
		}
		return nil
	}
	if err := generateAgentOutput(exe.namespace, true); err != nil {
		return err
	}
	// Flush occurs in generateAgentOutput
	return nil
}

func generateDetachedAgentOutput() error {
	detachedAgents := config.GetDetachedAgents()
	// Make an index of agents the client knows about and pre-process any info
	agentsToPrint := make(map[string]client.AgentInfo)
	for _, agent := range detachedAgents {
		agentsToPrint[agent.GetName()] = client.AgentInfo{
			Name:              agent.GetName(),
			IPAddressExternal: agent.GetHost(),
		}
	}
	return tabulateAgents(agentsToPrint)
}

func generateAgentOutput(namespace string, printNS bool) error {
	// Get Config
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Make an index of agents the client knows about and pre-process any info
	agentsToPrint := make(map[string]client.AgentInfo)
	for _, agent := range ns.GetAgents() {
		agentsToPrint[agent.GetName()] = client.AgentInfo{
			Name:              agent.GetName(),
			IPAddressExternal: agent.GetHost(),
		}
	}

	// Connect to Controller if it is ready
	// Instantiate client
	// Log into Controller
	ctrl, err := internal.NewControllerClient(namespace)
	if err != nil {
		return tabulateAgents(agentsToPrint)
	}

	// Get Agents from Controller
	listAgentsResponse, err := ctrl.ListAgents(client.ListAgentsRequest{})
	if err != nil {
		return err
	}

	// Process Agents
	newAgents := false
	for _, remoteAgent := range listAgentsResponse.Agents {
		// Server may have agents that the client is not aware of, update config if so
		if _, exists := agentsToPrint[remoteAgent.Name]; !exists {
			// TODO: Check whether its local or remote
			// Overwrite config based on backend
			if err := config.UpdateAgent(
				namespace,
				&rsc.RemoteAgent{
					Name: remoteAgent.Name,
					UUID: remoteAgent.UUID,
					Host: remoteAgent.IPAddressExternal,
				}); err != nil {
				return err
			}
			newAgents = true
		}

		// Use the pre-processed default info if necessary
		if remoteAgent.IPAddressExternal == "0.0.0.0" {
			remoteAgent.IPAddressExternal = agentsToPrint[remoteAgent.Name].IPAddressExternal
		}

		// Add details for output
		agentsToPrint[remoteAgent.Name] = remoteAgent
	}
	if newAgents {
		// Save the new agents
		if err := config.Flush(); err != nil {
			return err
		}
	}

	if printNS {
		printNamespace(namespace)
	}

	return tabulateAgents(agentsToPrint)
}

func tabulateAgents(agentInfos map[string]client.AgentInfo) error {
	// Generate table and headers
	table := make([][]string, len(agentInfos)+1)
	headers := []string{
		"AGENT",
		"STATUS",
		"AGE",
		"UPTIME",
		"ADDR",
		"VERSION",
	}
	table[0] = append(table[0], headers...)
	// Populate rows
	idx := 0
	for _, agent := range agentInfos {
		// if UUID is empty, we assume the agent is not provisioned
		if agent.UUID == "" {
			row := []string{
				agent.Name,
				"not provisioned",
				"-",
				"-",
				agent.IPAddressExternal,
				"-",
			}
			table[idx+1] = append(table[idx+1], row...)
		} else {
			age, _ := util.ElapsedRFC(agent.CreatedTimeRFC3339, util.NowRFC())
			uptime := time.Duration(agent.UptimeMs) * time.Millisecond
			row := []string{
				agent.Name,
				agent.DaemonStatus,
				age,
				util.FormatDuration(uptime),
				agent.IPAddressExternal,
				agent.Version,
			}
			table[idx+1] = append(table[idx+1], row...)
		}
		idx = idx + 1
	}

	// Print table
	return print(table)
}
