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

package install

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Agent interface {
	Bootstrap() error
	getProvisionKey(string, IofogUser) (string, string, error)
}

// defaultAgent implements commong behavior
type defaultAgent struct {
	name string
	uuid string
}

func (agent *defaultAgent) getProvisionKey(controllerEndpoint string, user IofogUser) (key string, err error) {
	// Connect to controller
	ctrl, err := client.NewAndLogin(client.Options{Endpoint: controllerEndpoint}, user.Email, user.Password)

	// Log in
	Verbose("Accessing Controller to generate Provisioning Key")
	loginRequest := client.LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}
	if err = ctrl.Login(loginRequest); err != nil {
		return
	}

	if agent.uuid != "" {
		err = util.NewInternalError("Agent does not have UUID")
	}

	// Get provisioning key
	provisionResponse, err := ctrl.GetAgentProvisionKey(agent.uuid)
	if err != nil {
		return
	}
	key = provisionResponse.Key
	return
}
