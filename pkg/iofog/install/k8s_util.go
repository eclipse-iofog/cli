/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package install

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	extsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

func newService(namespace string, ms *microservice) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
			Labels: map[string]string{
				"name": ms.name,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:                  corev1.ServiceTypeLoadBalancer,
			ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
			LoadBalancerIP:        ms.IP,
			Selector: map[string]string{
				"name": ms.name,
			},
		},
	}
	// Add ports
	for idx, port := range ms.ports {
		svcPort := corev1.ServicePort{
			Name:       ms.name + strconv.Itoa(idx),
			Port:       int32(port),
			TargetPort: intstr.FromInt(int(port)),
			Protocol:   corev1.Protocol("TCP"),
		}
		svc.Spec.Ports = append(svc.Spec.Ports, svcPort)
	}
	return svc
}

func newDeployment(namespace string, ms *microservice) *appsv1.Deployment {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
			Labels: map[string]string{
				"name": ms.name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &ms.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"name": ms.name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"name": ms.name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: ms.name,
				},
			},
		},
	}
	containers := &dep.Spec.Template.Spec.Containers
	for _, msCont := range ms.containers {
		cont := corev1.Container{
			Name:            msCont.name,
			Image:           msCont.image,
			ImagePullPolicy: corev1.PullPolicy(msCont.imagePullPolicy),
			Args:            msCont.args,
			ReadinessProbe:  msCont.readinessProbe,
			Ports:           msCont.ports,
			Env:             msCont.env,
			Command:         msCont.command,
		}
		*containers = append(*containers, cont)
	}
	return dep
}

func newServiceAccount(namespace string, ms *microservice) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
		},
	}
}

func getClusterRoleBindingName(namespace, resourceName string) string {
	return namespace + "-" + resourceName
}

func newClusterRoleBinding(namespace string, ms *microservice) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: getClusterRoleBindingName(namespace, ms.name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      ms.name,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func newRoleBinding(namespace string, ms *microservice) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: ms.name,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     ms.name,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
}

func newRole(namespace string, ms *microservice) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ms.name,
			Namespace: namespace,
		},
		Rules: ms.rbacRules,
	}
}

func newKogCRD() *extsv1.CustomResourceDefinition {
	return &extsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kogs.iofog.org",
		},
		Spec: extsv1.CustomResourceDefinitionSpec{
			Group: "iofog.org",
			Names: extsv1.CustomResourceDefinitionNames{
				Kind:     "Kog",
				ListKind: "KogList",
				Plural:   "kogs",
				Singular: "kog",
			},
			Scope:   extsv1.ResourceScope("Namespaced"),
			Version: "v1",
			Subresources: &extsv1.CustomResourceSubresources{
				Status: &extsv1.CustomResourceSubresourceStatus{},
			},
		},
	}
}

func newAppCRD() *extsv1.CustomResourceDefinition {
	return &extsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "apps.iofog.org",
		},
		Spec: extsv1.CustomResourceDefinitionSpec{
			Group: "iofog.org",
			Names: extsv1.CustomResourceDefinitionNames{
				Kind:     "Application",
				ListKind: "ApplicationList",
				Plural:   "apps",
				Singular: "app",
			},
			Scope:   extsv1.ResourceScope("Namespaced"),
			Version: "v1",
			Subresources: &extsv1.CustomResourceSubresources{
				Status: &extsv1.CustomResourceSubresourceStatus{},
			},
		},
	}
}
