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
	"strconv"
	"strings"
	"time"

	as "github.com/aerospike/aerospike-client-go"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"

	"github.com/travelaudience/aerospike-operator/pkg/logfields"
	"github.com/travelaudience/aerospike-operator/pkg/meta"
	"github.com/travelaudience/aerospike-operator/pkg/utils/selectors"
)

const (
	// migrationTimeout is the amount of time we will wait for migrations to complete before deleting the pod
	migrationTimeout = 1 * time.Hour
)

type byIndex []*v1.Pod

func (p byIndex) Len() int {
	return len(p)
}

func (p byIndex) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p byIndex) Less(i, j int) bool {
	idx1 := podIndex(p[i])
	idx2 := podIndex(p[j])
	return idx1 < idx2
}

func podIndex(pod *v1.Pod) int {
	res, err := strconv.Atoi(strings.TrimPrefix(pod.Name, fmt.Sprintf("%s-", pod.ObjectMeta.Labels[selectors.LabelClusterKey])))
	if err != nil {
		return -1
	}
	return res
}

func isPodRunningAndReady(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodRunning && podutil.IsPodReady(pod)
}

func waitPodReadyToShutdown(pod *v1.Pod) error {
	client, err := as.NewClient(pod.Status.PodIP, servicePort)
	if err != nil {
		return err
	}
	defer client.Close()
	for _, node := range client.GetNodes() {
		if node.GetHost().Name == pod.Status.PodIP {
			log.WithFields(log.Fields{
				logfields.AerospikeCluster: pod.Labels[selectors.LabelClusterKey],
				logfields.Pod:              meta.Key(pod),
			}).Debug("waiting for migrations to finish")
			return node.WaitUntillMigrationIsFinished(migrationTimeout)
		}
	}
	return nil
}
