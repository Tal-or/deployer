/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2021 Red Hat, Inc.
 */

package manifests

import (
	"embed"
	"fmt"
	"io"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"k8s.io/client-go/kubernetes/scheme"
)

const (
	ComponentAPI                      = "api"
	ComponentSchedulerPlugin          = "sched"
	ComponentResourceTopologyExporter = "rte"
)

//go:embed yaml
var src embed.FS

func init() {
	apiextensionv1.AddToScheme(scheme.Scheme)
}

func Namespace(component string) (*corev1.Namespace, error) {
	if err := validateComponent(component); err != nil {
		return nil, err
	}

	obj, err := loadObject(filepath.Join("yaml", component, "namespace.yaml"))
	if err != nil {
		return nil, err
	}

	ns, ok := obj.(*corev1.Namespace)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return ns, nil
}

func ServiceAccount(component string) (*corev1.ServiceAccount, error) {
	if err := validateComponent(component); err != nil {
		return nil, err
	}

	obj, err := loadObject(filepath.Join("yaml", component, "serviceaccount.yaml"))
	if err != nil {
		return nil, err
	}

	sa, ok := obj.(*corev1.ServiceAccount)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return sa, nil
}

func ClusterRole(component string) (*rbacv1.ClusterRole, error) {
	if err := validateComponent(component); err != nil {
		return nil, err
	}

	obj, err := loadObject(filepath.Join("yaml", component, "clusterrole.yaml"))
	if err != nil {
		return nil, err
	}

	cr, ok := obj.(*rbacv1.ClusterRole)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return cr, nil
}

func APICRD() (*apiextensionv1.CustomResourceDefinition, error) {
	obj, err := loadObject("yaml/api/crd.yaml")
	if err != nil {
		return nil, err
	}

	crd, ok := obj.(*apiextensionv1.CustomResourceDefinition)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return crd, nil
}

func SchedulerPluginConfigMap() (*corev1.ConfigMap, error) {
	obj, err := loadObject("yaml/sched/configmap.yaml")
	if err != nil {
		return nil, err
	}

	crd, ok := obj.(*corev1.ConfigMap)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return crd, nil
}

func SchedulerPluginDeployment() (*appsv1.Deployment, error) {
	obj, err := loadObject("yaml/sched/deployment.yaml")
	if err != nil {
		return nil, err
	}

	dp, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return dp, nil
}

func loadClusterRoleBinding(component, detail string) (*rbacv1.ClusterRoleBinding, error) {
	if err := validateComponent(component); err != nil {
		return nil, err
	}

	crbName := "clusterrolebinding.yaml"
	if detail != "" {
		crbName = fmt.Sprintf("clusterrolebinding-%s.yaml", detail)
	}
	obj, err := loadObject(filepath.Join("yaml", component, crbName))
	if err != nil {
		return nil, err
	}

	crb, ok := obj.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return crb, nil
}

func SchedulerPluginClusterRoleBindingKubeScheduler() (*rbacv1.ClusterRoleBinding, error) {
	return loadClusterRoleBinding(ComponentSchedulerPlugin, "kube-sched")
}

func SchedulerPluginClusterRoleBindingNodeResourceTopology() (*rbacv1.ClusterRoleBinding, error) {
	return loadClusterRoleBinding(ComponentSchedulerPlugin, "node-res-topo")
}

func SchedulerPluginClusterRoleBindingVolumeScheduler() (*rbacv1.ClusterRoleBinding, error) {
	return loadClusterRoleBinding(ComponentSchedulerPlugin, "vol-sched")
}

func SchedulerPluginRoleBindingKubeScheduler() (*rbacv1.RoleBinding, error) {
	obj, err := loadObject("yaml/sched/rolebinding-kube-sched.yaml")
	if err != nil {
		return nil, err
	}

	crb, ok := obj.(*rbacv1.RoleBinding)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return crb, nil
}

func ResourceTopologyExporterClusterRoleBinding() (*rbacv1.ClusterRoleBinding, error) {
	return loadClusterRoleBinding(ComponentResourceTopologyExporter, "")
}

func ResourceTopologyExporterDaemonSet() (*appsv1.DaemonSet, error) {
	obj, err := loadObject("yaml/rte/daemonset.yaml")
	if err != nil {
		return nil, err
	}

	ds, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return nil, fmt.Errorf("unexpected type, got %t", obj)
	}
	return ds, nil
}

func SerializeObject(obj runtime.Object, out io.Writer) error {
	srz := json.NewYAMLSerializer(json.DefaultMetaFactory, scheme.Scheme, scheme.Scheme)
	return srz.Encode(obj, out)
}

func loadObject(path string) (runtime.Object, error) {
	data, err := src.ReadFile(path)
	if err != nil {
		return nil, err
	}

	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func validateComponent(component string) error {
	if component == "api" || component == "rte" || component == "sched" {
		return nil
	}
	return fmt.Errorf("unknown component: %s", component)
}
