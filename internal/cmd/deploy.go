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

package cmd

import (
	"github.com/eclipse-iofog/iofogctl/internal/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newDeployCommand() *cobra.Command {
	// Instantiate options
	opt := &deploy.Options{}

	// Instantiate command
	cmd := &cobra.Command{
		Use:     "deploy",
		Example: `deploy -f platform.yaml`,
		Short:   "Deploy ioFog platform or components on existing infrastructure",
		Long: `Deploy ioFog platform or individual components on existing infrastructure.

The YAML resource specification file should look like this (two Controllers specified for example only):` + "\n```\n" + `kind: ControlPlane
apiVersion: iofog.org/v1
metadata:
  name: alpaca-1 # ControlPlane name
spec:
  controllers:
  - name: k8s # Controller name
    kubeConfig: ~/.kube/conf # Will deploy a controller in a kubernetes cluster
  - name: vanilla
    ssh:
      user: serge # SSH user
      host: 35.239.157.151 # SSH Host - Will deploy a controller as a standalone binary
      keyFile: ~/.ssh/id_rsa # SSH private key
---
apiVersion: iofog.org/v1
kind: Agent
metadata:
  name: agent1 # Agent name
spec:
  ssh:
    user: serge # SSH User
    host: 35.239.157.151 # SSH host
    keyFile: ~/.ssh/id_rsa # SSH private key
---
apiVersion: iofog.org/v1
kind: Agent
metadata:
  name: agent2
spec:
  ssh:
    user: serge
    host: 35.232.114.32
    keyFile: ~/.ssh/id_rsa
` + "\n```\n" + `The complete description of yaml file definition can be found at iofog.org`,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			// Get namespace
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)

			// Execute command
			err = deploy.Execute(opt)
			util.Check(err)

			util.PrintSuccess("Successfully deployed resources to namespace " + opt.Namespace)
		},
	}

	// Register flags
	cmd.Flags().StringVarP(&opt.InputFile, "file", "f", "", "YAML file containing resource definitions for Controllers, Agents, and Microservice to deploy")

	return cmd
}
