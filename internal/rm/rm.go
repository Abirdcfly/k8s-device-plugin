/*
 * Copyright (c) 2019-2022, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY Type, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rm

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	spec "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
)

var _ ResourceManager = (*resourceManager)(nil)

// resourceManager implements the ResourceManager interface
type resourceManager struct {
	config   *spec.Config
	resource string
	devices  []*Device
}

// ResourceManager provides an interface for listing a set of Devices and checking health on them
type ResourceManager interface {
	Resource() string
	Devices() []*Device
	CheckHealth(stop <-chan interface{}, devices []*Device, unhealthy chan<- *Device) error
}

// NewResourceManagers returns a []ResourceManager, one for each resource in 'config'.
func NewResourceManagers(config *spec.Config) ([]ResourceManager, error) {
	nvml.Init()
	defer nvml.Shutdown()

	deviceMap, err := buildDeviceMap(config)
	if err != nil {
		return nil, fmt.Errorf("error building device map: %v", err)
	}

	var rms []ResourceManager
	for resourceName, devices := range deviceMap {
		r := &resourceManager{
			config:   config,
			resource: spec.AddResourceNamePrefix(resourceName),
			devices:  devices,
		}
		if len(r.Devices()) != 0 {
			rms = append(rms, r)
		}
	}

	return rms, nil
}

// Resource gets the resource name associated with the ResourceManager
func (r *resourceManager) Resource() string {
	return r.resource
}

// Resource gets the devices managed by the ResourceManager
func (r *resourceManager) Devices() []*Device {
	return r.devices
}

// CheckHealth performs health checks on a set of devices, writing to the 'unhealthy' channel with any unhealthy devices
func (r *resourceManager) CheckHealth(stop <-chan interface{}, devices []*Device, unhealthy chan<- *Device) error {
	return r.checkHealth(stop, devices, unhealthy)
}
