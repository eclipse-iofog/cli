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
	"github.com/spf13/cobra"
)

func newRenameCommand() *cobra.Command {
	// Instantiate command
	cmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename the iofog resources that are currently deployed",
		Long:  `Rename the iofog resources that are currently deployed`,
		Example: `iofogctl rename namespace NAME NEW_NAME
				iofogctl rename controlplane NAME NEW_NAME
				iofogctl rename controller NAME NEW_NAME
				iofogctl rename agent NAME NEW_NAME
				iofogctl rename microservice NAME NEW_NAME
				iofogctl rename application NAME NEW_NAME`,
	}

	// Add subcommands
	cmd.AddCommand(
		newRenameNamespaceCommand(),
		newRenameControllerCommand(),
		newRenameConnectorCommand(),
		newRenameAgentCommand(),
		newRenameApplicationCommand(),
		newRenameMicroserviceCommand(),
	)

	return cmd
}
