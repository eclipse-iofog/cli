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

package install

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

// Remote agent uses SSH
type RemoteAgent struct {
	defaultAgent
	ssh           *util.SecureShellClient
	version       string
	repo          string
	token         string
	dir           string
	procs         AgentProcedures
	customInstall bool // Flag set when custom install scripts are provided
}

type AgentProcedures struct {
	check          Entrypoint `yaml:"-"` // Check prereqs script (runs for default and custom procedures)
	Deps           Entrypoint `yaml:"deps,omitempty"`
	Install        Entrypoint `yaml:"install,omitempty"`
	Uninstall      Entrypoint `yaml:"uninstall,omitempty"`
	scriptNames    []string   `yaml:"-"` // List of all script names to be pushed to Agent
	scriptContents []string   `yaml:"-"` // List of contents of scripts to be pushed to Agent
}

type Entrypoint struct {
	Name     string   `yaml:"entrypoint"`
	Args     []string `yaml:"args"`
	destPath string   `yaml:"-"` // Dir + filename on Agent
}

func (script *Entrypoint) getCommand() string {
	args := strings.Join(script.Args, " ")
	return fmt.Sprintf("%s %s", script.destPath, args)
}

func NewRemoteAgent(user, host string, port int, privKeyFilename, agentName, agentUUID string) (*RemoteAgent, error) {
	ssh, err := util.NewSecureShellClient(user, host, privKeyFilename)
	if err != nil {
		return nil, err
	}
	ssh.SetPort(port)
	agent := &RemoteAgent{
		defaultAgent: defaultAgent{name: agentName, uuid: agentUUID},
		ssh:          ssh,
		version:      util.GetAgentVersion(),
		dir:          pkg.agentDir,
		procs: AgentProcedures{
			check: Entrypoint{
				Name:     pkg.scriptPrereq,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptPrereq),
			},
			Deps: Entrypoint{
				Name:     pkg.scriptInstallDeps,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptInstallDeps),
			},
			Install: Entrypoint{
				Name:     pkg.scriptInstallIofog,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptInstallIofog),
				Args: []string{
					util.GetAgentVersion(),
					"",
					"",
				},
			},
			Uninstall: Entrypoint{
				Name:     pkg.scriptUninstallIofog,
				destPath: util.JoinAgentPath(pkg.agentDir, pkg.scriptUninstallIofog),
			},
			scriptNames: []string{
				pkg.scriptPrereq,
				pkg.scriptInit,
				pkg.scriptInstallDeps,
				pkg.scriptInstallJava,
				pkg.scriptInstallDocker,
				pkg.scriptInstallIofog,
				pkg.scriptUninstallIofog,
			},
		},
	}
	// Get script contents from embedded files
	for _, scriptName := range agent.procs.scriptNames {
		scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(scriptName))
		if err != nil {
			return nil, err
		}
		agent.procs.scriptContents = append(agent.procs.scriptContents, scriptContent)
	}
	return agent, nil
}

func (agent *RemoteAgent) CustomizeProcedures(dir string, procs *AgentProcedures) error {
	// Format source directory of script files
	dir, err := util.FormatPath(dir)
	if err != nil {
		return err
	}

	// Load script files into memory
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !file.IsDir() {
			procs.scriptNames = append(procs.scriptNames, file.Name())
			content, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, string(content))
		}
	}

	// Add prereq script and entrypoint
	procs.scriptNames = append(procs.scriptNames, pkg.scriptPrereq)
	prereqContent, err := util.GetStaticFile(addAgentAssetPrefix(pkg.scriptPrereq))
	if err != nil {
		return err
	}
	procs.scriptContents = append(procs.scriptContents, prereqContent)
	procs.check.destPath = util.JoinAgentPath(agent.dir, pkg.scriptPrereq)

	// Add default entrypoints and scripts if necessary (user not provided)
	if procs.Deps.Name == "" {
		procs.Deps = agent.procs.Deps
		for _, script := range []string{pkg.scriptInstallDeps, pkg.scriptInstallDocker, pkg.scriptInstallJava} {
			procs.scriptNames = append(procs.scriptNames, script)
			scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(script))
			if err != nil {
				return err
			}
			procs.scriptContents = append(procs.scriptContents, scriptContent)
		}
	}
	if procs.Install.Name == "" {
		procs.Install = agent.procs.Install
		procs.scriptNames = append(procs.scriptNames, pkg.scriptInstallIofog)
		scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(pkg.scriptInstallIofog))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	} else {
		agent.customInstall = true
	}
	if procs.Uninstall.Name == "" {
		procs.Uninstall = agent.procs.Uninstall
		procs.scriptNames = append(procs.scriptNames, pkg.scriptUninstallIofog)
		scriptContent, err := util.GetStaticFile(addAgentAssetPrefix(pkg.scriptUninstallIofog))
		if err != nil {
			return err
		}
		procs.scriptContents = append(procs.scriptContents, scriptContent)
	}

	// Set destination paths where scripts appear on Agent
	procs.Deps.destPath = util.JoinAgentPath(agent.dir, procs.Deps.Name)
	procs.Install.destPath = util.JoinAgentPath(agent.dir, procs.Install.Name)
	procs.Uninstall.destPath = util.JoinAgentPath(agent.dir, procs.Uninstall.Name)

	agent.procs = *procs
	return nil
}

func (agent *RemoteAgent) SetVersion(version string) {
	if version == "" || agent.customInstall {
		return
	}
	agent.version = version
	agent.procs.Install.Args[0] = version
}

func (agent *RemoteAgent) SetRepository(repo, token string) {
	if repo == "" || agent.customInstall {
		return
	}
	agent.repo = repo
	agent.procs.Install.Args[1] = repo
	agent.token = token
	agent.procs.Install.Args[2] = token
}

func (agent *RemoteAgent) Bootstrap() error {
	// Prepare Agent for bootstrap
	if err := agent.copyInstallScriptsToAgent(); err != nil {
		return err
	}

	// Define bootstrap commands
	cmds := []command{
		{
			cmd: agent.procs.check.getCommand(),
			msg: "Checking prerequisites on Agent " + agent.name,
		},
		{
			cmd: agent.procs.Deps.getCommand(),
			msg: "Installing dependancies on Agent " + agent.name,
		},
		{
			cmd: fmt.Sprintf("sudo %s", agent.procs.Install.getCommand()),
			msg: "Installing ioFog daemon on Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err := agent.run(cmds); err != nil {
		return err
	}

	return nil
}

func (agent *RemoteAgent) Configure(controllerEndpoint string, user IofogUser) (string, error) {
	key, err := agent.getProvisionKey(controllerEndpoint, user)
	if err != nil {
		return "", err
	}

	controllerBaseURL, err := util.GetBaseURL(controllerEndpoint)
	if err != nil {
		return "", err
	}
	// Instantiate commands
	cmds := []command{
		{
			cmd: "sudo iofog-agent config -a " + controllerBaseURL.String(),
			msg: "Configuring Agent " + agent.name + " with Controller URL " + controllerBaseURL.String(),
		},
		{
			cmd: "sudo iofog-agent provision " + key,
			msg: "Provisioning Agent " + agent.name + " with Controller",
		},
	}

	// Execute commands on remote server
	if err := agent.run(cmds); err != nil {
		return "", err
	}

	return agent.uuid, nil
}

func (agent *RemoteAgent) Deprovision() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo iofog-agent deprovision",
			msg: "Deprovisioning Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil && !isNotProvisionedError(err) {
		return
	}

	return
}

func (agent *RemoteAgent) Stop() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo iofog-agent deprovision",
			msg: "Deprovisioning Agent " + agent.name,
		},
	}
	if err = agent.run(cmds); err != nil && !isNotProvisionedError(err) {
		return err
	}

	cmds = []command{
		{
			cmd: "sudo -S service iofog-agent stop",
			msg: "Stopping Agent " + agent.name,
		},
	}
	if err := agent.run(cmds); err != nil {
		return err
	}

	return
}

func isNotProvisionedError(err error) bool {
	return strings.Contains(err.Error(), "not provisioned")
}

func (agent *RemoteAgent) Prune() (err error) {
	// Prepare commands
	cmds := []command{
		{
			cmd: "sudo -S service iofog-agent prune",
			msg: "Pruning Agent " + agent.name,
		},
	}

	// Execute commands on remote server
	if err := agent.run(cmds); err != nil {
		return err
	}

	return
}

func (agent *RemoteAgent) Uninstall() (err error) {
	// Stop iofog-agent properly
	if err = agent.Stop(); err != nil {
		return
	}

	// Prepare commands
	cmds := []command{
		// TODO: Implement purge on agent
		// {
		// 	cmd: "sudo iofog-agent purge",
		// 	msg: "Deprovisioning Agent " + agent.name,
		// },
		{
			cmd: agent.procs.Uninstall.getCommand(),
			msg: "Removing iofog-agent software " + agent.name,
		},
	}

	// Execute commands on remote server
	if err = agent.run(cmds); err != nil {
		return
	}

	return
}

func (agent *RemoteAgent) run(cmds []command) (err error) {
	// Establish SSH to agent
	if err = agent.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(agent.ssh.Disconnect)

	// Execute commands
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		if _, err = agent.ssh.Run(cmd.cmd); err != nil {
			return err
		}
	}

	return
}

func (agent *RemoteAgent) copyInstallScriptsToAgent() error {
	Verbose("Copying install scripts to Agent " + agent.name)
	cmds := []command{
		{
			cmd: fmt.Sprintf("sudo mkdir -p %s && sudo chmod -R 0777 %s", agent.dir, agent.dir),
			msg: "Creating Agent etc directory",
		},
	}
	if err := agent.run(cmds); err != nil {
		return err
	}
	return agent.copyScriptsToAgent()
}

func (agent *RemoteAgent) copyScriptsToAgent() error {
	// Establish SSH to agent
	if err := agent.ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(agent.ssh.Disconnect)

	// Copy scripts to remote host
	for idx, script := range agent.procs.scriptNames {
		content := agent.procs.scriptContents[idx]
		reader := strings.NewReader(content)
		if err := agent.ssh.CopyTo(reader, agent.dir, script, "0775", int64(len(content))); err != nil {
			return err
		}
	}
	return nil
}

func addAgentAssetPrefix(file string) string {
	return fmt.Sprintf("agent/%s", file)
}

type command struct {
	cmd string
	msg string
}
