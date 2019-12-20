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

package delete

import (
	"fmt"

	apps "github.com/eclipse-iofog/iofog-go-sdk/pkg/apps"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	deleteagent "github.com/eclipse-iofog/iofogctl/internal/delete/agent"
	deleteapplication "github.com/eclipse-iofog/iofogctl/internal/delete/application"
	deletecatalogitem "github.com/eclipse-iofog/iofogctl/internal/delete/catalog_item"
	deleteconnector "github.com/eclipse-iofog/iofogctl/internal/delete/connector"
	deletecontroller "github.com/eclipse-iofog/iofogctl/internal/delete/controller"
	deletecontrolplane "github.com/eclipse-iofog/iofogctl/internal/delete/controlplane"
	deletemicroservice "github.com/eclipse-iofog/iofogctl/internal/delete/microservice"
	"github.com/eclipse-iofog/iofogctl/internal/execute"
)

type Options struct {
	Namespace string
	InputFile string
	Soft      bool
}

var kindOrder = []apps.Kind{
	config.CatalogItemKind,
	apps.MicroserviceKind,
	apps.ApplicationKind,
	apps.AgentKind,
	apps.ConnectorKind,
	apps.ControllerKind,
	apps.ControlPlaneKind,
}

var kindHandlers = map[apps.Kind]func(execute.KindHandlerOpt) (execute.Executor, error){
	apps.ApplicationKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteapplication.NewExecutor(opt.Namespace, opt.Name)
	},
	apps.MicroserviceKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletemicroservice.NewExecutor(opt.Namespace, opt.Name)
	},
	apps.ControlPlaneKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontrolplane.NewExecutor(opt.Namespace, opt.Name, false)
	},
	apps.AgentKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteagent.NewExecutor(opt.Namespace, opt.Name, false, false)
	},
	apps.ConnectorKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deleteconnector.NewExecutor(opt.Namespace, opt.Name, false, false)
	},
	apps.ControllerKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecontroller.NewExecutor(opt.Namespace, opt.Name, false)
	},
	config.CatalogItemKind: func(opt execute.KindHandlerOpt) (exe execute.Executor, err error) {
		return deletecatalogitem.NewExecutor(opt.Namespace, opt.Name)
	},
}

func Execute(opt *Options) error {
	executorsMap, err := execute.GetExecutorsFromYAML(opt.InputFile, opt.Namespace, kindHandlers)
	if err != nil {
		return err
	}

	// Microservice, Application, Agent, Connector, Controller, ControlPlane
	for idx := range kindOrder {
		if err = execute.RunExecutors(executorsMap[kindOrder[idx]], fmt.Sprintf("delete %s", kindOrder[idx])); err != nil {
			return err
		}
	}

	return nil
}
