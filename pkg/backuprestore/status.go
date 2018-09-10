/*
Copyright 2018 The aerospike-operator Authors.

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

package backuprestore

import (
	"encoding/json"
	"time"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"

	"github.com/travelaudience/aerospike-operator/pkg/apis/aerospike/common"
	aerospikev1alpha2 "github.com/travelaudience/aerospike-operator/pkg/apis/aerospike/v1alpha2"
)

func (h *AerospikeBackupRestoreHandler) appendCondition(obj aerospikev1alpha2.BackupRestoreObject, condition apiextensions.CustomResourceDefinitionCondition) error {
	oldBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	condition.LastTransitionTime = metav1.NewTime(time.Now())
	obj.SetConditions(append(obj.GetConditions(), condition))

	newBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	switch obj.GetOperationType() {
	case common.OperationTypeBackup:
		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldBytes, newBytes, &aerospikev1alpha2.AerospikeNamespaceBackup{})
		if err != nil {
			return err
		}
		if _, err = h.aerospikeclientset.AerospikeV1alpha2().AerospikeNamespaceBackups(obj.GetNamespace()).Patch(obj.GetName(), types.MergePatchType, patchBytes); err != nil {
			return err
		}
	case common.OperationTypeRestore:
		patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldBytes, newBytes, &aerospikev1alpha2.AerospikeNamespaceRestore{})
		if err != nil {
			return err
		}
		if _, err = h.aerospikeclientset.AerospikeV1alpha2().AerospikeNamespaceRestores(obj.GetNamespace()).Patch(obj.GetName(), types.MergePatchType, patchBytes); err != nil {
			return err
		}
	}
	return nil
}

func (h *AerospikeBackupRestoreHandler) isFailedOrFinished(obj aerospikev1alpha2.BackupRestoreObject) bool {
	for _, c := range obj.GetConditions() {
		if (c.Type == obj.GetFinishedConditionType() || c.Type == obj.GetFailedConditionType()) && c.Status == apiextensions.ConditionTrue {
			return true
		}
	}
	return false
}
