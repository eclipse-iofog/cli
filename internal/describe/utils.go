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

package describe

import (
	"encoding/json"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog/client"
)

func MapClientMicroserviceToConfigMicroservice(msvc *client.MicroserviceInfo, clt *client.Client) (result *config.Microservice, err error) {
	agent, err := clt.GetAgentByID(msvc.AgentUUID)
	if err != nil {
		return
	}
	catalogItem, err := clt.GetCatalogItem(msvc.CatalogItemID)
	if err != nil {
		return
	}
	jsonConfig := make(map[string]interface{})
	if err = json.Unmarshal([]byte(msvc.Config), &jsonConfig); err != nil {
		return
	}
	result = new(config.Microservice)
	result.UUID = msvc.UUID
	result.Name = msvc.Name
	result.Agent = config.MicroserviceAgent{
		Name: agent.Name,
		Config: client.AgentConfiguration{
			DockerURL:                 &agent.DockerURL,
			DiskLimit:                 &agent.DiskLimit,
			DiskDirectory:             &agent.DiskDirectory,
			MemoryLimit:               &agent.MemoryLimit,
			CPULimit:                  &agent.CPULimit,
			LogLimit:                  &agent.LogLimit,
			LogDirectory:              &agent.LogDirectory,
			LogFileCount:              &agent.LogFileCount,
			StatusFrequency:           &agent.StatusFrequency,
			ChangeFrequency:           &agent.ChangeFrequency,
			DeviceScanFrequency:       &agent.DeviceScanFrequency,
			BluetoothEnabled:          &agent.BluetoothEnabled,
			WatchdogEnabled:           &agent.WatchdogEnabled,
			AbstractedHardwareEnabled: &agent.AbstractedHardwareEnabled,
		},
	}
	images := config.MicroserviceImages{
		CatalogID: catalogItem.ID,
		Registry:  catalogItem.RegistryID,
	}
	for _, img := range catalogItem.Images {
		if img.AgentTypeID == 1 {
			images.X86 = img.ContainerImage
		} else if img.AgentTypeID == 2 {
			images.ARM = img.ContainerImage
		}
	}
	result.Images = images
	result.Config = jsonConfig
	result.RootHostAccess = msvc.RootHostAccess
	result.Ports = msvc.Ports
	result.Volumes = msvc.Volumes
	result.Routes = msvc.Routes
	result.Env = msvc.Env
	result.Flow = &msvc.FlowID
	return
}
