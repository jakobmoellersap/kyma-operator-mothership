/*
Copyright 2022.

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
	istioOperatorApi "github.com/Tomasz-Smelcerz-SAP/kyma-operator-istio/k8s-api/api/v1alpha1"
	inventoryv1alpha1 "github.com/Tomasz-Smelcerz-SAP/kyma-operator-mothership/operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// KymaReconciler reconciles a Kyma object
type KymaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=kymas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=kymas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=kymas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Kyma object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *KymaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	obj := inventoryv1alpha1.Kyma{}
	err := r.Client.Get(ctx, req.NamespacedName, &obj)
	if apierrors.IsNotFound(err) {
		//object is deleted
		logger.Info("Object is deleted:", "object", req.NamespacedName)

		//try to delete related IstioConfiguration object
		istioObject := istioOperatorApi.IstioConfiguration{}
		istioObject.Name = req.Name
		istioObject.Namespace = req.Namespace

		err = r.Client.Delete(ctx, &istioObject)
		if apierrors.IsNotFound(err) {
			//IstioConfiguration does not exist. Success.
			return ctrl.Result{}, nil
		}
		if err != nil {
			logger.Error(err, "Error during IstioConfiguration delete")
			return ctrl.Result{}, err
		}
		logger.Info("Successfully deleted IstioConfiguration:", "object:", istioObject)

		return ctrl.Result{}, nil
	}

	if err != nil {
		logger.Error(err, "Error during reconciliation")
		return ctrl.Result{}, err
	}

	importantValue := obj.Spec.Foo

	istioObject := istioOperatorApi.IstioConfiguration{}
	istioObject.Name = obj.Name
	istioObject.Namespace = obj.Namespace

	istioObject.Spec = istioOperatorApi.IstioConfigurationSpec{}
	istioObject.Spec.Foo = importantValue + "_from_mothership"

	err = r.Client.Create(ctx, &istioObject)
	if err != nil {
		logger.Error(err, "Error creating IstioConfiguration")
		return ctrl.Result{}, err
	}
	logger.Info("Successfully created IstioConfiguration:", "object:", istioObject)

	logger.Info("Successfully reconciled Kyma:", "object:", obj)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KymaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inventoryv1alpha1.Kyma{}).
		Complete(r)
}
