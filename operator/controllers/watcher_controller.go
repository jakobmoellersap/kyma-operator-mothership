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
	"bytes"
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// WatcherReconciler reconciles a Watcher object
type WatcherReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	WatcherClient *http.Client
}

//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=watchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=watchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inventory.kyma-project.io,resources=watchers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Watcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *WatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	newConfigMap := &corev1.ConfigMap{}
	err := r.Get(ctx, req.NamespacedName, newConfigMap)
	if err != nil {
		logger.Info("Configmap deleted:", "name", req.Name)
	}
	runtimeID, found := newConfigMap.Data["RuntimeID"]
	if !found {
		logger.Error(err, "Configmap invalid: no runtimeID", "name", req.Name)
		return ctrl.Result{}, nil
	}
	componentStatus, found := newConfigMap.Data["ComponentStatus"]
	if !found {
		logger.Error(err, "Configmap invalid: no ComponentStatus", "name", req.Name)
		return ctrl.Result{}, nil
	}
	clusterInfo := &ClusterInfo{RuntimeID: runtimeID, ComponentStatus: componentStatus}
	logger.Info("Configmap updated:", "name", newConfigMap.Name)
	clusterInfoPayload, err := json.Marshal(clusterInfo)
	if err != nil {
		logger.Error(err, "Can't create clusterInfo payload", "name", newConfigMap.Name)
	}
	_, err = r.WatcherClient.Post("http://localhost:8090/callback", "application/json", bytes.NewBuffer(clusterInfoPayload))
	if err != nil {
		logger.Error(err, "Can't send callback", "name", newConfigMap.Name)
	}

	return ctrl.Result{}, nil
}

type ComponentStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ClusterInfo struct {
	RuntimeID string `json:"runtime_id"`
	//ComponentStatus []*ComponentStatus `json:"component_status"`
	ComponentStatus string `json:"component_status"`
}

// SetupWithManager sets up the controller with the Manager.
func (r *WatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		WithEventFilter(labelFilterPredicate()).
		Complete(r)
}

func labelFilterPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			labels := e.ObjectNew.GetLabels()
			if isWatchedByMothership(labels) {
				return e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion()
			}
			return false
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return false
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			labels := deleteEvent.Object.GetLabels()
			return isWatchedByMothership(labels)
		},
		CreateFunc: func(createEvent event.CreateEvent) bool {
			labels := createEvent.Object.GetLabels()
			return isWatchedByMothership(labels)
		},
	}
}

func isWatchedByMothership(labels map[string]string) bool {
	value, found := labels["kyma-project.io/watched-by"]
	return found && value == "mothership"
}
