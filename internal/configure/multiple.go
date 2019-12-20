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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
)

type multipleExecutor struct {
	opt Options
}

func newMultipleExecutor(opt Options) *multipleExecutor {
	return &multipleExecutor{
		opt: opt,
	}
}

func (exe *multipleExecutor) Execute() (err error) {
	// Instantiate executor list
	var executors []execute.Executor

	// Populate list
	if exe.opt.ResourceType == "all" || exe.opt.ResourceType == "agents" {
		executors, err = exe.AddAgentExecutors(executors)
		if err != nil {
			return err
		}
	}
	if exe.opt.ResourceType == "all" || exe.opt.ResourceType == "controllers" {
		executors, err = exe.AddControllerExecutors(executors)
		if err != nil {
			return err
		}
	}
	if exe.opt.ResourceType == "all" || exe.opt.ResourceType == "connectors" {
		executors, err = exe.AddConnectorExecutors(executors)
		if err != nil {
			return err
		}
	}

	// Execute
	for _, executor := range executors {
		if err := executor.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func (exe *multipleExecutor) AddAgentExecutors(executors []execute.Executor) ([]execute.Executor, error) {
	var agents []config.Agent
	var err error
	if exe.opt.UseDetached {
		agentsMap := config.GetDetachedResources().Agents
		for _, agent := range agentsMap {
			agents = append(agents, agent)
		}
	} else {
		agents, err = config.GetAgents(exe.opt.Namespace)
	}
	if err != nil {
		return nil, err
	}
	for _, agent := range agents {
		opt := exe.opt
		opt.Name = agent.Name
		executors = append(executors, newAgentExecutor(opt))
	}

	return executors, nil
}

func (exe *multipleExecutor) AddControllerExecutors(executors []execute.Executor) ([]execute.Executor, error) {
	controllers, err := config.GetControllers(exe.opt.Namespace)
	if err != nil {
		return nil, err
	}
	for _, controller := range controllers {
		opt := exe.opt
		opt.Name = controller.Name
		executors = append(executors, newControllerExecutor(opt))
	}

	return executors, nil
}

func (exe *multipleExecutor) AddConnectorExecutors(executors []execute.Executor) ([]execute.Executor, error) {
	var connectors []config.Agent
	var err error
	if exe.opt.UseDetached {
		connectorsMap := config.GetDetachedResources().Agents
		for _, connector := range connectorsMap {
			connectors = append(connectors, connector)
		}
	} else {
		connectors, err = config.GetAgents(exe.opt.Namespace)
	}
	if err != nil {
		return nil, err
	}
	for _, connector := range connectors {
		opt := exe.opt
		opt.Name = connector.Name
		executors = append(executors, newConnectorExecutor(opt))
	}

	return executors, nil
}

func (exe *multipleExecutor) GetName() string {
	return exe.opt.Name
}
