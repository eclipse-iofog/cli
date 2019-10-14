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

package deleteapplication

import (
	"github.com/eclipse-iofog/iofog-go-sdk/pkg/client"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
	name      string
	client    *client.Client
	flow      *client.FlowInfo
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
		name:      name,
	}

	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return exe.name
}

// Execute deletes application by deleting its associated flow
func (exe *Executor) Execute() (err error) {
	util.SpinStart("Deleting Application")
	// Get Control Plane from namespace
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil || len(controlPlane.Controllers) == 0 {
		util.PrintError("You must deploy a Controller to a namespace before deploying any Agents")
		return
	}

	// Init remote resources
	if err = exe.init(controlPlane); err != nil {
		return
	}

	// Delete flow
	if err = exe.client.DeleteFlow(exe.flow.ID); err != nil {
		return
	}

	return nil
}

func (exe *Executor) init(controlPlane config.ControlPlane) (err error) {
	// TODO: replace controllers[0] with controplane variable
	exe.client = client.New(controlPlane.Controllers[0].Endpoint)
	if err = exe.client.Login(client.LoginRequest{Email: controlPlane.IofogUser.Email, Password: controlPlane.IofogUser.Password}); err != nil {
		return
	}
	flow, err := exe.client.GetFlowByName(exe.name)
	if err != nil {
		return
	}
	exe.flow = flow
	return
}
