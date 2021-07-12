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

package install

import (
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

// LocalAgent uses Container exec commands
type LocalAgent struct {
	defaultAgent
	client           *LocalContainer
	localAgentConfig *LocalAgentConfig
}

func NewLocalAgent(localAgentConfig *LocalAgentConfig, client *LocalContainer) *LocalAgent {
	return &LocalAgent{
		defaultAgent:     defaultAgent{name: localAgentConfig.Name},
		localAgentConfig: localAgentConfig,
		client:           client,
	}
}

func (agent *LocalAgent) Bootstrap() error {
	// Nothing to do for local agent, bootstraping is done inside the image.
	return nil
}

func (agent *LocalAgent) Configure(controllerEndpoint string, user IofogUser) (string, error) {
	provisioningEndpoint := controllerEndpoint
	if controllerEndpoint == "" || util.IsLocalHost(controllerEndpoint) {
		localControllerEndpoint, err := agent.client.GetLocalControllerEndpoint()
		if err != nil {
			return "", err
		}
		provisioningEndpoint = "localhost:51121"
		controllerEndpoint = localControllerEndpoint
	}

	key, err := agent.getProvisionKey(provisioningEndpoint, user)
	if err != nil {
		return "", err
	}

	// Instantiate provisioning commands
	controllerBaseURL, err := util.GetBaseURL(controllerEndpoint)
	if err != nil {
		return "", err
	}
	cmds := [][]string{
		{"iofog-agent", "config", "-idc", "off"},
		{"iofog-agent", "config", "-a", controllerBaseURL.String()},
		{"iofog-agent", "provision", key},
		{"iofog-agent", "config", "-sf", "10", "-cf", "10"},
	}

	// Execute commands
	for _, cmd := range cmds {
		result, err := agent.client.ExecuteCmd(agent.localAgentConfig.ContainerName, cmd)
		if result.ExitCode != 0 {
			return "", util.NewError(fmt.Sprintf("Command: %v failed with exit code %d\nStdout: %s\n Stderr: %s\n", cmd, result.ExitCode, result.StdOut, result.StdErr))
		}
		if err != nil {
			return "", err
		}
	}

	return agent.uuid, nil
}
