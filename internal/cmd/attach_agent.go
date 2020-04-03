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
	attach "github.com/eclipse-iofog/iofogctl/v2/internal/attach/agent"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"github.com/spf13/cobra"
)

func newAttachAgentCommand() *cobra.Command {
	opt := attach.Options{}
	cmd := &cobra.Command{
		Use:   "agent NAME",
		Short: "Attach an Agent to an existing Namespace",
		Long: `Attach an Agent to an existing Namespace.

The Agent will be provisioned with the Controller within the Namespace.

Can be used after detach command to re-provision the Agent. Can also be used with Agents that have not been detached.
`,
		Example: `iofogctl attach agent NAME --detached
		iofogctl attach agent NAME --host AGENT_HOST --user SSH_USER --port SSH_PORT --key SSH_PRIVATE_KEY_PATH`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name and namespace of agent
			opt.Name = args[0]
			var err error
			opt.Namespace, err = cmd.Flags().GetString("namespace")
			util.Check(err)
			opt.UseDetached, err = cmd.Flags().GetBool("detached")
			util.Check(err)

			// Look inside detached resources if no connection info provided
			if opt.Host == "" && opt.User == "" && opt.KeyFile == "" {
				opt.UseDetached = true
			}
			if opt.UseDetached && opt.Host != "" {
				util.PrintNotify("Using detached resource list, host will be ignored")
			}

			// Run the command
			exe := attach.NewExecutor(opt)
			err = exe.Execute()
			util.Check(err)

			util.PrintSuccess("Successfully attached " + opt.Name + " to namespace " + opt.Namespace)
		},
	}
	cmd.Flags().StringVar(&opt.Host, "host", "", "Hostname of remote host")
	cmd.Flags().StringVar(&opt.User, "user", "", "Username of remote host")
	cmd.Flags().StringVar(&opt.KeyFile, "key", "", "Path to private SSH key")
	cmd.Flags().IntVar(&opt.Port, "port", 22, "Port number that iofogctl uses to SSH into remote hosts")

	return cmd
}
