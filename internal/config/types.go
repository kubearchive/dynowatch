/*
Copyright 2023 The KubeArchive Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

type DynowatchConfig struct {
	CloudEvents    CloudEventConfig `json:"cloud-events,omitempty"`
	Healthz        Healthz          `json:"healthz,omitempty"`
	LeaderElection bool             `json:"leader-election,omitempty"`
	Metrics        Metrics          `json:"metrics,omitempty"`
	Watches        []Watch          `json:"watches,omitempty"`
}

type CloudEventConfig struct {
	SourceURI     string `json:"source-uri,omitempty"`
	TargetAddress string `json:"target-address,omitempty"`
}

type Healthz struct {
	BindAddress string `json:"bind-address,omitempty"`
}

type Metrics struct {
	BindAddress string `json:"bind-address,omitempty"`
}

type Watch struct {
	Name    string `json:"name"`
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}
