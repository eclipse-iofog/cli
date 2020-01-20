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

package deployconnector

import (
	"github.com/eclipse-iofog/iofogctl/internal/config"
)

type kubernetesExecutor struct {
	namespace    string
	cnct         *config.Connector
	controlPlane config.ControlPlane
}

func newKubernetesExecutor(namespace string, cnct *config.Connector) *kubernetesExecutor {
	return &kubernetesExecutor{
		namespace: namespace,
		cnct:      cnct,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return exe.cnct.Name
}

func (exe *kubernetesExecutor) Execute() (err error) {
	return nil
	//// Get Control Plane
	//controlPlane, err := config.GetControlPlane(exe.namespace)
	//if err != nil || len(controlPlane.Controllers) == 0 {
	//util.PrintError("You must deploy a Controller to a namespace before deploying any Connector")
	//return err
	//}
	//exe.controlPlane = controlPlane

	//// Get Kubernetes installer
	//installer, err := install.NewKubernetes(exe.cnct.Kube.Config, exe.namespace)
	//if err != nil {
	//return
	//}

	//// Configure deploy
	//installer.SetConnectorImage(exe.cnct.Container.Image)
	//installer.SetConnectorServiceType(exe.cnct.Kube.ServiceType)

	//// Create connector on cluster
	//if err = installer.CreateConnector(exe.cnct.Name, install.IofogUser(exe.controlPlane.IofogUser)); err != nil {
	//return
	//}

	//// Update connector (its a pointer, this is returned to caller)
	//endpoint, err := installer.GetConnectorEndpoint(exe.cnct.Name)
	//if err != nil {
	//return
	//}
	//exe.cnct.Endpoint = endpoint
	//exe.cnct.Created = util.NowUTC()

	//return
}
