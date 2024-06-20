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

package functional_test

import (
	. "github.com/onsi/gomega" //revive:disable:dot-imports

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	heatv1 "github.com/openstack-k8s-operators/heat-operator/api/v1beta1"
	condition "github.com/openstack-k8s-operators/lib-common/modules/common/condition"
)

func GetDefaultHeatSpec() map[string]interface{} {
	return map[string]interface{}{
		"databaseInstance": "openstack",
		"secret":           SecretName,
		"heatEngine":       GetDefaultHeatEngineSpec(),
		"heatAPI":          GetDefaultHeatAPISpec(),
		"heatCfnAPI":       GetDefaultHeatCfnAPISpec(),
	}
}

func GetDefaultHeatAPISpec() map[string]interface{} {
	return map[string]interface{}{}
}

func GetDefaultHeatEngineSpec() map[string]interface{} {
	return map[string]interface{}{}
}

func GetDefaultHeatCfnAPISpec() map[string]interface{} {

	serviceOverride := map[string]interface{}{}
	serviceOverride["public"] = map[string]interface{}{
		"endpointURL": "https://heat-cfnapi-public-openstack.test",
		"metadata": map[string]map[string]string{
			"labels": {
				"osctlplane":         "",
				"osctlplane-service": "heat-cfn",
			},
		},
	}
	return map[string]interface{}{
		"databaseHostname":   "openstack",
		"secret":             SecretName,
		"containerImage":     "quay.io/podified-antelope-centos9/openstack-heat-api-cfn:current-podified",
		"serviceAccount":     "heat",
		"transportURLSecret": "rabbitmq-secret",

		"override": map[string]interface{}{
			"service": serviceOverride,
		},
	}
}

func CreateHeat(name types.NamespacedName, spec map[string]interface{}) client.Object {

	raw := map[string]interface{}{
		"apiVersion": "heat.openstack.org/v1beta1",
		"kind":       "Heat",
		"metadata": map[string]interface{}{
			"name":      name.Name,
			"namespace": name.Namespace,
		},
		"spec": spec,
	}
	return th.CreateUnstructured(raw)
}

func CreateHeatCfnAPI(name types.NamespacedName, spec map[string]interface{}) client.Object {

	parent := GetHeat(types.NamespacedName{Namespace: namespace, Name: "heat"})
	raw := map[string]interface{}{
		"apiVersion": "heat.openstack.org/v1beta1",
		"kind":       "HeatCfnAPI",
		"metadata": map[string]interface{}{
			"name":      name.Name,
			"namespace": name.Namespace,
			"ownerReferences": []map[string]interface{}{
				{
					"apiVersion":         "heat.openstack.org/v1beta1",
					"blockOwnerDeletion": true,
					"controller":         true,
					"kind":               "Heat",
					"name":               parent.GetObjectMeta().GetName(),
					"uid":                parent.GetObjectMeta().GetUID(),
				},
			},
		},
		"spec": spec,
	}
	return th.CreateUnstructured(raw)

}

func GetHeat(name types.NamespacedName) *heatv1.Heat {
	instance := &heatv1.Heat{}
	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, name, instance)).Should(Succeed())
	}, timeout, interval).Should(Succeed())
	return instance
}

func GetHeatCfnAPI(name types.NamespacedName) *heatv1.HeatCfnAPI {
	instance := &heatv1.HeatCfnAPI{}
	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, name, instance)).Should(Succeed())
	}, timeout, interval).Should(Succeed())
	return instance
}

func CreateHeatSecret(namespace string, name string) *corev1.Secret {
	return th.CreateSecret(
		types.NamespacedName{Namespace: namespace, Name: name},
		map[string][]byte{
			"HeatPassword":      []byte("12345678"),
			"AuthEncryptionKey": []byte("1234567812345678123456781212345678345678"),
		},
	)
}

func CreateHeatMessageBusSecret(namespace string, name string) *corev1.Secret {
	return th.CreateSecret(
		types.NamespacedName{Namespace: namespace, Name: name},
		map[string][]byte{
			"transport_url": []byte("rabbit://fake"),
		},
	)
}

func HeatConditionGetter(name types.NamespacedName) condition.Conditions {
	instance := GetHeat(name)
	return instance.Status.Conditions
}
