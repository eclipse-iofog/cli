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

package deletecontroller

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor interface {
	Execute() error
}

func NewExecutor(namespace, name string) (Executor, error) {
	// Get controller from config
	ctrl, err := config.GetController(namespace, name)
	if err != nil {
		return nil, err
	}

	// Local executor
	if util.IsLocalHost(ctrl.Host) {
		cli, err := iofog.NewLocalContainerClient()
		if err != nil {
			return nil, err
		}
		return newLocalExecutor(namespace, name, cli), nil
	}

	// Kubernetes executor
	if ctrl.KubeConfig != "" {
		return newKubernetesExecutor(namespace, name), nil
	}

	// Default executor
	return newRemoteExecutor(namespace, name), nil
}
