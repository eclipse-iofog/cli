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

package deploycontroller

import (
	"fmt"
	"os/user"
	"regexp"

	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
)

type localExecutor struct {
	namespace             string
	ctrl                  *rsc.Controller
	client                *install.LocalContainer
	localControllerConfig *install.LocalContainerConfig
	iofogUser             config.IofogUser
	containersNames       []string
}

func newLocalExecutor(namespace string, ctrl *rsc.Controller, controlPlane rsc.ControlPlane, client *install.LocalContainer) (*localExecutor, error) {
	return &localExecutor{
		namespace: namespace,
		ctrl:      ctrl,
		client:    client,
		localControllerConfig: install.NewLocalControllerConfig(ctrl.Container.Image, install.Credentials{
			User:     ctrl.Container.Credentials.User,
			Password: ctrl.Container.Credentials.Password,
		}),
		iofogUser: controlPlane.IofogUser,
	}, nil
}

func (exe *localExecutor) cleanContainers() {
	for _, name := range exe.containersNames {
		if errClean := exe.client.CleanContainer(name); errClean != nil {
			util.PrintNotify(fmt.Sprintf("Could not clean Controller container: %v", errClean))
		}
	}
}

func (exe *localExecutor) deployContainers() error {
	controllerContainerConfig := exe.localControllerConfig
	controllerContainerName := controllerContainerConfig.ContainerName

	// Deploy controller image
	util.SpinStart("Deploying Controller container")

	// If container already exists, clean it
	if _, err := exe.client.GetContainerByName(controllerContainerName); err == nil {
		if err := exe.client.CleanContainer(controllerContainerName); err != nil {
			return err
		}
	}

	_, err := exe.client.DeployContainer(controllerContainerConfig)
	if err != nil {
		return err
	}

	exe.containersNames = append(exe.containersNames, controllerContainerName)
	// Wait for public API
	util.SpinStart("Waiting for Controller API")
	if err = exe.client.WaitForCommand(
		install.GetLocalContainerName("controller", false),
		regexp.MustCompile("\"status\":[ |\t]*\"online\""),
		"iofog-controller",
		"controller",
		"status",
	); err != nil {
		return err
	}

	return nil
}

func (exe *localExecutor) GetName() string {
	return exe.ctrl.Name
}

func (exe *localExecutor) Execute() error {
	// Deploy Controller images
	if err := exe.deployContainers(); err != nil {
		exe.cleanContainers()
		return err
	}

	// Get current user
	currUser, err := user.Current()
	if err != nil {
		return err
	}

	// Update controller (its a pointer, this is returned to caller)
	controllerContainerConfig := exe.localControllerConfig
	exe.ctrl.Endpoint, err = util.GetControllerEndpoint(fmt.Sprintf("%s:%s", controllerContainerConfig.Host, controllerContainerConfig.Ports[0].Host))
	if err != nil {
		return err
	}
	exe.ctrl.Host = controllerContainerConfig.Host
	exe.ctrl.SSH.User = currUser.Username
	exe.ctrl.Created = util.NowUTC()

	return nil
}
