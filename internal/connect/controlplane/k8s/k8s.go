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

package connectk8scontrolplane

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/internal/execute"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog/install"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type kubernetesExecutor struct {
	controlPlane *rsc.KubernetesControlPlane
	namespace    string
}

func newKubernetesExecutor(controlPlane *rsc.KubernetesControlPlane, namespace string) *kubernetesExecutor {
	return &kubernetesExecutor{
		controlPlane: controlPlane,
		namespace:    namespace,
	}
}

func (exe *kubernetesExecutor) GetName() string {
	return "Kubernetes Control Plane"
}

func NewManualExecutor(namespace, name, endpoint, kubeConfig, email, password string) (execute.Executor, error) {
	fmtKubeConfig, err := util.FormatPath(kubeConfig)
	if err != nil {
		return nil, err
	}
	controlPlane := &rsc.KubernetesControlPlane{
		IofogUser: rsc.IofogUser{
			Email:    email,
			Password: password,
		},
		KubeConfig: fmtKubeConfig,
		ControllerPods: []*rsc.KubernetesController{
			{
				PodName:  name,
				Endpoint: formatEndpoint(endpoint),
			},
		},
	}

	return newKubernetesExecutor(controlPlane, namespace), nil
}

func NewExecutor(namespace, name string, yaml []byte, kind config.Kind) (execute.Executor, error) {
	// Read the input file
	controlPlane, err := unmarshallYAML(yaml)
	if err != nil {
		return nil, err
	}

	return newKubernetesExecutor(controlPlane, namespace), nil
}

func (exe *kubernetesExecutor) Execute() (err error) {
	// Instantiate Kubernetes cluster object
	k8s, err := install.NewKubernetes(exe.controlPlane.KubeConfig, exe.namespace)
	if err != nil {
		return err
	}

	// Check the resources exist in K8s namespace
	if err = k8s.ExistsInNamespace(exe.namespace); err != nil {
		return err
	}

	// Get Controller endpoint
	endpoint, err := k8s.GetControllerEndpoint()
	if err != nil {
		return err
	}

	// Establish connection
	err = connect(exe.controlPlane, endpoint, exe.namespace)
	if err != nil {
		return err
	}

	// TODO: Get Kubernetes pods
	for idx := int32(0); idx < exe.controlPlane.Replicas.Controller; idx++ {
		k8sPod := rsc.KubernetesController{
			Endpoint: endpoint,
			PodName:  fmt.Sprintf("Kubernetes-%d", idx),
			Created:  util.NowUTC(),
		}
		if err := exe.controlPlane.AddController(&k8sPod); err != nil {
			return err
		}
	}

	ns, err := config.GetNamespace(exe.namespace)
	if err != nil {
		return err
	}
	ns.SetControlPlane(exe.controlPlane)

	return config.Flush()
}

// TODO: remove duplication
func connect(ctrlPlane rsc.ControlPlane, endpoint, namespace string) error {
	// Connect to Controller
	ctrl, err := client.NewAndLogin(client.Options{Endpoint: endpoint}, ctrlPlane.GetUser().Email, ctrlPlane.GetUser().Password)
	if err != nil {
		return err
	}

	// Get Agents
	listAgentsResponse, err := ctrl.ListAgents()
	if err != nil {
		return err
	}

	// Update Agents config
	for _, agent := range listAgentsResponse.Agents {
		agentConfig := rsc.Agent{
			Name: agent.Name,
			UUID: agent.UUID,
			Host: agent.IPAddressExternal,
		}
		if err = config.AddAgent(namespace, agentConfig); err != nil {
			return err
		}
	}

	return nil
}

func formatEndpoint(endpoint string) string {
	before := util.Before(endpoint, ":")
	after := util.After(endpoint, ":")
	if after == "" {
		after = iofog.ControllerPortString
	}
	return before + ":" + after
}
