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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"gopkg.in/yaml.v2"
)

func unmarshallYAML(file []byte) (agent config.Agent, err error) {
	// Unmarshall the input file
	if err = yaml.UnmarshalStrict(file, &agent); err != nil {
		err = util.NewInputError("Could not unmarshall\n" + err.Error())
		return
	}

	if agent.Port == 0 {
		agent.Port = 22
	}
	// Format file paths
	if agent.KeyFile, err = util.FormatPath(agent.KeyFile); err != nil {
		return
	}

	return
}