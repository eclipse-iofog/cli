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
	"fmt"

	"github.com/eclipse-iofog/iofogctl/pkg/iofog/install"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

func k8sExecute(kubeConfig, namespace, podSelector string, cliCmd, cmd []string) {
	kubeConfig, err := util.FormatPath(kubeConfig)
	util.Check(err)
	// Connect to cluster
	//Execute
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	util.Check(err)
	// Instantiate Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	util.Check(err)
	podList, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: podSelector})
	if err != nil {
		return
	}
	podName := podList.Items[0].Name
	kubeArgs := []string{"exec", podName, "-n", namespace, "--"}
	kubeArgs = append(kubeArgs, cliCmd...)
	kubeArgs = append(kubeArgs, cmd...)
	out, err := util.Exec("KUBECONFIG="+kubeConfig, "kubectl", kubeArgs...)
	util.Check(err)
	fmt.Print(out.String())
}

func localExecute(container string, localCLI, localCmd []string) {
	// Execute command
	localContainerClient, err := install.NewLocalContainerClient()
	util.Check(err)
	cmd := append(localCLI, localCmd...)
	result, err := localContainerClient.ExecuteCmd(container, cmd)
	util.Check(err)
	fmt.Print(result.StdOut)
	if len(result.StdErr) > 0 {
		util.PrintError(result.StdErr)
	}
}

func remoteExec(user, host, keyFile string, port int, cliCmd string, cmd []string) {
	ssh := util.NewSecureShellClient(user, host, keyFile)
	ssh.SetPort(port)
	util.Check(ssh.Connect())
	defer ssh.Disconnect()

	sshCmd := cliCmd
	for _, arg := range cmd {
		sshCmd = sshCmd + " " + arg
	}
	logs, err := ssh.Run(sshCmd)
	util.Check(err)
	fmt.Print(logs.String())
}

// NOTE: (Serge) This code will be discarded eventually. Keeping it one file.
func newLegacyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "legacy resource RESOURCE COMMAND ARGS...",
		Short: "Execute commands using legacy CLI",
		Long:  `Execute commands using legacy CLI`,
		Example: `iofogctl get all
iofogctl legacy controller NAME iofog
iofogctl legacy connector NAME status
iofogctl legacy agent NAME status`,
		Args: cobra.MinimumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			// Get resource type arg
			resource := args[0]
			// Get resource name
			name := args[1]

			// Get namespace option
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)

			switch resource {
			case "controller":
				// Get config
				ctrl, err := config.GetController(namespace, name)
				util.Check(err)
				cliCommand := []string{"sudo", "iofog-controller"}
				if ctrl.KubeConfig != "" {
					k8sExecute(ctrl.KubeConfig, namespace, "name=controller", cliCommand, args[2:])
				} else if util.IsLocalHost(ctrl.Host) {
					localExecute(install.GetLocalContainerName("controller"), cliCommand, args[2:])
				} else {
					if ctrl.Host == "" || ctrl.User == "" || ctrl.KeyFile == "" || ctrl.Port == 0 {
						util.Check(util.NewError("Cannot execute legacy command because SSH details for this Controller are not available"))
					}
					remoteExec(ctrl.User, ctrl.Host, ctrl.KeyFile, ctrl.Port, "sudo iofog-controller", args[2:])
				}
			case "agent":
				// Get config
				agent, err := config.GetAgent(namespace, name)
				util.Check(err)
				if util.IsLocalHost(agent.Host) {
					localExecute(install.GetLocalContainerName("agent"), []string{"iofog-agent"}, args[2:])
					return
				} else {
					// SSH connect
					if agent.Host == "" || agent.User == "" || agent.KeyFile == "" || agent.Port == 0 {
						util.Check(util.NewError("Cannot execute legacy command because SSH details for this Agent are not available"))
					}
					remoteExec(agent.User, agent.Host, agent.KeyFile, agent.Port, "sudo iofog-agent", args[2:])
				}
			case "connector":
				// Get config
				connector, err := config.GetConnector(namespace, name)
				util.Check(err)
				cliCommand := []string{"sudo", "iofog-connector"}
				if connector.KubeConfig != "" {
					k8sExecute(connector.KubeConfig, namespace, "name=connector-"+name, cliCommand, args[2:])
				} else if util.IsLocalHost(connector.Host) {
					localExecute(install.GetLocalContainerName("connector"), cliCommand, args[2:])
				} else {
					if connector.Host == "" || connector.User == "" || connector.KeyFile == "" || connector.Port == 0 {
						util.Check(util.NewError("Cannot execute legacy command because SSH details for this Connector are not available"))
					}
					remoteExec(connector.User, connector.Host, connector.KeyFile, connector.Port, "sudo iofog-connector", args[2:])
				}
			default:
				util.Check(util.NewInputError("Unknown legacy CLI " + resource))
			}
		},
	}

	return cmd
}
