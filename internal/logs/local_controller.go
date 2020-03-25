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

package logs

import (
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
)

type localControllerExecutor struct {
	controlPlane *rsc.LocalControlPlane
	namespace    string
	name         string
}

func newLocalControllerExecutor(controlPlane *rsc.LocalControlPlane, namespace, name string) *localControllerExecutor {
	return &localControllerExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
		name:         name,
	}
}

func (ctrl *localControllerExecutor) GetName() string {
	return ctrl.name
}

func (exe *localControllerExecutor) Execute() error {
	lc, err := install.NewLocalContainerClient()
	if err != nil {
		return err
	}
	containerName := install.GetLocalContainerName("controller", false)
	stdout, stderr, err := lc.GetLogsByName(containerName)
	if err != nil {
		return err
	}

	printContainerLogs(stdout, stderr)

	return nil
}
