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

package execute

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/eclipse-iofog/iofogctl/v2/internal"
	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
	"gopkg.in/yaml.v2"
)

type emptyExecutor struct {
	name string
}

func (exe *emptyExecutor) Execute() error {
	return nil
}
func (exe *emptyExecutor) GetName() string {
	return exe.name
}

// NewEmptyExecutor return an executor that does nothing
func NewEmptyExecutor(name string) Executor {
	return &emptyExecutor{
		name: name,
	}
}

func generateExecutor(header config.Header, namespace string, kindHandlers map[config.Kind]func(KindHandlerOpt) (Executor, error)) (exe Executor, err error) {
	// Check namespace exists
	if len(header.Metadata.Namespace) > 0 {
		namespace = header.Metadata.Namespace
	}

	if _, err := config.GetNamespace(namespace); err != nil {
		return nil, err
	}

	if err = internal.ValidateHeader(header); err != nil {
		return nil, err
	}

	subYamlBytes, err := yaml.Marshal(header.Spec)
	if err != nil {
		return exe, err
	}

	createExecutorFunc, found := kindHandlers[header.Kind]
	if !found {
		util.PrintNotify(fmt.Sprintf("Could not handle kind %s. Skipping document\n", header.Kind))
		return nil, nil
	}

	return createExecutorFunc(KindHandlerOpt{
		Kind:      header.Kind,
		Namespace: namespace,
		Name:      header.Metadata.Name,
		YAML:      subYamlBytes,
	})
}

type KindHandlerOpt struct {
	Kind      config.Kind
	Namespace string
	Name      string
	YAML      []byte
}

func GetExecutorsFromYAML(inputFile, namespace string, kindHandlers map[config.Kind]func(KindHandlerOpt) (Executor, error)) (executorsMap map[config.Kind][]Executor, err error) {
	yamlFile, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return
	}

	r := bytes.NewReader(yamlFile)
	dec := yaml.NewDecoder(r)
	dec.SetStrict(true)

	var raw yaml.MapSlice
	header := config.Header{
		Spec: raw,
	}

	// Generate all executors
	executorsMap = make(map[config.Kind][]Executor)
	decodeErr := dec.Decode(&header)
	for decodeErr == nil {
		exe, err := generateExecutor(header, namespace, kindHandlers)
		if err != nil {
			return nil, err
		}
		if exe != nil {
			executorsMap[header.Kind] = append(executorsMap[header.Kind], exe)
		}
		decodeErr = dec.Decode(&header)
	}
	if decodeErr != io.EOF && decodeErr != nil {
		return nil, decodeErr
	}

	return
}
