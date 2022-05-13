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

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	submarinev1alpha1 "github.com/apache/submarine/submarine-operator/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *SubmarineReconciler) newSubmarineObserverRole(ctx context.Context, submarine *submarinev1alpha1.Submarine) *rbacv1.Role {
	l := log.FromContext(ctx)
	role, err := ParseRoleYaml(observerRbacYamlPath)
	if err != nil {
		l.Error(err, "ParseRoleYaml")
	}
	role.Namespace = submarine.Namespace
	err = controllerutil.SetControllerReference(submarine, role, r.Scheme)
	if err != nil {
		l.Error(err, "Set Role ControllerReference")
	}
	return role
}

func (r *SubmarineReconciler) newSubmarineObserverRoleBinding(ctx context.Context, submarine *submarinev1alpha1.Submarine) *rbacv1.RoleBinding {
	l := log.FromContext(ctx)
	roleBinding, err := ParseRoleBindingYaml(observerRbacYamlPath)
	if err != nil {
		l.Error(err, "Set RoleBinding ControllerReference")
	}
	roleBinding.Namespace = submarine.Namespace
	err = controllerutil.SetControllerReference(submarine, roleBinding, r.Scheme)
	if err != nil {
		l.Error(err, "Set RoleBinding ControllerReference")
	}
	return roleBinding
}

// createSubmarineObserverRBAC is a function to create RBAC for submarine-observer which will be binded on service account: default.
// Reference: https://github.com/apache/submarine/blob/master/helm-charts/submarine/templates/rbac.yaml
func (r *SubmarineReconciler) createSubmarineObserverRBAC(ctx context.Context, submarine *submarinev1alpha1.Submarine) error {
	l := log.FromContext(ctx)
	l.Info("[createSubmarineObserverRBAC]")

	// Step1: Create Role
	role := &rbacv1.Role{}
	err := r.Get(ctx, types.NamespacedName{Name: observerName, Namespace: submarine.Namespace}, role)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		role = r.newSubmarineObserverRole(ctx, submarine)
		err = r.Create(ctx, role)
		l.Info("Create Role", "role name", role.Name)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(role, submarine) {
		msg := fmt.Sprintf(MessageResourceExists, role.Name)
		//c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
		l.Error(fmt.Errorf(msg), "Check Role ControllerReference", "role name", role.Name)
		return fmt.Errorf(msg)
	}

	// Step2: Create Role Binding
	rolebinding := &rbacv1.RoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: observerName, Namespace: submarine.Namespace}, rolebinding)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		rolebinding = r.newSubmarineObserverRoleBinding(ctx, submarine)
		err = r.Create(ctx, rolebinding)
		l.Info("Create RoleBinding", "rolebinding name", rolebinding.Name)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(rolebinding, submarine) {
		msg := fmt.Sprintf(MessageResourceExists, rolebinding.Name)
		//c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
		l.Error(fmt.Errorf(msg), "Check RoleBinding ControllerReference", "rolebinding name", rolebinding.Name)
		return fmt.Errorf(msg)
	}

	return nil
}
