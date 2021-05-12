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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"context"

	dropletv1alpha1 "github.com/hobbyfarm/droplet-operator/pkg/api/v1alpha1"
	"github.com/hobbyfarm/droplet-operator/pkg/do"
)

// ImportKeyPairReconciler reconciles a ImportKeyPair object
type ImportKeyPairReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=droplet.cattle.io,resources=importkeypairs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=droplet.cattle.io,resources=importkeypairs/status,verbs=get;update;patch

func (r *ImportKeyPairReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("importkeypair", req.NamespacedName)
	key := &dropletv1alpha1.ImportKeyPair{}
	var requeue bool
	if err := r.Get(ctx, req.NamespacedName, key); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "unable to fetch key")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	nsSecretName := types.NamespacedName{Namespace: key.Namespace, Name: key.Spec.Secret}
	secret, err := secretExists(r.Client, ctx, nsSecretName)
	if err != nil {
		return ctrl.Result{}, err
	}

	doClient, err := do.NewClient(secret)
	if err != nil {
		return ctrl.Result{}, nil
	}

	if key.ObjectMeta.DeletionTimestamp.IsZero() {
		status := key.Status.DeepCopy()
		switch status.Status {
		case "":
			log.Info("Creating KeyPair")
			status, err = doClient.CreateKeyPair(ctx, key)
		case do.Provisioned:
			log.Info("ImportKey Reconcilled")
			return ctrl.Result{}, nil
		}

		if err != nil {
			status.Message = err.Error()
		}
		key.Status = *status
		requeue = true
		controllerutil.AddFinalizer(key, Finalizer)
	} else {
		if controllerutil.ContainsFinalizer(key, Finalizer) {
			log.Info("Removing key")
			err = doClient.RemoveKeyPair(ctx, key)
			if err != nil {
				return ctrl.Result{}, err
			}
			controllerutil.RemoveFinalizer(key, Finalizer)
		}

	}

	return ctrl.Result{Requeue: requeue}, r.Update(ctx, key)
}

func (r *ImportKeyPairReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dropletv1alpha1.ImportKeyPair{}).
		Complete(r)
}
