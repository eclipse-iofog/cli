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
	"fmt"
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type controllerExecutor struct {
	namespace string
	name      string
}

func newControllerExecutor(namespace, name string) *controllerExecutor {
	exe := &controllerExecutor{}
	exe.namespace = namespace
	exe.name = name
	return exe
}

func (ctrl *controllerExecutor) GetName() string {
	return ctrl.name
}

func (exe *controllerExecutor) Execute() error {
	// Get controller config
	ctrl, err := config.GetController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	// Local
	if ctrl.Host == "localhost" {
		return util.NewInternalError("Not Implemented")
	}

	// K8s
	if ctrl.Kube.Config != "" {
		out, err := util.Exec("KUBECONFIG="+ctrl.Kube.Config, "kubectl", "logs", "-l", "name=controller", "-n", "iofog")
		if err != nil {
			return err
		}
		fmt.Print(out.String())
		return nil
	}

	// Remote
	if ctrl.Host == "" || ctrl.SSH.User == "" || ctrl.SSH.KeyFile == "" || ctrl.SSH.Port == 0 {
		util.Check(util.NewNoConfigError("Controller"))
	}
	ssh := util.NewSecureShellClient(ctrl.SSH.User, ctrl.Host, ctrl.SSH.KeyFile)
	ssh.SetPort(ctrl.SSH.Port)

	if err = ssh.Connect(); err != nil {
		return err
	}

	// Get logs
	out, err := ssh.Run("sudo cat /var/log/iofog-controller/*")
	if err != nil {
		return err
	}
	fmt.Print(out.String())

	return nil
}
