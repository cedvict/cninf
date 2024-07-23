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

	// Go imports
	"fmt"
	"golang.org/x/exp/slices"
	"reflect"

	// AWS SDK imports
	// "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	// Kubernetes object imports
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// Kubernetes imports
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cninfv1 "github.com/cedvict/cninf.git/api/v1"
)

func Remove[T any](slice []T, element T) []T {
	// Iterate through the slice and create a new slice without the element
	var result []T
	for _, item := range slice {
		if !reflect.DeepEqual(item, element) {
			result = append(result, item)
		}
	}
	return result
}

func RemoveComparable[T comparable](slice []T, element T) []T {
	var result []T
	for _, item := range slice {
		if item != element {
			result = append(result, item)
		}
	}
	return result
}

const (
	configMapName = "%s-configmap"
	finalizerName = "stores.cninf.uman.test/finalizer"
)

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	S3svc  *s3.S3
}

// +kubebuilder:rbac:groups=cninf.uman.test,resources=stores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cninf.uman.test,resources=stores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cninf.uman.test,resources=stores/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Store object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logCtx := log.FromContext(ctx)

	instance := &cninfv1.Store{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if client.IgnoreNotFound(err) == nil {
			logCtx.Info("Store resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logCtx.Error(err, "Failed to get Store")
		return ctrl.Result{}, err

	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		if instance.Status.State == "" {
			instance.Status.State = cninfv1.PENDING_STATE
			instance.Status.Message = "Trying to create Store"
			if err := r.Status().Update(ctx, instance); err != nil {
				logCtx.Error(err, "Failed to update Store status")
				return ctrl.Result{}, err
			}
		}

		if !slices.Contains(instance.GetFinalizers(), finalizerName) {
			instance.SetFinalizers(append(instance.GetFinalizers(), finalizerName))
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}

		if instance.Status.State == cninfv1.PENDING_STATE {
			if err := r.CreateResources(ctx, instance); err != nil {
				instance.Status.State = cninfv1.ERROR_STATE
				instance.Status.Message = err.Error()
				logCtx.Error(err, "Error creating resources")
				if err := r.Status().Update(ctx, instance); err != nil {
					logCtx.Error(err, "Failed to update Store status")
					return ctrl.Result{}, err
				}
				return ctrl.Result{}, err
			}

		}
	} else {
		if err := r.DeleteResources(ctx, instance); err != nil {
			instance.Status.State = cninfv1.ERROR_STATE
			instance.Status.Message = err.Error()
			logCtx.Error(err, "Error deleting resources")
			if err := r.Status().Update(ctx, instance); err != nil {
				logCtx.Error(err, "Failed to update Store status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		if slices.Contains(instance.GetFinalizers(), finalizerName) {
			instance.SetFinalizers(RemoveComparable(instance.GetFinalizers(), finalizerName))
			if err := r.Update(ctx, instance); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil

	}
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cninfv1.Store{}).
		Complete(r)
}

// CreateResources creates the resources for the Store
func (r *StoreReconciler) CreateResources(ctx context.Context, instance *cninfv1.Store) error {
	// Add the status first
	instance.Status.State = cninfv1.CREATING_STATE
	instance.Status.Message = "Creating storage"
	if err := r.Status().Update(ctx, instance); err != nil {
		return err
	}

	// Create the input for the request
	bucketName := fmt.Sprintf("%s-%s", instance.Namespace, instance.Spec.Name)
	if instance.Spec.Shared {
		bucketName = instance.Spec.Name
		instance.Spec.Locked = true
	}
	/*input := &s3.CreateBucketInput{
		Bucket:                     aws.String(bucketName),
		ObjectLockEnabledForBucket: aws.Bool(instance.Spec.Locked),
	}

	// Create the bucket
	bucket, err := r.S3svc.CreateBucket(input)
	if err != nil {
		return err
	}

	// Wait for the bucket to be created
	err = r.S3svc.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(instance.Spec.Name),
	})
	if err != nil {
		return err
	}*/

	// Create the configmap
	configMap := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(configMapName, instance.Spec.Name),
			Namespace: instance.Namespace,
		},
		Data: map[string]string{
			"bucket": bucketName,
			//"location": *bucket.Location,
			"location": "",
		},
	}

	if err := r.Create(ctx, configMap); err != nil {
		return err
	}

	// Update the status
	instance.Status.State = cninfv1.CREATED_STATE
	instance.Status.Message = "Storage created"
	if err := r.Status().Update(ctx, instance); err != nil {
		return err
	}

	return nil
}

// DeleteResources deletes the resources for the Store
func (r *StoreReconciler) DeleteResources(ctx context.Context, instance *cninfv1.Store) error {
	// Add the status first
	instance.Status.State = cninfv1.DELETING_STATE
	instance.Status.Message = "Deleting storage"
	if err := r.Status().Update(ctx, instance); err != nil {
		return err
	}

	// Delete the bucket
	/*bucketName := fmt.Sprintf("%s-%s", instance.Namespace, instance.Spec.Name)
	if instance.Spec.Shared {
		bucketName = instance.Spec.Name
	}
	_, err := r.S3svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		_, err = r.S3svc.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			return err
		}

	}*/

	// Delete the configmap
	configMap := &v1.ConfigMap{}
	err := r.Get(ctx, client.ObjectKey{
		Namespace: instance.Namespace,
		Name:      fmt.Sprintf(configMapName, instance.Spec.Name),
	}, configMap)
	if err != nil {
		return err
	}

	if err := r.Delete(ctx, configMap); err != nil {
		return err
	}

	// Update the status
	instance.Status.State = cninfv1.DELETED_STATE
	instance.Status.Message = "Storage deleted"
	if err := r.Status().Update(ctx, instance); err != nil {
		return err
	}

	return nil
}
