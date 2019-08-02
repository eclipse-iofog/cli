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
	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

const TitleHeader = "     _       ____                 __  __    \n" +
	"    (_)___  / __/___  ____  _____/ /_/ / 	 \n" +
	"   / / __ \\/ /_/ __ \\/ __ `/ ___/ __/ /   \n" +
	"  / / /_/ / __/ /_/ / /_/ / /__/ /_/ /   	 \n" +
	" /_/\\____/_/  \\____/\\__, /\\___/\\__/_/  \n" +
	"                   /____/                   \n"

const TitleMessage = "Welcome to the cool new iofogctl Cli!\n" +
	"\n" +
	"Use `iofogctl version` to display the current version.\n\n"

func printHeader() {
	util.PrintInfo(TitleHeader)
	util.PrintInfo("\n")
	util.PrintInfo(TitleMessage)
}

func NewRootCommand() *cobra.Command {

	var cmd = &cobra.Command{
		Use: "iofogctl",
		//Short: "ioFog Unified Command Line Interface",
		PreRun: func(cmd *cobra.Command, args []string) {
			printHeader()
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.SetArgs([]string{"-h"})
			err := cmd.Execute()
			util.Check(err)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Initialize config filename
	cobra.OnInitialize(initConfig)

	// Global flags
	cmd.PersistentFlags().StringVar(&configFilename, "config", "", "CLI configuration file (default is "+config.DefaultConfigPath+")")
	cmd.PersistentFlags().StringP("namespace", "n", "default", "Namespace to execute respective command within")
	cmd.PersistentFlags().BoolVarP(&util.Quiet, "quiet", "q", false, "Toggle for displaying verbose output")
	cmd.PersistentFlags().BoolVarP(&client.Verbose, "verbose", "v", false, "Toggle for displaying verbose output of API client")

	// Register all commands
	cmd.AddCommand(
		newConnectCommand(),
		newDisconnectCommand(),
		newDeployCommand(),
		newDeleteCommand(),
		newCreateCommand(),
		newGetCommand(),
		newDescribeCommand(),
		newLogsCommand(),
		newLegacyCommand(),
		newVersionCommand(),
		newBashCompleteCommand(cmd),
	)

	return cmd
}

// Config file set by --config persistent flag
var configFilename string

// Callback for cobra on initialization
func initConfig() {
	config.Init(configFilename)
}
