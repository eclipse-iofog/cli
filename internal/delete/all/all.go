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

package deleteall

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	"github.com/eclipse-iofog/iofogctl/internal/delete/controller"
)

func Execute(namespace string) error {
	// Get namespace
	ns, err := config.GetNamespace(namespace)
	if err != nil {
		return err
	}

	// Delete agents
	for _, agent := range ns.Agents {
		exe, err := deleteagent.NewExecutor(namespace, agent.Name)
		if err != nil {
			return err
		}
		err = exe.Execute()
		if err != nil {
			return err
		}
	}

	// Delete controllers
	for _, ctrl := range ns.ControlPlane.Controllers {
		exe, err := deletecontroller.NewExecutor(namespace, ctrl.Name)
		if err != nil {
			return err
		}
		err = exe.Execute()
		if err != nil {
			return err
		}
	}

	// TODO: delete microservices
	return nil
}
