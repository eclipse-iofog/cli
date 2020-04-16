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

package movemicroservice

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func Execute(namespace, name, agent string) error {
	util.SpinStart(fmt.Sprintf("Moving microservice %s", name))

	// Update local cache based on Controller
	if err := iutil.UpdateAgentCache(namespace); err != nil {
		return err
	}

	// Init remote resources
	clt, err := iutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	msvc, err := clt.GetMicroserviceByName(name)
	if err != nil {
		return err
	}

	destAgent, err := clt.GetAgentByName(agent, false)
	if err != nil {
		return err
	}

	_, err = clt.UpdateMicroservice(client.MicroserviceUpdateRequest{
		UUID:      msvc.UUID,
		AgentUUID: &destAgent.UUID,
		// Ports and Routes get automatically updated by the SDK, to avoid deletion of port mapping or route, those fields are mandatory
		Ports:  msvc.Ports,
		Routes: msvc.Routes,
	})

	return err
}
