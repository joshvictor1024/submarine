/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	"context"
	// "fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	//"k8s.io/klog/v2"
	submarinev1 "github.com/apache/submarine/submarine-operator/api/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func newSubmarineServerServiceAccount() (*corev1.ServiceAccount, error) {
	serviceAccount, err := ParseServiceAccountYaml(serverYamlPath)
	if err != nil {
		//klog.Info("[Error] ParseServiceAccountYaml", err)
		return nil, err
	}
	// serviceAccount.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
	// 	*metav1.NewControllerRef(submarine, v1alpha1.SchemeGroupVersion.WithKind("Submarine")),
	// }
	return serviceAccount, nil
}

func newSubmarineServerService() (*corev1.Service, error) {
	service, err := ParseServiceYaml(serverYamlPath)
	if err != nil {
		//klog.Info("[Error] ParseServiceYaml", err)
		return nil, err
	}
	// service.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
	// 	*metav1.NewControllerRef(submarine, v1alpha1.SchemeGroupVersion.WithKind("Submarine")),
	// }
	return service, nil
}

func newSubmarineServerDeployment(submarine *submarinev1.Submarine) (*appsv1.Deployment, error) {
	serverReplicas := *submarine.Spec.Server.Replicas

	//ownerReference := *metav1.NewControllerRef(submarine, v1alpha1.SchemeGroupVersion.WithKind("Submarine"))
	operatorEnv := []corev1.EnvVar{
		{
			Name:  "SUBMARINE_SERVER_DNS_NAME",
			Value: serverName + "." + submarine.Namespace,
		},
		{
			Name:  "ENV_NAMESPACE",
			Value: submarine.Namespace,
		},
		{
			Name:  "SUBMARINE_APIVERSION",
			Value: submarine.APIVersion,
		},
		{
			Name:  "SUBMARINE_KIND",
			Value: submarine.Kind,
		},
		{
			Name:  "SUBMARINE_NAME",
			Value: submarine.Name,
		},
		{
			Name:  "SUBMARINE_UID",
			Value: string(submarine.UID),
		},
	}

	deployment, err := ParseDeploymentYaml(serverYamlPath)
	if err != nil {
		// klog.Info("[Error] ParseDeploymentYaml", err)
		return nil, err
	}
	// deployment.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
	// 	ownerReference,
	// }
	deployment.Spec.Replicas = &serverReplicas
	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, operatorEnv...)

	return deployment, nil
}

// createSubmarineServer is a function to create submarine-server.
// Reference: https://github.com/apache/submarine/blob/master/helm-charts/submarine/templates/submarine-server.yaml
func (r *SubmarineReconciler) createSubmarineServer(ctx context.Context, submarine *submarinev1.Submarine) error {
	l := log.FromContext(ctx)
	l.Info("[createSubmarineServer]")
	//klog.Info("[createSubmarineServer]")

	// Step1: Create ServiceAccount
	serviceaccount := &corev1.ServiceAccount{}
	err := r.Get(ctx, types.NamespacedName{Name: serverName, Namespace: submarine.Namespace}, submarine)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		if serviceaccount, err = newSubmarineServerServiceAccount(); err != nil {
			return err
		}
		if err = controllerutil.SetControllerReference(submarine, serviceaccount, r.Scheme); err != nil {
			//return reconcile.Result{}, err
			return err
		}
		if err = r.Create(ctx, serviceaccount); err != nil {
			// return ctrl.Result{}, err
			return err
		}
		//klog.Info("	Create ServiceAccount: ", serviceaccount.Name)
	}
	if err != nil {
		if errors.IsNotFound(err) {
			//return ctrl.Result{}, nil
			return nil
		}
		// return ctrl.Result{}, err
		return err
	}

	// serviceaccount, err := c.serviceaccountLister.ServiceAccounts(submarine.Namespace).Get(serverName)
	// // If the resource doesn't exist, we'll create it
	// if errors.IsNotFound(err) {
	// 	serviceaccount, err = c.kubeclientset.CoreV1().ServiceAccounts(submarine.Namespace).Create(context.TODO(), newSubmarineServerServiceAccount(submarine), metav1.CreateOptions{})
	// 	klog.Info("	Create ServiceAccount: ", serviceaccount.Name)
	// }

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	// TODO
	// if !metav1.IsControlledBy(serviceaccount, submarine) {
	// 	msg := fmt.Sprintf(MessageResourceExists, serviceaccount.Name)
	// 	c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
	// 	return fmt.Errorf(msg)
	// }

	// // Step2: Create Service
	// service, err := c.serviceLister.Services(submarine.Namespace).Get(serverName)
	// // If the resource doesn't exist, we'll create it
	// if errors.IsNotFound(err) {
	// 	service, err = c.kubeclientset.CoreV1().Services(submarine.Namespace).Create(context.TODO(), newSubmarineServerService(submarine), metav1.CreateOptions{})
	// 	klog.Info("	Create Service: ", service.Name)
	// }

	// // If an error occurs during Get/Create, we'll requeue the item so we can
	// // attempt processing again later. This could have been caused by a
	// // temporary network failure, or any other transient reason.
	// if err != nil {
	// 	return err
	// }

	// if !metav1.IsControlledBy(service, submarine) {
	// 	msg := fmt.Sprintf(MessageResourceExists, service.Name)
	// 	c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
	// 	return fmt.Errorf(msg)
	// }

	// // Step3: Create Deployment
	// deployment, err := c.deploymentLister.Deployments(submarine.Namespace).Get(serverName)
	// // If the resource doesn't exist, we'll create it
	// if errors.IsNotFound(err) {
	// 	deployment, err = c.kubeclientset.AppsV1().Deployments(submarine.Namespace).Create(context.TODO(), newSubmarineServerDeployment(submarine), metav1.CreateOptions{})
	// 	klog.Info("	Create Deployment: ", deployment.Name)
	// }

	// // If an error occurs during Get/Create, we'll requeue the item so we can
	// // attempt processing again later. This could have been caused by a
	// // temporary network failure, or any other transient reason.
	// if err != nil {
	// 	return err
	// }

	// if !metav1.IsControlledBy(deployment, submarine) {
	// 	msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
	// 	c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
	// 	return fmt.Errorf(msg)
	// }

	// // Update the replicas of the server deployment if it is not equal to spec
	// if submarine.Spec.Server.Replicas != nil && *submarine.Spec.Server.Replicas != *deployment.Spec.Replicas {
	// 	klog.V(4).Infof("Submarine %s server spec replicas: %d, actual replicas: %d", submarine.Name, *submarine.Spec.Server.Replicas, *deployment.Spec.Replicas)
	// 	_, err = c.kubeclientset.AppsV1().Deployments(submarine.Namespace).Update(context.TODO(), newSubmarineServerDeployment(submarine), metav1.UpdateOptions{})
	// }

	if err != nil {
		return err
	}

	return nil
}

// func (r *Reconciler) createNewExtendedDaemonSet(logger logr.Logger, dda *datadoghqv1alpha1.DatadogAgent, newStatus *datadoghqv1alpha1.DatadogAgentStatus) (reconcile.Result, error) {
// 	var err error
// 	// ExtendedDaemonSet up to date didn't exist yet, create a new one
// 	var newEDS *edsdatadoghqv1alpha1.ExtendedDaemonSet
// 	var hashEDS string
// 	if newEDS, hashEDS, err = newExtendedDaemonSetFromInstance(logger, dda, nil); err != nil {
// 		return reconcile.Result{}, err
// 	}

// 	// Set ExtendedDaemonSet instance as the owner and controller
// 	if err = controllerutil.SetControllerReference(dda, newEDS, r.scheme); err != nil {
// 		return reconcile.Result{}, err
// 	}

// 	logger.Info("Creating a new ExtendedDaemonSet", "extendedDaemonSet.Namespace", newEDS.Namespace, "extendedDaemonSet.Name", newEDS.Name, "agentdeployment.Status.Agent.CurrentHash", hashEDS)

// 	err = r.client.Create(context.TODO(), newEDS)
// 	if err != nil {
// 		return reconcile.Result{}, err
// 	}
// 	event := buildEventInfo(newEDS.Name, newEDS.Namespace, extendedDaemonSetKind, datadog.CreationEvent)
// 	r.recordEvent(dda, event)
// 	now := metav1.NewTime(time.Now())
// 	newStatus.Agent = updateExtendedDaemonSetStatus(newEDS, newStatus.Agent, &now)

// 	return reconcile.Result{}, nil
// }
