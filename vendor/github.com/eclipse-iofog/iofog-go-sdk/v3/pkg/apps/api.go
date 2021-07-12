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

package apps

import (
	"net/url"
)

func DeployApplicationTemplate(controller IofogController, controllerBaseURL url.URL, template ApplicationTemplate) error {
	exe := newApplicationTemplateExecutor(controller, controllerBaseURL, template)
	return exe.execute()
}

func DeployApplication(controller IofogController, application *Application) error {
	exe := newApplicationExecutor(controller, application)
	return exe.execute()
}

func DeployMicroservice(controller IofogController, microservice *Microservice) error {
	exe := newMicroserviceExecutor(controller, microservice)
	return exe.execute()
}
