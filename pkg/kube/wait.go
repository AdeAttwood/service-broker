package kube

import (
	"context"
	"flag"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	bV1 "k8s.io/client-go/kubernetes/typed/batch/v1"
)

func isJobComplete(client bV1.JobInterface, name string) wait.ConditionFunc {
	return func() (bool, error) {
		if flag.Lookup("test.v") != nil {
			return true, nil
		}

		job, err := client.Get(context.TODO(), name, metaV1.GetOptions{})
		if err != nil {
			return false, err
		}

		if job.Status.Active+job.Status.Succeeded+job.Status.Failed == 0 {
			return false, nil
		}

		if job.Status.Active > 0 {
			return false, nil
		}

		return true, nil
	}
}

func isDeploymentReady(client v1.DeploymentInterface, name string) wait.ConditionFunc {
	return func() (bool, error) {
		if flag.Lookup("test.v") != nil {
			return true, nil
		}

		deployment, err := client.Get(context.TODO(), name, metaV1.GetOptions{})
		if err != nil {
			return false, err
		}

		if deployment.Status.Replicas == 0 {
			return false, nil
		}

		if deployment.Status.UnavailableReplicas > 0 {
			return false, nil
		}

		return true, nil
	}
}
