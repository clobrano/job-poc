/*
Copyright 2023.

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

package controllers

import (
	"context"

	jobpocremediationv1alpha1 "github.com/clobrano/job-poc/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// JobPocRemediationReconciler reconciles a JobPocRemediation object
type JobPocRemediationReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=job-poc-remediation.medik8s.io,resources=jobpocremediations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=job-poc-remediation.medik8s.io,resources=jobpocremediations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=job-poc-remediation.medik8s.io,resources=jobpocremediations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the JobPocRemediation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *JobPocRemediationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("Reconciling JobPocRemediation")
	// get the CR
	var cr jobpocremediationv1alpha1.JobPocRemediation
	if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
		if apierrors.IsNotFound(err) {
			l.Error(err, "unable to fetch JobPocRemediation")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	l.Info("JobPocRemediation fetched", "image", cr.Spec.Image)

	// check if a Job already exists
	key := client.ObjectKey{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}
	var job batchv1.Job
	if err := r.Get(ctx, key, &job); err == nil {
		l.Info("Job already exists", "job", job)
		return ctrl.Result{}, nil
	}

	// Create a Kubernetes Job with the image specified in the CR
	job = batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "job-poc-remediation",
							Image:   cr.Spec.Image,
							Command: cr.Spec.Command,
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}
	if err := r.Create(ctx, &job); err != nil {
		l.Error(err, "Unable to create Job")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobPocRemediationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jobpocremediationv1alpha1.JobPocRemediation{}).
		Complete(r)
}
