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

package pruneagent

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type executor struct {
	name      string
	namespace string
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	return executor{name: name, namespace: namespace}, nil
}

func (exe executor) GetName() string {
	return exe.name
}

func (exe executor) Execute() error {
	util.SpinStart("Pruning Agent")
	agent, err := config.GetAgent(exe.namespace, exe.name)
	if err == nil {
		// Deprovision agent
		if util.IsLocalHost(agent.Host) {
			if err = exe.localAgentPrune(); err != nil {
				return err
			}
		} else {
			if err = exe.remoteAgentPrune(agent); err != nil {
				return err
			}
		}
	}
	return nil
}
