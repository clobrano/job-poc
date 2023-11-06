package utils

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var retriesIfPodFails int32 = 0
var isSuspended bool = false

func NewJob(name, namespace, image string, command []string) batchv1.Job {
	// TODO: consider setting the CR as owner of the Job
	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: batchv1.JobSpec{
			Suspend:      &isSuspended,
			BackoffLimit: &retriesIfPodFails,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "job-poc-remediation",
							Image:   image,
							Command: command,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
}
