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

package connectagent

import (
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

func unmarshallYAML(file []byte) (agent rsc.Agent, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &agent); err != nil {
		err = util.NewUnmarshalError(err.Error())
		return
	}

	if agent.Host != "" && agent.SSH.Port == 0 {
		agent.SSH.Port = 22
	}
	// Format file paths
	if agent.SSH.KeyFile, err = util.FormatPath(agent.SSH.KeyFile); err != nil {
		return
	}

	return
}
