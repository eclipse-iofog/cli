/*
 *  *******************************************************************************
 *  * Copyright (c) 2020 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package route

import (
	"fmt"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

func Execute(namespace, name, newName string) error {

	// Init remote resources
	clt, err := iutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	route, err := clt.GetRoute(name)
	if err != nil {
		return err
	}

	util.SpinStart(fmt.Sprintf("Renaming route %s", name))
	route.Name = newName

	if err := clt.PatchRoute(name, route); err != nil {
		return err
	}

	return err
}
