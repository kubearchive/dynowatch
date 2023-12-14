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

package controller

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("dynamic reconciler", func() {

	BeforeEach(func() {
		Expect(testServer.GetEvents()).Should(BeEmpty())
	})

	AfterEach(func() {
		testServer.StopRecorder()
		testServer.ClearEvents()
	})

	It("sends a CloudEvent when a Job is created", func(ctx SpecContext) {
		testServer.StartRecorder()
		defer testServer.StopRecorder()
		By("creating a Job object")
		job := createJobFixture("default", "created-job")
		Expect(k8sClient.Create(ctx, job)).Should(Succeed())
		Eventually(ctx, testServer.GetEvents).ShouldNot(BeEmpty())
	})

	It("sends a CloudEvent when a Job is updated", func(ctx SpecContext) {
		By("creating a Job object")
		job := createJobFixture("default", "updated-job")
		Expect(k8sClient.Create(ctx, job)).Should(Succeed(), "create job fixture")
		Expect(testServer.GetEvents()).Should(BeEmpty(), "events are not recorded")
		testServer.StartRecorder()
		defer testServer.StopRecorder()
		// Simulate that a pod is running
		By("updating a Job object")
		now := metav1.Now()
		start := metav1.NewTime(now.Add(-1 * time.Second))
		job.Status = batchv1.JobStatus{
			Conditions: []batchv1.JobCondition{
				{
					Type:               batchv1.JobComplete,
					Status:             corev1.ConditionUnknown,
					Reason:             "Running",
					LastProbeTime:      now,
					LastTransitionTime: now,
				},
			},
			Active:    1,
			StartTime: &start,
		}
		Expect(k8sClient.Status().Update(ctx, job)).Should(Succeed(), "update job fixture")
		Eventually(ctx, testServer.GetEvents).ShouldNot(BeEmpty())
	})

	It("sends a CloudEvent when a Job is deleted", func(ctx SpecContext) {
		By("creating a Job object")
		job := createJobFixture("default", "deleted-job")
		Expect(k8sClient.Create(ctx, job)).Should(Succeed(), "create job fixture")
		Expect(testServer.GetEvents()).Should(BeEmpty(), "events are not recorded")
		testServer.StartRecorder()
		defer testServer.StopRecorder()
		By("deleting a Job object")
		Expect(k8sClient.Delete(ctx, job)).Should(Succeed(), "delete job fixture")
		Eventually(ctx, testServer.GetEvents).ShouldNot(BeEmpty())
	})
})

func createJobFixture(namespace string, name string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "hello",
							Image: "busybox:1.28",
							Command: []string{
								"/bin/sh",
								"-c",
								"date; echo Hello world",
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyOnFailure,
				},
			},
		},
	}
}
