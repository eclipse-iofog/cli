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

package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func newViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Open ECN Viewer",
		Run: func(cmd *cobra.Command, args []string) {
			// Get Control Plane
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			ns, err := config.GetNamespace(namespace)
			util.Check(err)
			if len(ns.GetControllers()) == 0 {
				util.PrintError("You must deploy a Control Plane to a namespace to see an ECN Viewer")
				os.Exit(1)
			}
			ctrl := ns.GetControllers()[0]
			URL, err := url.Parse(ctrl.GetEndpoint())
			if err != nil || URL.Host == "" {
				URL, err = url.Parse("//" + ctrl.GetEndpoint()) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
			}
			util.Check(err)
			if URL.Scheme == "" {
				URL.Scheme = "http"
			}
			host, _, err := net.SplitHostPort(URL.Host)
			util.Check(err)
			if util.IsLocalHost(host) {
				host += ":" + iofog.ControllerHostECNViewerPortString
			}
			URL.Host = host
			ecnViewer := URL.String()
			if err := browser.OpenURL(ecnViewer); err != nil {
				util.PrintInfo("To see the ECN Viewer, open your browser and go to:\n")
				util.PrintInfo(fmt.Sprintf("%s\n", URL))
			}
		},
	}
	return cmd
}
