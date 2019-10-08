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
	update "github.com/eclipse-iofog/iofogctl/internal/update/catalog_item"
	deploy "github.com/eclipse-iofog/iofogctl/pkg/iofog/deploy"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
)

func newUpdateCatalogItemCommand() *cobra.Command {
	opt := deploy.CatalogItem{}
	cmd := &cobra.Command{
		Use:     "catalogitem NAME",
		Short:   "Update a catalog item",
		Long:    `Update a catalog item on the ioFog controller`,
		Example: `iofogctl update catalogitem NAME [--x86 x86_IMAGE] [--arm arm_IMAGE] [--registry <remote|local>] [--description DESCRIPTION]`,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Get name
			opt.Name = args[0]

			// Get namespace
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			// Run the command
			err = update.Execute(opt, namespace)
			util.Check(err)

			util.PrintSuccess("Successfully updated catalog item " + opt.Name)
		},
	}

	// Register flags
	cmd.Flags().StringVar(&opt.X86, "x86", "", "Container image to use on x86 agents")
	cmd.Flags().StringVar(&opt.ARM, "arm", "", "Container image to use on arm agents")
	cmd.Flags().StringVarP(&opt.Registry, "registry", "r", "", "Container registry to use. Either 'remote' or 'local'")
	cmd.Flags().StringVarP(&opt.Description, "description", "d", "", "Description of catalog item purpose")

	return cmd
}
