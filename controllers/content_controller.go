/*
Copyright 2019 Bob Tribit.

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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	filterv1alpha1 "tribit.io/content-filter-controller/api/v1alpha1"
)

type Data interface{}

type Monad func(error) (Data, error)

func Get(d Data) Monad {
	return func(e error) (Data, error) {
		return d, e
	}
}

func Next(m Monad, f func(Data) Monad) Monad {
	return func(e error) (Data, error) {
		newData, newError := m(e)
		if newError != nil {
			return nil, newError
		}
		return f(newData)(newError)
	}
}

// ContentReconciler reconciles a Content object
type ContentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=filter.tribit.io,resources=contents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=filter.tribit.io,resources=contents/status,verbs=get;update;patch

func (r *ContentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.Log.WithValues("content", req.NamespacedName)

	contentFilter := &filterv1alpha1.Content{}

	r.Client.Get(context.Background(), req.NamespacedName, contentFilter)

	if contentFilter.Status.Provisioned {
		return ctrl.Result{}, nil
	}

	// your logic here
	step := Get(contentFilter)
	step = Next(step, updateStatus)
	_, err := step(nil)

	if err != nil {
		logger.Error(err, "failed to reconcile content filter")
	}

	err = r.Client.Status().Update(context.Background(), contentFilter)

	if err != nil {
		logger.Error(err, "failed to update content filter")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{Requeue: true}, nil
}

func updateStatus(d Data) Monad {
	pipe := d.(*filterv1alpha1.Content)
	return func(e error) (Data, error) {

		pipe.Status.Provisioned = true
		return pipe, e
	}
}

func (r *ContentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&filterv1alpha1.Content{}).
		Complete(r)
}
