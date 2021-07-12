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
	"regexp"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/v3/pkg/util"
)

type RemoteSystemImages struct {
	ARM string `yaml:"arm,omitempty"`
	X86 string `yaml:"x86,omitempty"`
}

type RemoteSystemMicroservices struct {
	Router RemoteSystemImages `yaml:"router,omitempty"`
	Proxy  RemoteSystemImages `yaml:"proxy,omitempty"`
}

type ControllerOptions struct {
	User                string
	Host                string
	Port                int
	PrivKeyFilename     string
	Version             string
	Repo                string
	Token               string
	SystemMicroservices RemoteSystemMicroservices
	PidBaseDir          string
	EcnViewerPort       int
}

type database struct {
	databaseName string
	provider     string
	host         string
	user         string
	password     string
	port         int
}

type Controller struct {
	*ControllerOptions
	ssh      *util.SecureShellClient
	db       database
	ctrlDir  string
	iofogDir string
	svcDir   string
}

func NewController(options *ControllerOptions) (*Controller, error) {
	ssh, err := util.NewSecureShellClient(options.User, options.Host, options.PrivKeyFilename)
	if err != nil {
		return nil, err
	}
	ssh.SetPort(options.Port)
	if options.Version == "" || options.Version == "latest" {
		options.Version = util.GetControllerVersion()
	}
	return &Controller{
		ControllerOptions: options,
		ssh:               ssh,
		iofogDir:          "/etc/iofog",
		ctrlDir:           "/etc/iofog/controller",
		svcDir:            "/etc/iofog/controller/service",
	}, nil
}

func (ctrl *Controller) SetControllerExternalDatabase(host, user, password, provider, databaseName string, port int) {
	if provider == "" {
		provider = "postgres"
	}
	if databaseName == "" {
		databaseName = "iofogcontroller"
	}
	ctrl.db = database{
		databaseName: databaseName,
		provider:     provider,
		host:         host,
		user:         user,
		password:     password,
		port:         port,
	}
}

func (ctrl *Controller) CopyScript(srcDir, filename, destDir string) (err error) {
	// Read script from assets
	if srcDir != "" {
		srcDir = util.AddTrailingSlash(srcDir)
	}
	staticFile, err := util.GetStaticFile(srcDir + filename)
	if err != nil {
		return err
	}

	// Copy to /tmp for backwards compatability
	reader := strings.NewReader(staticFile)
	if err := ctrl.ssh.CopyTo(reader, destDir, filename, "0775", int64(len(staticFile))); err != nil {
		return err
	}

	return nil
}

func (ctrl *Controller) Uninstall() (err error) {
	// Stop controller gracefully
	if err := ctrl.Stop(); err != nil {
		return err
	}

	// Connect to server
	Verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(ctrl.ssh.Disconnect)

	// Copy uninstallation scripts to remote host
	Verbose("Copying install files to server")
	scripts := []string{
		"uninstall_iofog.sh",
	}
	for _, script := range scripts {
		if err := ctrl.CopyScript("controller", script, ctrl.ctrlDir); err != nil {
			return err
		}
	}

	cmds := []command{
		{
			cmd: fmt.Sprintf("sudo %s/uninstall_iofog.sh", ctrl.ctrlDir),
			msg: "Uninstalling controller on host " + ctrl.Host,
		},
	}

	// Execute commands
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		_, err = ctrl.ssh.Run(cmd.cmd)
		if err != nil {
			return
		}
	}
	return nil
}

func (ctrl *Controller) Install() (err error) {
	// Connect to server
	Verbose("Connecting to server")
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(ctrl.ssh.Disconnect)

	// Copy installation scripts to remote host
	Verbose("Copying install files to server")
	if _, err = ctrl.ssh.Run(fmt.Sprintf("sudo mkdir -p %s && sudo chmod -R 0777 %s/", ctrl.svcDir, ctrl.iofogDir)); err != nil {
		return err
	}
	scripts := []string{
		"check_prereqs.sh",
		"install_node.sh",
		"install_iofog.sh",
		"set_env.sh",
	}
	for _, script := range scripts {
		if err := ctrl.CopyScript("controller", script, ctrl.ctrlDir); err != nil {
			return err
		}
	}

	// Copy service scripts to remote host
	Verbose("Copying service files to server")
	scripts = []string{
		"iofog-controller.initctl",
		"iofog-controller.systemd",
		"iofog-controller.update-rc",
	}
	for _, script := range scripts {
		if err := ctrl.CopyScript("controller/service", script, ctrl.svcDir); err != nil {
			return err
		}
	}

	// Encode environment variables
	env := []string{}
	if ctrl.db.host != "" {
		env = append(env,
			fmt.Sprintf(`"DB_PROVIDER=%s"`, ctrl.db.provider),
			fmt.Sprintf(`"DB_HOST=%s"`, ctrl.db.host),
			fmt.Sprintf(`"DB_USER=%s"`, ctrl.db.user),
			fmt.Sprintf(`"DB_PASSWORD=%s"`, ctrl.db.password),
			fmt.Sprintf(`"DB_PORT=%d"`, ctrl.db.port),
			fmt.Sprintf(`"DB_NAME=%s"`, ctrl.db.databaseName))
	}
	if ctrl.PidBaseDir != "" {
		env = append(env, fmt.Sprintf("\"PID_BASE=%s\"", ctrl.PidBaseDir))
	}
	if ctrl.EcnViewerPort != 0 {
		env = append(env, fmt.Sprintf("\"VIEWER_PORT=%d\"", ctrl.EcnViewerPort))
	}
	if ctrl.SystemMicroservices.Proxy.X86 != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Proxy_1=%s\"", ctrl.SystemMicroservices.Proxy.X86))
	}
	if ctrl.SystemMicroservices.Proxy.ARM != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Proxy_2=%s\"", ctrl.SystemMicroservices.Proxy.ARM))
	}
	if ctrl.SystemMicroservices.Router.X86 != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Router_1=%s\"", ctrl.SystemMicroservices.Router.X86))
	}
	if ctrl.SystemMicroservices.Router.ARM != "" {
		env = append(env, fmt.Sprintf("\"SystemImages_Router_2=%s\"", ctrl.SystemMicroservices.Router.ARM))
	}

	envString := strings.Join(env, " ")

	// Define commands
	cmds := []command{
		{
			cmd: fmt.Sprintf("%s/check_prereqs.sh", ctrl.ctrlDir),
			msg: "Checking prerequisites on Controller " + ctrl.Host,
		},
		{
			cmd: fmt.Sprintf("sudo %s/install_node.sh", ctrl.ctrlDir),
			msg: "Installing Node.js on Controller " + ctrl.Host,
		},
		{
			cmd: fmt.Sprintf("sudo %s/set_env.sh %s", ctrl.ctrlDir, envString),
			msg: "Setting up environment variables for Controller " + ctrl.Host,
		},
		{
			cmd: fmt.Sprintf("sudo %s/install_iofog.sh %s %s %s", ctrl.ctrlDir, ctrl.Version, ctrl.Repo, ctrl.Token),
			msg: "Installing ioFog on Controller " + ctrl.Host,
		},
	}

	// Execute commands
	for _, cmd := range cmds {
		Verbose(cmd.msg)
		_, err = ctrl.ssh.Run(cmd.cmd)
		if err != nil {
			return
		}
	}

	// Specify errors to ignore while waiting
	ignoredErrors := []string{
		"Process exited with status 7", // curl: (7) Failed to connect to localhost port 8080: Connection refused
	}
	// Wait for Controller
	Verbose("Waiting for Controller " + ctrl.Host)
	if err = ctrl.ssh.RunUntil(
		regexp.MustCompile("\"status\":\"online\""),
		fmt.Sprintf("curl --request GET --url http://localhost:%s/api/v3/status", iofog.ControllerPortString),
		ignoredErrors,
	); err != nil {
		return
	}

	// Wait for API
	endpoint := fmt.Sprintf("%s:%s", ctrl.Host, iofog.ControllerPortString)
	if err = WaitForControllerAPI(endpoint); err != nil {
		return
	}

	return nil
}

func (ctrl *Controller) Stop() (err error) {
	// Connect to server
	if err = ctrl.ssh.Connect(); err != nil {
		return
	}
	defer util.Log(ctrl.ssh.Disconnect)

	// TODO: Clear the database
	// Define commands
	cmds := []string{
		"sudo iofog-controller stop",
	}

	// Execute commands
	for _, cmd := range cmds {
		_, err = ctrl.ssh.Run(cmd)
		if err != nil {
			return
		}
	}

	return
}

func WaitForControllerAPI(endpoint string) (err error) {
	baseURL, err := util.GetBaseURL(endpoint)
	if err != nil {
		return err
	}
	ctrlClient := client.New(client.Options{BaseURL: *baseURL})

	seconds := 0
	for seconds < 60 {
		// Try to create the user, return if success
		if _, err = ctrlClient.GetStatus(); err == nil {
			return
		}
		// Connection failed, wait and retry
		time.Sleep(time.Millisecond * 1000)
		seconds++
	}

	// Return last error
	return
}
