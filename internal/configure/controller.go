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

package configure

import (
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type controllerExecutor struct {
	namespace string
	name      string
	host      string
	keyFile   string
	user      string
	port      int
}

func newControllerExecutor(opt Options) *controllerExecutor {
	return &controllerExecutor{
		namespace: opt.Namespace,
		name:      opt.Name,
		keyFile:   opt.KeyFile,
		user:      opt.User,
		port:      opt.Port,
		host:      opt.Host,
	}
}

func (exe *controllerExecutor) GetName() string {
	return exe.name
}

func (exe *controllerExecutor) Execute() error {
	// Get config
	controlPlane, err := config.GetControlPlane(exe.namespace)
	if err != nil {
		return err
	}
	controller, err := controlPlane.GetController(exe.name)
	if err != nil {
		return err
	}

	// Disallow editing host if already exists
	if controller.Host != "" && exe.host != "" {
		util.NewInputError("Cannot edit existing host address of Controller. Can only add host address where it doesn't exist after running connect command")
	}

	// Disallow editing vanilla fields for k8s Controller
	if controlPlane.Kube.Config != "" && (exe.port != 0 || exe.keyFile != "" || exe.user != "" || exe.host != "") {
		return util.NewInputError("Controller " + exe.name + " is deployed on Kubernetes. You cannot add SSH details to this Controller")
	}

	// Only add/overwrite values provided
	if exe.host != "" {
		controller.Host = exe.host
		controller.Endpoint, err = util.GetControllerEndpoint(exe.host)
		if err != nil {
			return err
		}
	}
	if exe.keyFile != "" {
		controller.SSH.KeyFile, err = util.FormatPath(exe.keyFile)
		if err != nil {
			return err
		}
	}
	if exe.user != "" {
		controller.SSH.User = exe.user
	}
	if exe.port != 0 {
		controller.SSH.Port = exe.port
	}

	// Add port if not specified or existing
	if controller.Host != "" && controller.SSH.Port == 0 {
		controller.SSH.Port = 22
	}

	// Save config
	if err = config.UpdateController(exe.namespace, controller); err != nil {
		return err
	}

	return config.Flush()
}
