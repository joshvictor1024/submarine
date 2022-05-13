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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	submarinev1alpha1 "github.com/apache/submarine/submarine-operator/api/v1alpha1"
)

const (
	serverName = "submarine-server"
	// databaseName                = "submarine-database"
	// databasePort                = 3306
	// tensorboardName             = "submarine-tensorboard"
	// mlflowName                  = "submarine-mlflow"
	// minioName                   = "submarine-minio"
	// ingressName                 = serverName + "-ingress"
	// databasePvcName             = databaseName + "-pvc"
	// tensorboardPvcName          = tensorboardName + "-pvc"
	// tensorboardServiceName      = tensorboardName + "-service"
	// tensorboardIngressRouteName = tensorboardName + "-ingressroute"
	// mlflowPvcName               = mlflowName + "-pvc"
	// mlflowServiceName           = mlflowName + "-service"
	// mlflowIngressRouteName      = mlflowName + "-ingressroute"
	// minioPvcName                = minioName + "-pvc"
	// minioServiceName            = minioName + "-service"
	// minioIngressRouteName       = minioName + "-ingressroute"
	artifactPath = "./artifacts/"
	// databaseYamlPath            = artifactPath + "submarine-database.yaml"
	// ingressYamlPath             = artifactPath + "submarine-ingress.yaml"
	// minioYamlPath               = artifactPath + "submarine-minio.yaml"
	// mlflowYamlPath              = artifactPath + "submarine-mlflow.yaml"
	serverYamlPath = artifactPath + "submarine-server.yaml"
	// tensorboardYamlPath         = artifactPath + "submarine-tensorboard.yaml"
	// rbacYamlPath                = artifactPath + "submarine-rbac.yaml"
)

// SubmarineReconciler reconciles a Submarine object
type SubmarineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=submarine.apache.org,resources=submarines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=submarine.apache.org,resources=submarines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=submarine.apache.org,resources=submarines/finalizers,verbs=update

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
		// ok, err := c.checkSubmarineDependentsReady(submarineCopy)
		// if err != nil {
		// 	submarine.Status.SubmarineState.State = submarinev1alpha1.FailedState
		// 	submarine.Status.SubmarineState.ErrorMessage = err.Error()
		// 	// c.recordSubmarineEvent(submarineCopy)
		// }
		// if ok {
		// 	submarine.Status.SubmarineState.State = submarinev1alpha1.RunningState
		// 	// c.recordSubmarineEvent(submarineCopy)
		// }
		// case submarinev1alpha1.RunningState:
		// 	if err := c.createSubmarine(submarineCopy); err != nil {
		// 		submarine.Status.SubmarineState.State = submarinev1alpha1.FailedState
		// 		submarine.Status.SubmarineState.ErrorMessage = err.Error()
		// 		c.recordSubmarineEvent(submarineCopy)
		// 	}
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
	err = r.createSubmarineServer(ctx, submarine)
	if err != nil {
		return err
	}

	// err = r.createSubmarineDatabase(submarine)
	// if err != nil {
	// 	return err
	// }

	// err = r.createIngress(submarine)
	// if err != nil {
	// 	return err
	// }

	// err = r.createSubmarineServerRBAC(submarine)
	// if err != nil {
	// 	return err
	// }

	// err = r.createSubmarineTensorboard(submarine)
	// if err != nil {
	// 	return err
	// }

	// err = r.createSubmarineMlflow(submarine)
	// if err != nil {
	// 	return err
	// }

	// err = r.createSubmarineMinio(submarine)
	// if err != nil {
	// 	return err
	// }

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SubmarineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&submarinev1alpha1.Submarine{}).
		Complete(r)
}
