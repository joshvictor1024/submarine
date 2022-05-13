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
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	submarinev1alpha1 "github.com/apache/submarine/submarine-operator/api/v1alpha1"
)

const (
	serverName   = "submarine-server"
	observerName = "submarine-observer"
	databaseName = "submarine-database"
	// databasePort    = 3306
	tensorboardName = "submarine-tensorboard"
	mlflowName      = "submarine-mlflow"
	minioName       = "submarine-minio"
	storageName     = "submarine-storage"
	// ingressName = serverName + "-ingress"
	virtualServiceName     = "submarine-virtual-service"
	databasePvcName        = databaseName + "-pvc"
	tensorboardPvcName     = tensorboardName + "-pvc"
	tensorboardServiceName = tensorboardName + "-service"
	// tensorboardIngressRouteName = tensorboardName + "-ingressroute"
	mlflowPvcName     = mlflowName + "-pvc"
	mlflowServiceName = mlflowName + "-service"
	// mlflowIngressRouteName      = mlflowName + "-ingressroute"
	minioPvcName     = minioName + "-pvc"
	minioServiceName = minioName + "-service"
	// minioIngressRouteName       = minioName + "-ingressroute"
	artifactPath         = "./artifacts/"
	databaseYamlPath     = artifactPath + "submarine-database.yaml"
	ingressYamlPath      = artifactPath + "submarine-ingress.yaml"
	minioYamlPath        = artifactPath + "submarine-minio.yaml"
	mlflowYamlPath       = artifactPath + "submarine-mlflow.yaml"
	serverYamlPath       = artifactPath + "submarine-server.yaml"
	tensorboardYamlPath  = artifactPath + "submarine-tensorboard.yaml"
	rbacYamlPath         = artifactPath + "submarine-rbac.yaml"
	observerRbacYamlPath = artifactPath + "submarine-observer-rbac.yaml"
	storageRbacYamlPath  = artifactPath + "submarine-storage-rbac.yaml"
)

var dependents = []string{serverName, tensorboardName, mlflowName, minioName}

const (
	// SuccessSynced is used as part of the Event 'reason' when a Submarine is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Submarine fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Submarine"
	// MessageResourceSynced is the message used for an Event fired when a
	// Submarine is synced successfully
	MessageResourceSynced = "Submarine synced successfully"
)

// Default k8s anyuid role rule
var k8sAnyuidRoleRule = rbacv1.PolicyRule{
	APIGroups:     []string{"policy"},
	Verbs:         []string{"use"},
	Resources:     []string{"podsecuritypolicies"},
	ResourceNames: []string{"submarine-anyuid"},
}

// Openshift anyuid role rule
var openshiftAnyuidRoleRule = rbacv1.PolicyRule{
	APIGroups:     []string{"security.openshift.io"},
	Verbs:         []string{"use"},
	Resources:     []string{"securitycontextconstraints"},
	ResourceNames: []string{"anyuid"},
}

// SubmarineReconciler reconciles a Submarine object
type SubmarineReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Incluster               bool
	ClusterType             string
	CreatePodSecurityPolicy bool
}

//+kubebuilder:rbac:groups=submarine.apache.org,resources=submarines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=submarine.apache.org,resources=submarines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=submarine.apache.org,resources=submarines/finalizers,verbs=update

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac,resources=roles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rbac,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

//+kubebuilder:rbac:groups=networking.istio.io,resources=virtualservices,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Submarine object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *SubmarineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	l.Info("Enter Reconcile", "req", req)

	// TODO(user): your logic here

	submarine := &submarinev1alpha1.Submarine{}
	err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, submarine)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	l.Info("Enter Reconcile", "spec", submarine.Spec, "status", submarine.Status)

	// Take action based on submarine state
	switch submarine.Status.State {
	case submarinev1alpha1.NewState:
		//c.recordSubmarineEvent(submarineCopy)
		if err := r.validateSubmarine(submarine); err != nil {
			submarine.Status.State = submarinev1alpha1.FailedState
			submarine.Status.ErrorMessage = err.Error()
			// c.recordSubmarineEvent(submarineCopy)
		} else {
			submarine.Status.State = submarinev1alpha1.CreatingState
			// c.recordSubmarineEvent(submarineCopy)
		}
	case submarinev1alpha1.CreatingState:
		//if err := c.createSubmarine(submarineCopy); err != nil {
		if err := r.createSubmarine(ctx, submarine); err != nil {
			submarine.Status.State = submarinev1alpha1.FailedState
			submarine.Status.ErrorMessage = err.Error()
			// c.recordSubmarineEvent(submarineCopy)
		}
		ok, err := r.checkSubmarineDependentsReady(ctx, submarine)
		if err != nil {
			submarine.Status.State = submarinev1alpha1.FailedState
			submarine.Status.ErrorMessage = err.Error()
			// c.recordSubmarineEvent(submarineCopy)
		}
		if ok {
			submarine.Status.State = submarinev1alpha1.RunningState
			// c.recordSubmarineEvent(submarineCopy)
		}
	case submarinev1alpha1.RunningState:
		if err := r.createSubmarine(ctx, submarine); err != nil {
			submarine.Status.State = submarinev1alpha1.FailedState
			submarine.Status.ErrorMessage = err.Error()
			// c.recordSubmarineEvent(submarineCopy)
		}
	}

	r.Status().Update(ctx, submarine)
	return ctrl.Result{}, nil
}

func (r *SubmarineReconciler) validateSubmarine(submarine *submarinev1alpha1.Submarine) error {

	// Print out the spec of the Submarine resource
	b, err := json.MarshalIndent(submarine.Spec, "", "  ")
	fmt.Println(string(b))

	if err != nil {
		return err
	}

	return nil
}

func (r *SubmarineReconciler) createSubmarine(ctx context.Context, submarine *submarinev1alpha1.Submarine) error {
	var err error
	// We create rbac first, this ensures that any dependency based on it will not go wrong
	err = r.createSubmarineServerRBAC(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createSubmarineStorageRBAC(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createSubmarineObserverRBAC(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createSubmarineServer(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createSubmarineDatabase(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createIngress(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createSubmarineTensorboard(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createSubmarineMlflow(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = r.createSubmarineMinio(ctx, submarine)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func (r *SubmarineReconciler) getDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment, nil
}

func (r *SubmarineReconciler) getStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	statefulset := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, statefulset)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return statefulset, nil
}

func (r *SubmarineReconciler) checkSubmarineDependentsReady(ctx context.Context, submarine *submarinev1alpha1.Submarine) (bool, error) {
	// deployment dependents check
	for _, name := range dependents {
		deployment, err := r.getDeployment(ctx, submarine.Namespace, name)
		if err != nil {
			return false, err
		} else if deployment == nil {
			return false, nil
		}
		// check if deployment replicas failed
		for _, condition := range deployment.Status.Conditions {
			if condition.Type == appsv1.DeploymentReplicaFailure {
				return false, fmt.Errorf("failed creating replicas of %s, message: %s", deployment.Name, condition.Message)
			}
		}
		// check if ready replicas are same as targeted replicas
		if deployment.Status.ReadyReplicas != deployment.Status.Replicas {
			return false, nil
		}
	}
	// database check
	// statefulset.Status.Conditions does not have the specified type enum like
	// deployment.Status.Conditions => DeploymentConditionType ,
	// so we will ignore the verification status for the time being
	statefulset, err := r.getStatefulSet(ctx, submarine.Namespace, databaseName)
	if err != nil {
		return false, err
	} else if statefulset == nil || statefulset.Status.Replicas != statefulset.Status.ReadyReplicas {
		return false, nil
	}

	return true, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubmarineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&submarinev1alpha1.Submarine{}).
		Complete(r)
}
