/*
Copyright 2025.

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

// in internal/controller/ofbiz_controller.go

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ofbizv1alpha1 "github.com/halims/ofbiz-operator/api/v1alpha1"
)

// OfbizReconciler reconciles a Ofbiz object
type OfbizReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ofbiz.ofbiz.apache.org,resources=ofbizzes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ofbiz.ofbiz.apache.org,resources=ofbizzes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ofbiz.ofbiz.apache.org,resources=ofbizzes/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods;services;secrets;configmaps;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

func (r *OfbizReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Ofbiz instance
	ofbiz := &ofbizv1alpha1.Ofbiz{}
	err := r.Get(ctx, req.NamespacedName, ofbiz)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Ofbiz resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Ofbiz resource")
		return ctrl.Result{}, err
	}

	// Reconcile Service
	service := r.serviceForOfbiz(ofbiz)
	err = r.reconcileService(ctx, ofbiz, service)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Reconcile StatefulSet
	statefulSet := r.statefulSetForOfbiz(ofbiz)
	err = r.reconcileStatefulSet(ctx, ofbiz, statefulSet)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update Status - a simple example of updating status
	// A more robust implementation would check pod readiness and other conditions.
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(ofbiz.Namespace),
		client.MatchingLabels{"app": ofbiz.Name},
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		logger.Error(err, "Failed to list pods", "Ofbiz.Namespace", ofbiz.Namespace, "Ofbiz.Name", ofbiz.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	ofbiz.Status.Nodes = podNames
	if err := r.Status().Update(ctx, ofbiz); err != nil {
		logger.Error(err, "Failed to update Ofbiz status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// serviceForOfbiz defines the Service for the Ofbiz cluster
func (r *OfbizReconciler) serviceForOfbiz(ofbiz *ofbizv1alpha1.Ofbiz) *corev1.Service {
	labels := map[string]string{"app": ofbiz.Name}
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ofbiz.Name,
			Namespace: ofbiz.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8080,
					TargetPort: intstr.FromInt(8080),
				},
				{
					Name:       "https",
					Port:       8443,
					TargetPort: intstr.FromInt(8443),
				},
			},
			Type: corev1.ServiceTypeLoadBalancer, // Or ClusterIP/NodePort
		},
	}
	ctrl.SetControllerReference(ofbiz, svc, r.Scheme)
	return svc
}

// reconcileService ensures the Service exists and is up-to-date
func (r *OfbizReconciler) reconcileService(ctx context.Context, ofbiz *ofbizv1alpha1.Ofbiz, service *corev1.Service) error {
	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.FromContext(ctx).Info("Creating a new Service", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		return r.Create(ctx, service)
	} else if err != nil {
		return err
	}
	// Update logic could be added here if the service spec needs to change
	return nil
}

// statefulSetForOfbiz defines the StatefulSet for the Ofbiz cluster
func (r *OfbizReconciler) statefulSetForOfbiz(ofbiz *ofbizv1alpha1.Ofbiz) *appsv1.StatefulSet {
	labels := map[string]string{"app": ofbiz.Name}
	replicas := ofbiz.Spec.Size

	// --- Define Volumes ---
	volumes := []corev1.Volume{}
	// Mount the configuration ConfigMap
	if ofbiz.Spec.Storage.ConfigurationConfigMapName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "config-volume",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ofbiz.Spec.Storage.ConfigurationConfigMapName,
					},
				},
			},
		})
	}
	// Mount the SSL Secret
	if ofbiz.Spec.Web.SslSecretName != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "ssl-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: ofbiz.Spec.Web.SslSecretName,
				},
			},
		})
	}

	// --- Define Volume Mounts ---
	volumeMounts := []corev1.VolumeMount{}
	if ofbiz.Spec.Storage.ConfigurationConfigMapName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name: "config-volume",
			// This path should correspond to where OFBiz looks for its config
			MountPath: "/ofbiz/framework/entity/config/entityengine.xml",
			SubPath:   "entityengine.xml", // Assumes this key exists in the ConfigMap
		})
	}
	if ofbiz.Spec.Web.SslSecretName != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "ssl-certs",
			MountPath: "/ofbiz/framework/base/config/ssl",
			ReadOnly:  true,
		})
	}
	// Mount for persistent data
	if ofbiz.Spec.Storage.Persistence.Enabled {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "ofbiz-data",
			MountPath: "/ofbiz/runtime/data",
		})
	}

	// --- Define Environment Variables ---
	envVars := []corev1.EnvVar{
		// Database Configuration
		{Name: "OFBIZ_DB_HOST", Value: ofbiz.Spec.Database.Host},
		{Name: "OFBIZ_DB_PORT", Value: fmt.Sprintf("%d", ofbiz.Spec.Database.Port)},
		{Name: "OFBIZ_DB_NAME", Value: ofbiz.Spec.Database.Name},
		{Name: "OFBIZ_DB_USER", Value: ofbiz.Spec.Database.User},
		{
			Name: "OFBIZ_DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ofbiz.Spec.Database.PasswordSecretName,
					},
					Key: "password",
				},
			},
		},
	}
	// Admin Password
	if ofbiz.Spec.InitialAdmin.PasswordSecretName != "" {
		envVars = append(envVars, corev1.EnvVar{
			Name: "OFBIZ_ADMIN_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ofbiz.Spec.InitialAdmin.PasswordSecretName,
					},
					Key: "password",
				},
			},
		})
	}

	// --- Define Volume Claim Templates for Persistent Storage ---
	pvcTemplates := []corev1.PersistentVolumeClaim{}
	if ofbiz.Spec.Storage.Persistence.Enabled {
		pvcTemplates = append(pvcTemplates, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ofbiz-data",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
				//				Resources: corev1.ResourceRequirements{
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: ofbiz.Spec.Storage.Persistence.Size,
					},
				},
				StorageClassName: &ofbiz.Spec.Storage.Persistence.StorageClassName,
			},
		})
	}

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ofbiz.Name,
			Namespace: ofbiz.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: ofbiz.Spec.Image,
						Name:  "ofbiz",
						Env:   envVars,
						Ports: []corev1.ContainerPort{
							{ContainerPort: 8080, Name: "http"},
							{ContainerPort: 8443, Name: "https"},
						},
						VolumeMounts: volumeMounts,
					}},
					Volumes: volumes,
				},
			},
			VolumeClaimTemplates: pvcTemplates,
		},
	}

	ctrl.SetControllerReference(ofbiz, sts, r.Scheme)
	return sts
}

// reconcileStatefulSet ensures the StatefulSet exists and is up-to-date
func (r *OfbizReconciler) reconcileStatefulSet(ctx context.Context, ofbiz *ofbizv1alpha1.Ofbiz, sts *appsv1.StatefulSet) error {
	found := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: sts.Name, Namespace: sts.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.FromContext(ctx).Info("Creating a new StatefulSet", "StatefulSet.Namespace", sts.Namespace, "StatefulSet.Name", sts.Name)
		return r.Create(ctx, sts)
	} else if err != nil {
		return err
	}

	// Update the found StatefulSet and apply the changes
	// This simple example just updates replicas. A real operator would compare more fields.
	if *found.Spec.Replicas != ofbiz.Spec.Size {
		found.Spec.Replicas = &ofbiz.Spec.Size
		log.FromContext(ctx).Info("Updating StatefulSet replicas", "Name", found.Name, "Replicas", ofbiz.Spec.Size)
		return r.Update(ctx, found)
	}

	return nil
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *OfbizReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ofbizv1alpha1.Ofbiz{}).
		Owns(&appsv1.StatefulSet{}). // The operator owns the StatefulSets it creates
		Owns(&corev1.Service{}).     // and the Services
		Complete(r)
}

/* originally scaffolded
package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	ofbizv1alpha1 "github.com/halims/ofbiz-operator/api/v1alpha1"
)

// OfbizReconciler reconciles a Ofbiz object
type OfbizReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ofbiz.ofbiz.apache.org,resources=ofbizzes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ofbiz.ofbiz.apache.org,resources=ofbizzes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ofbiz.ofbiz.apache.org,resources=ofbizzes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ofbiz object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *OfbizReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OfbizReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ofbizv1alpha1.Ofbiz{}).
		Named("ofbiz").
		Complete(r)
}
*/
