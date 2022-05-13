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

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	submarinev1alpha1 "github.com/apache/submarine/submarine-operator/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *SubmarineReconciler) newSubmarineServerServiceAccount(ctx context.Context, submarine *submarinev1alpha1.Submarine) *corev1.ServiceAccount {
	l := log.FromContext(ctx)
	serviceAccount, err := ParseServiceAccountYaml(serverYamlPath)
	if err != nil {
		l.Error(err, "ParseServiceAccountYaml")
	}
	serviceAccount.Namespace = submarine.Namespace
	err = controllerutil.SetControllerReference(submarine, serviceAccount, r.Scheme)
	if err != nil {
		l.Error(err, "Set ServiceAccount ControllerReference")
	}
	return serviceAccount
}

func (r *SubmarineReconciler) newSubmarineServerRole(ctx context.Context, submarine *submarinev1alpha1.Submarine) *rbacv1.Role {
	l := log.FromContext(ctx)
	role, err := ParseRoleYaml(rbacYamlPath)
	if err != nil {
		l.Error(err, "ParseRoleYaml")
	}
	role.Namespace = submarine.Namespace
	err = controllerutil.SetControllerReference(submarine, role, r.Scheme)
	if err != nil {
		l.Error(err, "Set Role ControllerReference")
	}

	if r.CreatePodSecurityPolicy {
		// If cluster type is openshift and need create pod security policy, we need add anyuid scc, or we add k8s psp
		if r.ClusterType == "openshift" {
			role.Rules = append(role.Rules, openshiftAnyuidRoleRule)
		} else {
			role.Rules = append(role.Rules, k8sAnyuidRoleRule)
		}
	}

	return role
}

func (r *SubmarineReconciler) newSubmarineServerRoleBinding(ctx context.Context, submarine *submarinev1alpha1.Submarine) *rbacv1.RoleBinding {
	l := log.FromContext(ctx)
	roleBinding, err := ParseRoleBindingYaml(rbacYamlPath)
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

// createSubmarineServerRBAC is a function to create RBAC for submarine-server.
// Reference: https://github.com/apache/submarine/blob/master/helm-charts/submarine/templates/rbac.yaml
func (r *SubmarineReconciler) createSubmarineServerRBAC(ctx context.Context, submarine *submarinev1alpha1.Submarine) error {
	l := log.FromContext(ctx)
	l.Info("[createSubmarineServerRBAC]")

	// Step1: Create ServiceAccount
	serviceaccount := &corev1.ServiceAccount{}
	err := r.Get(ctx, types.NamespacedName{Name: serverName, Namespace: submarine.Namespace}, serviceaccount)

	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		serviceaccount = r.newSubmarineServerServiceAccount(ctx, submarine)
		err = r.Create(ctx, serviceaccount)
		l.Info("Create ServiceAccount", "serviceaccount name", serviceaccount.Name)
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return err
	}

	if !metav1.IsControlledBy(serviceaccount, submarine) {
		msg := fmt.Sprintf(MessageResourceExists, serviceaccount.Name)
		//c.recorder.Event(submarine, corev1.EventTypeWarning, ErrResourceExists, msg)
		l.Error(fmt.Errorf(msg), "Check ServiceAccount ControllerReference", "service name", serviceaccount.Name)
		return fmt.Errorf(msg)
	}

	// Step2: Create Role
	role := &rbacv1.Role{}
	err = r.Get(ctx, types.NamespacedName{Name: serverName, Namespace: submarine.Namespace}, role)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		role = r.newSubmarineServerRole(ctx, submarine)
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

	// Step3: Create Role Binding
	rolebinding := &rbacv1.RoleBinding{}
	err = r.Get(ctx, types.NamespacedName{Name: serverName, Namespace: submarine.Namespace}, rolebinding)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		rolebinding = r.newSubmarineServerRoleBinding(ctx, submarine)
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
		l.Error(fmt.Errorf(msg), "Check RoleBinding ControllerReference", "role name", rolebinding.Name)
		return fmt.Errorf(msg)
	}

	return nil
}
