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
	"github.com/eclipse-iofog/iofogctl/v2/internal"

	"github.com/eclipse-iofog/iofogctl/v2/internal/config"
	"github.com/eclipse-iofog/iofogctl/v2/pkg/util"
)

type volumeExecutor struct {
	namespace string
	name      string
	filename  string
}

func newVolumeExecutor(namespace, name, filename string) *volumeExecutor {
	a := &volumeExecutor{}
	a.namespace = namespace
	a.name = name
	a.filename = filename
	return a
}

func (exe *volumeExecutor) GetName() string {
	return exe.name
}

func (exe *volumeExecutor) Execute() (err error) {
	volume, err := config.GetVolume(exe.namespace, exe.name)
	if err != nil {
		return err
	}

	header := config.Header{
		APIVersion: internal.LatestAPIVersion,
		Kind:       config.VolumeKind,
		Metadata: config.HeaderMetadata{
			Namespace: exe.namespace,
			Name:      exe.name,
		},
		Spec: volume,
	}

	if exe.filename == "" {
		if err = util.Print(header); err != nil {
			return err
		}
	} else {
		if err = util.FPrint(header, exe.filename); err != nil {
			return err
		}
	}
	return nil
}
