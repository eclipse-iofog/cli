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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
)

type localExecutor struct {
	namespace             string
	name                  string
	client                *iofog.LocalContainer
	localControllerConfig *iofog.LocalControllerConfig
}

func newLocalExecutor(namespace, name string, client *iofog.LocalContainer) *localExecutor {
	exe := &localExecutor{
		namespace:             namespace,
		name:                  name,
		client:                client,
		localControllerConfig: iofog.NewLocalControllerConfig(name),
	}
	return exe
}

func (exe *localExecutor) Execute() error {
	// Clean controller and connector containers
	for _, name := range exe.localControllerConfig.ContainerNames {
		if errClean := exe.client.CleanContainer(name); errClean != nil {
			fmt.Printf("Could not clean Controller container: %v", errClean)
		}
	}

	// Update configuration
	err := config.DeleteController(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	fmt.Printf("\nController %s/%s successfully deleted.\n", exe.namespace, exe.name)

	return config.Flush()
}
