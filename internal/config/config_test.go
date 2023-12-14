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

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"
)

func TestDefaultConfig(t *testing.T) {
	o := NewWithT(t)
	config := NewConfig()
	config.Init()

	o.Expect(config.GetString(MetricsBindAddressKey)).To(Equal(":8080"))
	o.Expect(config.GetString(HealthzBindAddressKey)).To(Equal(":8081"))
	o.Expect(config.GetBool(LeaderElectionKey)).To(Equal(false))
	o.Expect(config.GetString(CloudEventsSourceURIKey)).To(Equal("localhost"))
	o.Expect(config.GetString(CloudEventsTargetAddressKey)).To(Equal("http://localhost:8082"))
}

var fullYaml = `
metrics:
  bind-address: ":9000"
healthz:
  bind-address: ":9001"
leader-election: true
cloud-events:
  source-uri: https://github.com/kubarchive/dynowatch
  target-address: https://splunk.mycorp.com/events
watches:
  - name: deployments
    group: apps
    version: v1
    kind: Deployment
  - name: jobs
    group: batch
    version: v1
    kind: Job
`

func TestReadConfig(t *testing.T) {
	o := NewWithT(t)
	config := NewConfig()
	config.Init()

	o.Expect(config.ReadConfig(bytes.NewBufferString(fullYaml))).To(Succeed())

	o.Expect(config.GetString(MetricsBindAddressKey)).To(Equal(":9000"))
	o.Expect(config.GetString(HealthzBindAddressKey)).To(Equal(":9001"))
	o.Expect(config.GetBool(LeaderElectionKey)).To(Equal(true))
	o.Expect(config.GetString(CloudEventsSourceURIKey)).To(Equal("https://github.com/kubarchive/dynowatch"))
	o.Expect(config.GetString(CloudEventsTargetAddressKey)).To(Equal("https://splunk.mycorp.com/events"))
	o.Expect(config.Get(ObjectWatchesKey)).ToNot(BeEmpty())

	watches, err := config.GetWatches()
	o.Expect(err).NotTo(HaveOccurred())
	o.Expect(watches).ToNot(BeEmpty())

	expected := []Watch{
		{
			Name:    "deployments",
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		{
			Name:    "jobs",
			Group:   "batch",
			Version: "v1",
			Kind:    "Job",
		},
	}
	o.Expect(watches).To(BeEquivalentTo(expected))
}

func TestGetWatches(t *testing.T) {
	o := NewWithT(t)
	config := NewConfig()
	config.Init()

	o.Expect(config.GetWatches()).To(BeEmpty())
	watchYaml := `
watches:
  - name: deployments
    group: apps
    version: v1
    kind: Deployment
  - name: jobs
    group: batch
    version: v1
    kind: Job`
	o.Expect(config.ReadConfig(bytes.NewBufferString(watchYaml))).To(Succeed())
	watches, err := config.GetWatches()
	o.Expect(err).NotTo(HaveOccurred())
	o.Expect(watches).ToNot(BeEmpty())

	expected := []Watch{
		{
			Name:    "deployments",
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		{
			Name:    "jobs",
			Group:   "batch",
			Version: "v1",
			Kind:    "Job",
		},
	}
	o.Expect(watches).To(BeEquivalentTo(expected))
}
