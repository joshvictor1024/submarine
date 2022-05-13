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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	submarinev1alpha1 "github.com/apache/submarine/submarine-operator/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *SubmarineReconciler) newSubmarineServerService(ctx context.Context, submarine *submarinev1alpha1.Submarine) *corev1.Service {
	l := log.FromContext(ctx)
	service, err := ParseServiceYaml(serverYamlPath)
	if err != nil {
		l.Error(err, "ParseServiceYaml")
	}
	service.Namespace = submarine.Namespace
	err = controllerutil.SetControllerReference(submarine, service, r.Scheme)
	if err != nil {
		l.Error(err, "Set Service ControllerReference")
	}
	return service
}

func (r *SubmarineReconciler) newSubmarineServerDeployment(ctx context.Context, submarine *submarinev1alpha1.Submarine) *appsv1.Deployment {
	l := log.FromContext(ctx)
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
		l.Error(err, "ParseDeploymentYaml")
	}
	deployment.Namespace = submarine.Namespace
	err = controllerutil.SetControllerReference(submarine, deployment, r.Scheme)
	if err != nil {
		l.Error(err, "Set Deployment ControllerReference")
	}
	deployment.Spec.Replicas = &serverReplicas
	deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, operatorEnv...)

	return deployment
}

// createSubmarineServer is a function to create submarine-server.
// Reference: https://github.com/apache/submarine/blob/master/helm-charts/submarine/templates/submarine-server.yaml
func (r *SubmarineReconciler) createSubmarineServer(ctx context.Context, submarine *submarinev1alpha1.Submarine) error {
	l := log.FromContext(ctx)
	l.Info("[createSubmarineServer]")

	// Step1: Create Service
	service := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: serverName, Namespace: submarine.Namespace}, service)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		service = r.newSubmarineServerService(ctx, submarine)
		err = r.Create(ctx, service)
		l.Info("Create Service", "service name", service.Name)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if !metav1.IsControlledBy(service, submarine) {
		msg := fmt.Sprintf(MessageResourceExists, service.Name)
		//c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
		l.Error(fmt.Errorf(msg), "Check Service ControllerReference", "service name", service.Name)
		return fmt.Errorf(msg)
	}

	// Step2: Create Deployment
	deployment := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: serverName, Namespace: submarine.Namespace}, deployment)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment = r.newSubmarineServerDeployment(ctx, submarine)
		err = r.Create(ctx, deployment)
		l.Info("Create Deployment", "Deployment name", deployment.Name)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if !metav1.IsControlledBy(deployment, submarine) {
		msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
		//c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}

	// Update the replicas of the server deployment if it is not equal to spec
	if submarine.Spec.Server.Replicas != nil && *submarine.Spec.Server.Replicas != *deployment.Spec.Replicas {
		msg := fmt.Sprintf("Submarine %s server spec replicas", submarine.Name)
		l.Info(msg, "server spec", *submarine.Spec.Server.Replicas, "actual", *deployment.Spec.Replicas)

		deployment = r.newSubmarineServerDeployment(ctx, submarine)
		err = r.Update(ctx, deployment)
	}

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
