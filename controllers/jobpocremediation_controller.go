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
	"fmt"
	"time"

	jobpocremediationv1alpha1 "github.com/clobrano/job-poc/api/v1alpha1"
	utils "github.com/clobrano/job-poc/pkg"
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// JobPocRemediationReconciler reconciles a JobPocRemediation object
type JobPocRemediationReconciler struct {
	client.Client
	Log    logr.Logger
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
	l := r.Log.WithName("jobpocremediation_controller")

	l.Info("Reconciling JobPocRemediation")
	var cr jobpocremediationv1alpha1.JobPocRemediation
	if err := r.Get(ctx, req.NamespacedName, &cr); err != nil {
		if !apierrors.IsNotFound(err) {
			l.Error(err, "unable to fetch JobPocRemediation")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	l.Info("JobPocRemediation fetched", "Name", cr.Name, "Namespace", cr.Namespace, "Image", cr.Spec.Image, "Command", cr.Spec.Command)

	key := client.ObjectKey{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}
	var job batchv1.Job
	if err := r.Get(ctx, key, &job); err != nil {
		if !apierrors.IsNotFound(err) {
			l.Error(err, "unable to fetch Job")
			return ctrl.Result{}, err
		}

		// Job does not exist, create a new one
		job = utils.NewJob(cr.Name, cr.Namespace, cr.Spec.Image, cr.Spec.Command)
		if err := r.Create(ctx, &job); err != nil {
			l.Error(err, "Unable to create Job")
			return ctrl.Result{}, nil
		}
		l.Info("Job created")
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	l.Info("Job exists, check status", "Active", job.Status.Active, "Succeeded", job.Status.Succeeded, "Failed", job.Status.Failed, "Backofflimit", job.Spec.BackoffLimit)

	if job.Status.Active == 0 && job.Status.Succeeded == 0 && job.Status.Failed == 0 {
		l.Info("Job hasn't started yet")
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	if job.Status.Active > 0 {
		if job.Status.Failed > 0 {
			l.Info("Job has failed")
			return ctrl.Result{}, fmt.Errorf("Job has failed with error")
		}

		l.Info("Job is still running")
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}

	if job.Status.Succeeded > 0 {
		l.Info("Job has succeeded")

		err := r.Delete(ctx, &job, client.PropagationPolicy(metav1.DeletePropagationBackground))
		if err != nil {
			l.Error(err, "Unable to delete Job")
		}
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, fmt.Errorf("Job has failed with error")
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobPocRemediationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&jobpocremediationv1alpha1.JobPocRemediation{}).
		Complete(r)
}
