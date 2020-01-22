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

package deploy

import (
	"fmt"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deployagent "github.com/eclipse-iofog/iofogctl/internal/deploy/agent"
	deployagentconfig "github.com/eclipse-iofog/iofogctl/internal/deploy/agent_config"
	deployapplication "github.com/eclipse-iofog/iofogctl/internal/deploy/application"
	deploycatalogitem "github.com/eclipse-iofog/iofogctl/internal/deploy/catalog_item"
	deployconnector "github.com/eclipse-iofog/iofogctl/internal/deploy/connector"
	deploycontroller "github.com/eclipse-iofog/iofogctl/internal/deploy/controller"
	deploycontrolplane "github.com/eclipse-iofog/iofogctl/internal/deploy/controlplane"
	deploymicroservice "github.com/eclipse-iofog/iofogctl/internal/deploy/microservice"
	deployregistry "github.com/eclipse-iofog/iofogctl/internal/deploy/registry"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

var kindOrder = []apps.Kind{
	// Connector cannot be ran in parallel.
	// apps.ControlPlaneKind,
	// apps.ControllerKind,
	// apps.ConnectorKind,
	// apps.AgentKind,
	config.AgentConfigKind,
	config.RegistryKind,
	config.CatalogItemKind,
	apps.ApplicationKind,
	apps.MicroserviceKind,
}

type Options struct {
	Namespace string
	InputFile string
}

func deployCatalogItem(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycatalogitem.NewExecutor(deploycatalogitem.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployApplication(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployapplication.NewExecutor(deployapplication.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployMicroservice(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploymicroservice.NewExecutor(deploymicroservice.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployControlPlane(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycontrolplane.NewExecutor(deploycontrolplane.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployAgent(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployagent.NewExecutor(deployagent.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployAgentConfig(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployagentconfig.NewExecutor(deployagentconfig.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployConnector(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployconnector.NewExecutor(deployconnector.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployController(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deploycontroller.NewExecutor(deploycontroller.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

func deployRegistry(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
	return deployregistry.NewExecutor(deployregistry.Options{Namespace: opt.Namespace, Yaml: opt.YAML, Name: opt.Name})
}

var kindHandlers = map[apps.Kind]func(execute.KindHandlerOpt) (execute.Executor, error){
	apps.ApplicationKind:   deployApplication,
	config.CatalogItemKind: deployCatalogItem,
	apps.MicroserviceKind:  deployMicroservice,
	apps.ControlPlaneKind:  deployControlPlane,
	apps.AgentKind:         deployAgent,
	config.AgentConfigKind: deployAgentConfig,
	apps.ConnectorKind:     deployConnector,
	apps.ControllerKind:    deployController,
	config.RegistryKind:    deployRegistry,
}

// Execute deploy from yaml file
func Execute(opt *Options) (err error) {
	executorsMap, err := execute.GetExecutorsFromYAML(opt.InputFile, opt.Namespace, kindHandlers)
	if err != nil {
		return err
	}

	// Execute in parallel by priority order
	// Connector cannot be deployed in parallel

	// Controlplane
	if err = execute.RunExecutors(executorsMap[apps.ControlPlaneKind], "deploy control plane"); err != nil {
		return
	}

	// Controller
	if err = execute.RunExecutors(executorsMap[apps.ControllerKind], "deploy controller"); err != nil {
		return
	}

	// Connector
	for idx := range executorsMap[apps.ConnectorKind] {
		if err = executorsMap[apps.ConnectorKind][idx].Execute(); err != nil {
			util.PrintNotify("Error from " + executorsMap[apps.ConnectorKind][idx].GetName() + ": " + err.Error())
			return util.NewError("Failed to deploy")
		}
	}

	// Agents can be deployed with an AgentConfig
	for _, agentGenericExecutor := range executorsMap[apps.AgentKind] {
		agentName := agentGenericExecutor.GetName()
		for i, agentConfigGenericExecutor := range executorsMap[config.AgentConfigKind] {
			// If agent config is provided alonside agent
			if agentName == agentConfigGenericExecutor.GetName() {
				// Get more specialised interfaces
				agentConfigExecutor, configOk := agentConfigGenericExecutor.(deployagentconfig.AgentConfigExecutor)
				agentExecutor, agentOk := agentGenericExecutor.(deployagent.AgentExecutor)
				if !configOk || !agentOk {
					return util.NewInternalError("Agent executor: Could not convert executor")
				}
				// Update agent executor to deploy with config
				agentConfig := agentConfigExecutor.GetConfiguration()
				agentExecutor.SetAgentConfig(&agentConfig)
				// Remove agent config executor from list
				executorsMap[config.AgentConfigKind] = append(executorsMap[config.AgentConfigKind][:i], executorsMap[config.AgentConfigKind][i+1:]...)
				break
			}
		}
		if err = agentGenericExecutor.Execute(); err != nil {
			util.PrintNotify("Error from " + agentName + ": " + err.Error())
			return util.NewError("Failed to deploy")
		}
	}

	// AgentConfig (left overs after agent), CatalogItem, Application, Microservice
	for idx := range kindOrder {
		if err = execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("deploy %s", kindOrder[idx])); err != nil {
			return
		}
	}

	return nil
}
