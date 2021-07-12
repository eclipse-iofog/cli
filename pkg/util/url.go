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

package util

import (
	"fmt"
	"net/url"
)

func GetBaseURL(controllerEndpoint string) (*url.URL, error) {
	u, err := url.Parse(controllerEndpoint)
	if err != nil || u.Host == "" {
		// Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
		u, err = url.Parse("//" + controllerEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Controller URL (%s): %s", controllerEndpoint, err.Error())
		}
	}

	// Default protocol
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	// Default path
	if u.Path == "" {
		u.Path = "api/v3"
	}
	u.RawQuery = ""
	u.Fragment = ""

	return u, nil
}
