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

package manager

import (
	"testing"

	. "github.com/onsi/gomega"

	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/kubearchive/dynowatch/internal/config"
)

func TestSetupControllers(t *testing.T) {
	o := NewWithT(t)
	watches := []config.Watch{
		{
			Name:    "deployments",
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		},
		{
			Name:    "pipelineruns",
			Group:   "tekton.dev",
			Version: "v1",
			Kind:    "PipelineRun",
		},
	}
	restConfig := &rest.Config{}
	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{})
	o.Expect(err).NotTo(HaveOccurred())
	o.Expect(SetupControllers(mgr, nil, watches, "localhost", "https://splunk.mycompany.com")).To(Succeed())
}
