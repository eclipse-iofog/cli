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

package connectcontrolplane

import (
	client "github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

func connect(ctrlPlane config.ControlPlane, endpoint, namespace string) error {
	// Connect to Controller
	ctrl := client.New(endpoint)

	// Get sanitized endpoint
	endpoint = ctrl.GetEndpoint()

	// Login user
	loginRequest := client.LoginRequest{
		Email:    ctrlPlane.IofogUser.Email,
		Password: ctrlPlane.IofogUser.Password,
	}
	if err := ctrl.Login(loginRequest); err != nil {
		return err
	}

	// Get Agents
	listAgentsResponse, err := ctrl.ListAgents()
	if err != nil {
		return err
	}

	// Update Agents config
	for _, agent := range listAgentsResponse.Agents {
		agentConfig := config.Agent{
			Name: agent.Name,
			UUID: agent.UUID,
			Host: agent.IPAddressExternal,
		}
		if err = config.AddAgent(namespace, agentConfig); err != nil {
			return err
		}
	}

	return nil
}
