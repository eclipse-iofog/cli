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

package get

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v2/pkg/client"
	rsc "github.com/eclipse-iofog/iofogctl/v2/internal/resource"
	iutil "github.com/eclipse-iofog/iofogctl/v2/internal/util"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type applicationExecutor struct {
	namespace    string
	client       *client.Client
	flows        []client.FlowInfo
	msvcsPerFlow map[int][]client.MicroserviceInfo
}

func newApplicationExecutor(namespace string) *applicationExecutor {
	c := &applicationExecutor{}
	c.namespace = namespace
	c.msvcsPerFlow = make(map[int][]client.MicroserviceInfo)
	return c
}

func (exe *applicationExecutor) GetName() string {
	return ""
}

func (exe *applicationExecutor) Execute() error {
	// Fetch data
	if err := exe.init(); err != nil {
		return err
	}
	printNamespace(exe.namespace)
	table, err := exe.generateApplicationOutput()
	if err != nil {
		return err
	}
	return print(table)
}

func (exe *applicationExecutor) init() (err error) {
	exe.client, err = iutil.NewControllerClient(exe.namespace)
	if err != nil {
		if rsc.IsNoControlPlaneError(err) {
			return nil
		}
		return err
	}
	flows, err := exe.client.GetAllFlows()
	if err != nil {
		return
	}
	exe.flows = flows.Flows
	for _, flow := range exe.flows {
		listMsvcs, err := exe.client.GetMicroservicesPerFlow(flow.ID)
		if err != nil {
			return err
		}

		// Filter System microservices
		for _, ms := range listMsvcs.Microservices {
			if util.IsSystemMsvc(ms) {
				continue
			}
			exe.msvcsPerFlow[flow.ID] = append(exe.msvcsPerFlow[flow.ID], ms)
		}
	}
	return
}

func (exe *applicationExecutor) generateApplicationOutput() (table [][]string, err error) {
	// Generate table and headers
	table = make([][]string, len(exe.flows)+1)
	headers := []string{"APPLICATION", "RUNNING", "MICROSERVICES"}
	table[0] = append(table[0], headers...)

	// Populate rows
	for idx, flow := range exe.flows {
		nbMsvcs := len(exe.msvcsPerFlow[flow.ID])
		runningMsvcs := 0
		msvcs := ""
		first := true
		for _, msvc := range exe.msvcsPerFlow[flow.ID] {
			if first == true {
				msvcs += fmt.Sprintf("%s", msvc.Name)
			} else {
				msvcs += fmt.Sprintf(", %s", msvc.Name)
			}
			first = false
			if msvc.Status.Status == "RUNNING" {
				runningMsvcs++
			}
		}

		if nbMsvcs > 5 {
			msvcs = fmt.Sprintf("%d microservices", len(exe.msvcsPerFlow[flow.ID]))
		}

		status := fmt.Sprintf("%d/%d", runningMsvcs, nbMsvcs)

		row := []string{
			flow.Name,
			status,
			msvcs,
		}
		table[idx+1] = append(table[idx+1], row...)
	}

	return
}
