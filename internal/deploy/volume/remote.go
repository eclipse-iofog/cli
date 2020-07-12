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

package deployvolume

import (
	"errors"
	"fmt"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type remoteExecutor struct {
	volume rsc.Volume
	ns     *rsc.Namespace
	agents []*rsc.RemoteAgent
}

func (exe remoteExecutor) GetName() string {
	return "deploying Volume " + exe.volume.Name
}

func (exe remoteExecutor) Execute() error {
	util.SpinStart("Pushing volumes to Agents")
	// Transfer files
	nbAgents := len(exe.agents)
	ch := make(chan error, nbAgents)
	for idx := range exe.agents {
		go exe.execute(idx, ch)
	}
	for idx := 0; idx < nbAgents; idx++ {
		if err := <-ch; err != nil {
			return err
		}
	}
	// Update config
	exe.ns.UpdateVolume(exe.volume)
	return config.Flush()
}

func (exe remoteExecutor) execute(agentIdx int, ch chan error) {
	agent := exe.agents[agentIdx]

	// Connect
	ssh := util.NewSecureShellClient(agent.SSH.User, agent.Host, agent.SSH.KeyFile)
	if err := ssh.Connect(); err != nil {
		msg := `Failed to Connect to Agent %s.
%s`
		ch <- errors.New(fmt.Sprintf(msg, agent.Name, err.Error()))
		return
	}
	defer ssh.Disconnect()

	// Create dest dir
	if err := ssh.CreateFolder(exe.volume.Destination); err != nil {
		msg := `Failed to create base directory %s on Agent %s.
%s`
		ch <- errors.New(fmt.Sprintf(msg, exe.volume.Destination, agent.Name, err.Error()))
		return
	}
	// Create tmp dir
	tmp := "/tmp/iofogctlssh" + exe.volume.Destination
	if err := ssh.CreateFolder(tmp); err != nil {
		msg := `Failed to create base directory %s on Agent %s.
		%s`
		ch <- errors.New(fmt.Sprintf(msg, exe.volume.Destination, agent.Name, err.Error()))
		return
	}
	// Copy volume to tmp
	if err := ssh.CopyFolderTo(exe.volume.Source, tmp, exe.volume.Permissions, true); err != nil {
		msg := `Failed to copy volume to Agent %s.
%s`
		ch <- errors.New(fmt.Sprintf(msg, agent.Name, err.Error()))
		return
	}
	// Move volume from tmp to dest
	if stdout, err := ssh.Run(fmt.Sprintf("sudo -S cp -pR %s/* %s", tmp, exe.volume.Destination)); err != nil {
		msg := `Failed to move volume to from %s to %s on Agent %s.
	%s
	%s`
		ch <- errors.New(fmt.Sprintf(msg, tmp, exe.volume.Destination, agent.Name, err.Error(), stdout))
		return
	}
	// Remove tmp
	if stdout, err := ssh.Run(fmt.Sprintf("rm -rf %s", tmp)); err != nil {
		msg := `Failed clearing tmp volume data %s from Agent %s.
	%s
	%s`
		ch <- errors.New(fmt.Sprintf(msg, tmp, agent.Name, err.Error(), stdout))
		return
	}

	ch <- nil
}
