/*
Copyright 2024.

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
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	compositionv1alpha1 "github.com/cloud-native-compositions/compositions/composition/api/v1alpha1"
	"github.com/cloud-native-compositions/compositions/composition/pkg/crds"
)

// FacadeReconciler reconciles a Facade object
type FacadeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=composition.google.com,resources=facades,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=composition.google.com,resources=facades/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=composition.google.com,resources=facades/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Facade object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *FacadeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Got a new request!", "request", req)

	var facade compositionv1alpha1.Facade
	if err := r.Client.Get(ctx, req.NamespacedName, &facade); err != nil {
		logger.Error(err, "unable to fetch Facade")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Grab status for comparison later
	oldStatus := facade.Status.DeepCopy()

	// Try updating status before returning
	defer func() {
		if !reflect.DeepEqual(oldStatus, facade.Status) {
			newStatus := facade.Status.DeepCopy()
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				nn := types.NamespacedName{Namespace: facade.Namespace, Name: facade.Name}
				err := r.Client.Get(ctx, nn, &facade)
				if err != nil {
					return err
				}
				facade.Status = *newStatus.DeepCopy()
				return r.Client.Status().Update(ctx, &facade)
			})
			if err != nil {
				logger.Error(err, "unable to update Composition status")
			}
		}
	}()

	logger = logger.WithName(facade.Name).WithName(fmt.Sprintf("%d", facade.Generation))

	logger.Info("Validating Facade object")
	if !facade.Validate() {
		logger.Info("Validation Failed")
		return ctrl.Result{}, fmt.Errorf("Validation failed")
	}

	logger.Info("Processing Facade object")
	facade.Status.ClearCondition(compositionv1alpha1.Error)
	if err := r.createCRD(ctx, &facade, logger); err != nil {
		logger.Info("Error creating CRD")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *FacadeReconciler) createCRD(ctx context.Context,
	c *compositionv1alpha1.Facade, logger logr.Logger) error {

	logger = logger.WithName(c.Spec.FacadeKind)

	// Construct Facade CRD from the openAPI Schema
	gvk := schema.GroupVersionKind{
		Kind: c.Spec.FacadeKind,
	}
	crdInfo := crds.NewFacadeCRDInfo(gvk, "", nil, nil, nil)
	err := crdInfo.SetSpec(c.Spec.OpenAPIV3Schema)
	if err == nil {
		err = crdInfo.InstallCRD(ctx, logger, r.Client, r.Scheme)
	} else {
		logger.Error(err, "Unable to set CRD Spec from Schema")
	}

	if err != nil {
		c.Status.Conditions = append(c.Status.Conditions, metav1.Condition{
			LastTransitionTime: metav1.Now(),
			Message:            err.Error(),
			Reason:             "CreateFacadeCRDFailed",
			Type:               string(compositionv1alpha1.Error),
			Status:             metav1.ConditionTrue,
		})
	}
	logger.Info("Created Facade CRD", "crd", crdInfo.Name())
	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *FacadeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compositionv1alpha1.Facade{}).
		Complete(r)
}
