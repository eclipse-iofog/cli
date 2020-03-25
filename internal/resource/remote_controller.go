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

package resource

import (
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type RemoteController struct {
	Name        string  `yaml:"name"`
	Host        string  `yaml:"host"`
	SSH         SSH     `yaml:"ssh,omitempty"`
	Endpoint    string  `yaml:"endpoint,omitempty"`
	Created     string  `yaml:"created,omitempty"`
	Package     Package `yaml:"package,omitempty"`
	SystemAgent Package `yaml:"systemAgent,omitempty"`
}

func (ctrl RemoteController) GetName() string {
	return ctrl.Name
}

func (ctrl RemoteController) GetEndpoint() string {
	return ctrl.Endpoint
}

func (ctrl RemoteController) GetCreatedTime() string {
	return ctrl.Created
}

func (ctrl *RemoteController) SetName(name string) {
	ctrl.Name = name
}

func (ctrl *RemoteController) Sanitize() (err error) {
	// Fix SSH port
	if ctrl.Host != "" && ctrl.SSH.Port == 0 {
		ctrl.SSH.Port = 22
	}
	// Format file paths
	if ctrl.SSH.KeyFile, err = util.FormatPath(ctrl.SSH.KeyFile); err != nil {
		return
	}
	return
}
