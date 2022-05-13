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

	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	submarinev1alpha1 "github.com/apache/submarine/submarine-operator/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *SubmarineReconciler) newSubmarineVirtualService(ctx context.Context, submarine *submarinev1alpha1.Submarine) *istiov1alpha3.VirtualService {
	l := log.FromContext(ctx)
	virtualService, err := ParseVirtualService(ingressYamlPath)
	if err != nil {
		l.Error(err, "ParseVirtualService")
	}
	virtualService.Namespace = submarine.Namespace
	err = controllerutil.SetControllerReference(submarine, virtualService, r.Scheme)
	if err != nil {
		l.Error(err, "Set VirtualService ControllerReference")
	}
	return virtualService
}

// createIngress is a function to create Ingress.
// Reference: https://github.com/apache/submarine/blob/master/helm-charts/submarine/templates/submarine-ingress.yaml
func (r *SubmarineReconciler) createIngress(ctx context.Context, submarine *submarinev1alpha1.Submarine) error {
	l := log.FromContext(ctx)
	l.Info("[createIngress]")

	virtualService := &istiov1alpha3.VirtualService{}
	err := r.Get(ctx, types.NamespacedName{Name: virtualServiceName, Namespace: submarine.Namespace}, virtualService)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		virtualService = r.newSubmarineVirtualService(ctx, submarine)
		err = r.Create(ctx, virtualService)
		l.Info("Create VirtualService", "virtualService name", virtualService.Name)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(virtualService, submarine) {
		msg := fmt.Sprintf(MessageResourceExists, virtualService.Name)
		//c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
		l.Error(fmt.Errorf(msg), "Check VirtualService ControllerReference", "virtualService name", virtualService.Name)
		return fmt.Errorf(msg)
	}

	return nil
}
