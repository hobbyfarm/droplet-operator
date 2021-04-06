/*


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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dropletv1alpha1 "github.com/ibrokethecloud/droplet-operator/pkg/api/v1alpha1"
	"github.com/ibrokethecloud/droplet-operator/pkg/do"
	corev1 "k8s.io/api/core/v1"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	Finalizer = "droplet.cattle.io"
)

// +kubebuilder:rbac:groups=droplet.cattle.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=droplet.cattle.io,resources=instances/status,verbs=get;update;patch

func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("instance", req.NamespacedName)
	var requeue bool
	instance := &dropletv1alpha1.Instance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "unable to fetch instance")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secret, err := r.secretExists(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	doClient, err := do.NewClient(secret)
	if err != nil {
		return ctrl.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// launch/modify instance
		status := instance.Status.DeepCopy()
		switch status.Status {
		case "":
			log.Info("Creating instance")
			status, err = doClient.CreateDroplet(ctx, instance)
		case do.Submitted:
			log.Info("Fetch details")
			status, err = doClient.FetchDetails(ctx, instance)
		case do.Provisioned:
			log.Info("Instance Provisioned")
			return ctrl.Result{}, nil
		}

		if err != nil {
			status.Message = err.Error()
		}
		instance.Status = *status
		requeue = true
		controllerutil.AddFinalizer(instance, Finalizer)
	} else {
		// clean up instance
		if controllerutil.ContainsFinalizer(instance, Finalizer) {
			log.Info("Terminating instance")
			controllerutil.RemoveFinalizer(instance, Finalizer)
		}
		err = doClient.DeleteInstance(ctx, instance)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{Requeue: requeue}, r.Update(ctx, instance)
}

func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dropletv1alpha1.Instance{}).
		Complete(r)
}

func (r *InstanceReconciler) secretExists(ctx context.Context, instance *dropletv1alpha1.Instance) (secret *corev1.Secret, err error) {
	secret = &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: *instance.Spec.Secret}, secret)
	return secret, err
}
