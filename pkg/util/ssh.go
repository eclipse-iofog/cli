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

package util

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type SecureShellClient struct {
	user            string
	host            string
	port            int
	privKeyFilename string
	config          *ssh.ClientConfig
	conn            *ssh.Client
}

func NewSecureShellClient(user, host, privKeyFilename string) *SecureShellClient {
	return &SecureShellClient{
		user:            user,
		host:            host,
		port:            22,
		privKeyFilename: privKeyFilename,
	}
}

func (cl *SecureShellClient) SetPort(port int) {
	SSHVerbose(fmt.Sprintf("Setting port to %v", port))
	cl.port = port
}

func (cl *SecureShellClient) Connect() (err error) {
	// Don't bother connecting twice
	SSHVerbose("Initialiasing connection")
	if cl.conn != nil {
		return nil
	}

	// Parse keys
	SSHVerbose("Parsing keys")
	key, err := cl.getPublicKey()
	if err != nil {
		return err
	}

	// Instantiate config
	SSHVerbose("Configuring SSH client")
	cl.config = &ssh.ClientConfig{
		User: cl.user,
		Auth: []ssh.AuthMethod{
			key,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	SSHVerbose("Config:")
	SSHVerbose(fmt.Sprintf("User: %s", cl.user))
	SSHVerbose(fmt.Sprintf("Auth method: %v", key))

	// Connect
	endpoint := cl.host + ":" + strconv.Itoa(cl.port)
	SSHVerbose(fmt.Sprintf("TCP dialing %s", endpoint))
	cl.conn, err = ssh.Dial("tcp", endpoint, cl.config)
	if err != nil {
		return err
	}

	return nil
}

func (cl *SecureShellClient) Disconnect() error {
	SSHVerbose("Disconnecting...")
	if cl.conn == nil {
		return nil
	}

	err := cl.conn.Close()
	if err != nil {
		return err
	}

	SSHVerbose("Connection closed")
	cl.conn = nil
	return nil
}

func (cl *SecureShellClient) Run(cmd string) (stdout bytes.Buffer, err error) {
	// Establish the session
	SSHVerbose("Creating session...")
	session, err := cl.conn.NewSession()
	if err != nil {
		return
	}
	defer session.Close()

	// Connect pipes
	session.Stdout = &stdout
	stderr, err := session.StderrPipe()
	if err != nil {
		return
	}

	// Run the command
	SSHVerbose(fmt.Sprintf("Running: %s", cmd))
	err = session.Run(cmd)
	if err != nil {
		stderrBuf := new(bytes.Buffer)
		stderrBuf.ReadFrom(stderr)
		err = format(err, &stdout, stderrBuf)
		return
	}
	return
}

func format(err error, stdout, stderr *bytes.Buffer) error {
	if err == nil {
		return err
	}
	msg := "Error during SSH Session"
	if stdout != nil && stdout.String() != "" {
		msg = fmt.Sprintf("%s\n%s", msg, stdout.String())
	}
	if stderr != nil && stderr.String() != "" {
		msg = fmt.Sprintf("%s\n%s", msg, stderr.String())
	}

	return errors.New(msg)
}

func (cl *SecureShellClient) getPublicKey() (authMeth ssh.AuthMethod, err error) {
	// Read priv key file, MUST BE RSA
	SSHVerbose(fmt.Sprintf("Reading private key: %s", cl.privKeyFilename))
	key, err := ioutil.ReadFile(cl.privKeyFilename)
	if err != nil {
		return
	}

	// Parse key
	SSHVerbose("Parsing key")
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return
	}

	// Return pubkey obj
	SSHVerbose("Creating auth method based on key pair")
	authMeth = ssh.PublicKeys(signer)

	return
}

func (cl *SecureShellClient) RunUntil(condition *regexp.Regexp, cmd string, ignoredErrors []string) (err error) {
	// Retry until string condition matches
	for iter := 0; iter < 30; iter++ {
		SSHVerbose(fmt.Sprintf("Try %v", iter))
		// Establish the session
		var session *ssh.Session
		SSHVerbose("Creating session...")
		session, err = cl.conn.NewSession()
		if err != nil {
			return
		}
		defer session.Close()

		// Connect pipes
		var stderr io.Reader
		stderr, err = session.StderrPipe()
		if err != nil {
			return
		}
		// Refresh stdout for every iter
		stdoutBuffer := bytes.Buffer{}
		session.Stdout = &stdoutBuffer

		// Run the command
		SSHVerbose(fmt.Sprintf("Running: %s", cmd))
		err = session.Run(cmd)
		// Ignore specified errors
		if err != nil {
			errMsg := err.Error()
			for _, toIgnore := range ignoredErrors {
				if strings.Contains(errMsg, toIgnore) {
					// ignore error
					SSHVerbose(fmt.Sprintf("Ignored error: %s", errMsg))
					err = nil
					break
				}
			}
		}
		if err != nil {
			stderrBuf := new(bytes.Buffer)
			stderrBuf.ReadFrom(stderr)
			err = format(err, &stdoutBuffer, stderrBuf)
			return
		}
		if condition.MatchString(stdoutBuffer.String()) {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return NewInternalError("Timed out waiting for condition '" + condition.String() + "' with SSH command: " + cmd)
}

func (cl *SecureShellClient) CopyTo(reader io.Reader, destPath, destFilename, permissions string, size int64) error {
	// Check permissions string
	SSHVerbose(fmt.Sprintf("Copying file %s...", destPath+destFilename))
	if !regexp.MustCompile(`\d{4}`).MatchString(permissions) {
		return NewError("Invalid file permission specified: " + permissions)
	}

	// Establish the session
	SSHVerbose("Creating session...")
	session, err := cl.conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Start routine to write file
	errChan := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Instantiate reference to stdin
		remoteStdin, err := session.StdinPipe()
		if err != nil {
			errChan <- err
		}
		defer remoteStdin.Close()

		// Write to stdin
		fmt.Fprintf(remoteStdin, "C%s %d %s\n", permissions, size, destFilename)
		written, err := io.Copy(remoteStdin, reader)
		if err != nil {
			errChan <- err
		}
		if written == 0 {
			errChan <- errors.New("Wrote 0 bytes during file copy to server")
		}
		fmt.Fprint(remoteStdin, "\x00")
	}()

	// Start the scp command
	cmd := "/usr/bin/scp -t "
	SSHVerbose(fmt.Sprintf("Running: %s", cmd+destPath))
	err = session.Run(cmd + destPath)

	// Wait for completion
	wg.Wait()

	// Check for errors
	close(errChan)
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	for copyErr := range errChan {
		if copyErr != nil {
			msg := `%s
%s`
			errMsg = fmt.Sprintf(msg, errMsg, copyErr.Error())
		}
	}
	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func (cl *SecureShellClient) CopyFolderTo(srcPath, destPath, permissions string, recurse bool) error {
	SSHVerbose("Copying folder...")
	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() && recurse {
			// Create the dir if necessary
			if err := cl.CreateFolder(AddTrailingSlash(destPath) + file.Name()); err != nil {
				return err
			}
			// Copy contents of dir
			if err = cl.CopyFolderTo(
				AddTrailingSlash(srcPath)+file.Name(),
				AddTrailingSlash(destPath)+file.Name(),
				permissions,
				true,
			); err != nil {
				return err
			}
		} else {
			// Read the file
			openFile, err := os.Open(AddTrailingSlash(srcPath) + file.Name())
			if err != nil {
				return err
			}
			// Copy the file
			if err := cl.CopyTo(openFile, AddTrailingSlash(destPath), file.Name(), addLeadingZero(permissions), file.Size()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cl *SecureShellClient) CreateFolder(path string) error {
	SSHVerbose(fmt.Sprintf("Creating folder %s", path))
	SSHVerbose(fmt.Sprintf("Running: %s", "mkdir -p "+AddTrailingSlash(path)))
	if _, err := cl.Run("mkdir -p " + AddTrailingSlash(path)); err != nil {
		if strings.Contains(err.Error(), "exists") {
			return nil
		}
		// Retry with sudo
		if strings.Contains(err.Error(), "Permission denied") {
			if _, sudoErr := cl.Run("sudo -S mkdir -p " + AddTrailingSlash(path)); sudoErr != nil {
				if !strings.Contains(sudoErr.Error(), "exists") {
					return sudoErr
				}
			}
		}
		return err
	}
	return nil
}

func addLeadingZero(in string) string {
	if len(in) == 0 {
		return in
	}
	if in[0:0] != "0" {
		in = "0" + in
	}
	return in
}

func AddTrailingSlash(in string) string {
	if len(in) == 0 {
		return in
	}
	if in[len(in)-1:] != "/" {
		in = in + "/"
	}
	return in
}

func SSHVerbose(msg string) {
	if IsDebug() {
		fmt.Println(fmt.Sprintf("[SSH]: %s", msg))
	}
}
