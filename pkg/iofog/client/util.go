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

package client

import (
	"bytes"
	"fmt"
	"io"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func getString(in io.Reader) (out string, err error) {
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(in); err != nil {
		return
	}

	out = buf.String()
	return
}

func checkStatusCode(code int, method, url string, body io.Reader) error {
	if code < 200 || code >= 300 {
		bodyString, err := getString(body)
		if err != nil {
			return err
		}
		return util.NewHTTPError(fmt.Sprintf("Received %d from %s %s\n%s", code, method, url, bodyString), code)
	}
	return nil
}
