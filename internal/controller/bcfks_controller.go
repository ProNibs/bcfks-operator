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

// cSpell:words bcfks,bcfksiov,corev,metav

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bcfksiov1alpha1 "github.com/ProNibs/bcfks-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BCFKSReconciler reconciles a BCFKS object
type BCFKSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=bcfks.io,resources=bcfks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bcfks.io,resources=bcfks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bcfks.io,resources=bcfks/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the BCFKS object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *BCFKSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch BCFKS instance
	log.Info("Starting fetch...")
	bcfks := &bcfksiov1alpha1.BCFKS{}
	err := r.Get(ctx, req.NamespacedName, bcfks)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("BCFKS resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get BCFKS object")
		return ctrl.Result{}, err
	}
	// Check if Secret exists, if not deploy one
	found := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: bcfks.Name, Namespace: bcfks.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new secret
		sec := r.secretForBCFKS(bcfks)
		log.Info("Creating a new Secret", "Secret.Namespace", sec.Namespace, "Secret.Name", sec.Name)
		err = r.Create(ctx, sec)
		if err != nil {
			log.Error(err, "Failed to create new Secret", "Secret.Namespace", sec.Namespace, "Secret.Name", sec.Name)
			return ctrl.Result{}, err
		}
		// Secret created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

func (r *BCFKSReconciler) secretForBCFKS(m *bcfksiov1alpha1.BCFKS) *corev1.Secret {
	secretName := m.Spec.SecretName

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: m.Namespace,
		},
		StringData: map[string]string{
			"test": "testing",
		},
		Type: "Opaque",
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(m, secret, r.Scheme)
	return secret
}

// SetupWithManager sets up the controller with the Manager.
func (r *BCFKSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bcfksiov1alpha1.BCFKS{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
