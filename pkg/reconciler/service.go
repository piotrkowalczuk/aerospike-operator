/*
Copyright 2018 The aerospike-controller Authors.

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

package reconciler

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	aerospikev1alpha1 "github.com/travelaudience/aerospike-operator/pkg/apis/aerospike/v1alpha1"
	"github.com/travelaudience/aerospike-operator/pkg/logfields"
	"github.com/travelaudience/aerospike-operator/pkg/meta"
	"github.com/travelaudience/aerospike-operator/pkg/pointers"
)

func (r *AerospikeClusterReconciler) ensureClientService(aerospikeCluster *aerospikev1alpha1.AerospikeCluster) error {
	serviceName := fmt.Sprintf("%s-%s", aerospikeCluster.Name, clientServiceSuffix)

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				labelAppKey:     labelAppVal,
				labelClusterKey: aerospikeCluster.Name,
			},
			Namespace: aerospikeCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         aerospikev1alpha1.SchemeGroupVersion.String(),
					Kind:               kind,
					Name:               aerospikeCluster.Name,
					UID:                aerospikeCluster.UID,
					Controller:         pointers.NewBool(true),
					BlockOwnerDeletion: pointers.NewBool(true),
				},
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				labelAppKey:     labelAppVal,
				labelClusterKey: aerospikeCluster.Name,
			},
			Ports: []v1.ServicePort{
				{
					Name:       servicePortName,
					Port:       servicePort,
					TargetPort: intstr.IntOrString{StrVal: servicePortName},
				},
			},
		},
	}

	if _, err := r.kubeclientset.CoreV1().Services(aerospikeCluster.Namespace).Create(service); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		log.WithFields(log.Fields{
			logfields.AerospikeCluster: meta.Key(aerospikeCluster),
			logfields.Service:          service.Name,
		}).Debug("client service already exists")
		return nil
	}

	log.WithFields(log.Fields{
		logfields.AerospikeCluster: meta.Key(aerospikeCluster),
		logfields.Service:          service.Name,
	}).Debug("client service created")
	return nil
}

func (r *AerospikeClusterReconciler) ensureHeadlessService(aerospikeCluster *aerospikev1alpha1.AerospikeCluster) error {
	serviceName := fmt.Sprintf("%s-%s", aerospikeCluster.Name, discoveryServiceSuffix)

	service := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			Labels: map[string]string{
				labelAppKey:     labelAppVal,
				labelClusterKey: aerospikeCluster.Name,
			},
			Namespace: aerospikeCluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         aerospikev1alpha1.SchemeGroupVersion.String(),
					Kind:               kind,
					Name:               aerospikeCluster.Name,
					UID:                aerospikeCluster.UID,
					Controller:         pointers.NewBool(true),
					BlockOwnerDeletion: pointers.NewBool(true),
				},
			},
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				labelAppKey:     labelAppVal,
				labelClusterKey: aerospikeCluster.Name,
			},
			Ports: []v1.ServicePort{
				{
					Name:       heartbeatPortName,
					Port:       heartbeatPort,
					TargetPort: intstr.IntOrString{StrVal: heartbeatPortName},
				},
			},
			ClusterIP: v1.ClusterIPNone,
		},
	}

	if _, err := r.kubeclientset.CoreV1().Services(aerospikeCluster.Namespace).Create(service); err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
		log.WithFields(log.Fields{
			logfields.AerospikeCluster: meta.Key(aerospikeCluster),
			logfields.Service:          service.Name,
		}).Debug("headless service already exists")
		return nil
	}

	log.WithFields(log.Fields{
		logfields.AerospikeCluster: meta.Key(aerospikeCluster),
		logfields.Service:          service.Name,
	}).Debug("headless service created")
	return nil
}