package kube

import (
	"context"
	"fmt"

	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Spec struct {
	Namespace   string
	Lables      map[string]string
	Secrets     []coreV1.Secret
	ConfigMaps  []coreV1.ConfigMap
	PVCS        []coreV1.PersistentVolumeClaim
	Deployments []appsV1.Deployment
	Services    []coreV1.Service
	Jobs        []batchV1.Job
}

func (s *Spec) InjectLabels(labels map[string]string) {
	for label, value := range labels {
		for i := 0; i < len(s.Secrets); i++ {
			if s.Secrets[i].ObjectMeta.Labels == nil {
				s.Secrets[i].ObjectMeta.Labels = map[string]string{}
			}

			s.Secrets[i].ObjectMeta.Labels[label] = value
		}

		for i := 0; i < len(s.ConfigMaps); i++ {
			if s.ConfigMaps[i].ObjectMeta.Labels == nil {
				s.ConfigMaps[i].ObjectMeta.Labels = map[string]string{}
			}

			s.ConfigMaps[i].ObjectMeta.Labels[label] = value
		}

		for i := 0; i < len(s.PVCS); i++ {
			if s.PVCS[i].ObjectMeta.Labels == nil {
				s.PVCS[i].ObjectMeta.Labels = map[string]string{}
			}

			s.PVCS[i].ObjectMeta.Labels[label] = value
		}

		for i := 0; i < len(s.Deployments); i++ {
			if s.Deployments[i].ObjectMeta.Labels == nil {
				s.Deployments[i].ObjectMeta.Labels = map[string]string{}
			}

			s.Deployments[i].ObjectMeta.Labels[label] = value
		}

		for i := 0; i < len(s.Services); i++ {
			if s.Services[i].ObjectMeta.Labels == nil {
				s.Services[i].ObjectMeta.Labels = map[string]string{}
			}

			s.Services[i].ObjectMeta.Labels[label] = value
		}

		for i := 0; i < len(s.Jobs); i++ {
			if s.Jobs[i].ObjectMeta.Labels == nil {
				s.Jobs[i].ObjectMeta.Labels = map[string]string{}
			}

			s.Jobs[i].ObjectMeta.Labels[label] = value
		}
	}

}

func (s *Spec) Delete(client kubernetes.Interface) error {
	for i := 0; i < len(s.Jobs); i++ {
		jobSpec := &s.Jobs[i]
		jobClient := client.BatchV1().Jobs(s.Namespace)
		jobErr := jobClient.Delete(context.TODO(), jobSpec.Name, metaV1.DeleteOptions{})
		if jobErr != nil {
			return jobErr
		}
		fmt.Printf("Deleted job %q.\n", jobSpec.Name)
	}

	for i := 0; i < len(s.Services); i++ {
		serviceSpec := &s.Services[i]
		serviceClient := client.CoreV1().Services(s.Namespace)
		serviceErr := serviceClient.Delete(context.TODO(), serviceSpec.Name, metaV1.DeleteOptions{})
		if serviceErr != nil {
			return serviceErr
		}
		fmt.Printf("Deleted service %q.\n", serviceSpec.Name)
	}

	for i := 0; i < len(s.Deployments); i++ {
		deploymentSpec := &s.Deployments[i]
		deploymentClient := client.AppsV1().Deployments(s.Namespace)
		deploymentErr := deploymentClient.Delete(context.TODO(), deploymentSpec.Name, metaV1.DeleteOptions{})
		if deploymentErr != nil {
			return deploymentErr
		}
		fmt.Printf("Deleted deployment %q.\n", deploymentSpec.Name)
	}

	for i := 0; i < len(s.PVCS); i++ {
		pvcSpec := &s.PVCS[i]
		pvcClient := client.CoreV1().PersistentVolumeClaims(s.Namespace)
		pvcErr := pvcClient.Delete(context.TODO(), pvcSpec.Name, metaV1.DeleteOptions{})
		if pvcErr != nil {
			return pvcErr
		}
		fmt.Printf("Deleted pvc %q.\n", pvcSpec.Name)
	}

	for i := 0; i < len(s.ConfigMaps); i++ {
		configMapSpec := &s.ConfigMaps[i]
		configMapClient := client.CoreV1().ConfigMaps(s.Namespace)
		configMapErr := configMapClient.Delete(context.TODO(), configMapSpec.Name, metaV1.DeleteOptions{})
		if configMapErr != nil {
			return configMapErr
		}
		fmt.Printf("Deleted config map %q.\n", configMapSpec.Name)
	}

	for i := 0; i < len(s.Secrets); i++ {
		secretSpec := &s.Secrets[i]
		secretsClient := client.CoreV1().Secrets(s.Namespace)
		secretErr := secretsClient.Delete(context.TODO(), secretSpec.Name, metaV1.DeleteOptions{})
		if secretErr != nil {
			return secretErr
		}
		fmt.Printf("Deleted secret %q.\n", secretSpec.Name)
	}

	return nil
}

func (s *Spec) Create(client kubernetes.Interface) error {
	s.InjectLabels(s.Lables)

	for i := 0; i < len(s.Secrets); i++ {
		secretSpec := &s.Secrets[i]
		secretsClient := client.CoreV1().Secrets(s.Namespace)
		secret, secretErr := secretsClient.Create(context.TODO(), secretSpec, metaV1.CreateOptions{})
		if secretErr != nil {
			return secretErr
		}
		fmt.Printf("Created secret %q.\n", secret.GetObjectMeta().GetName())
	}

	for i := 0; i < len(s.ConfigMaps); i++ {
		configMapSpec := &s.ConfigMaps[i]
		configMapClient := client.CoreV1().ConfigMaps(s.Namespace)
		configMap, configMapErr := configMapClient.Create(context.TODO(), configMapSpec, metaV1.CreateOptions{})
		if configMapErr != nil {
			return configMapErr
		}
		fmt.Printf("Created config map %q.\n", configMap.GetObjectMeta().GetName())
	}

	for i := 0; i < len(s.PVCS); i++ {
		pvcSpec := &s.PVCS[i]
		pvcClient := client.CoreV1().PersistentVolumeClaims(s.Namespace)
		pvc, pvcErr := pvcClient.Create(context.TODO(), pvcSpec, metaV1.CreateOptions{})
		if pvcErr != nil {
			return pvcErr
		}
		fmt.Printf("Created pvc %q.\n", pvc.GetObjectMeta().GetName())
	}

	for i := 0; i < len(s.Deployments); i++ {
		deploymentSpec := &s.Deployments[i]
		deploymentClient := client.AppsV1().Deployments(s.Namespace)
		deployment, deploymentErr := deploymentClient.Create(context.TODO(), deploymentSpec, metaV1.CreateOptions{})
		if deploymentErr != nil {
			return deploymentErr
		}
		fmt.Printf("Created deployment %q.\n", deployment.GetObjectMeta().GetName())
	}

	for i := 0; i < len(s.Services); i++ {
		serviceSpec := &s.Services[i]
		serviceClient := client.CoreV1().Services(s.Namespace)
		service, serviceErr := serviceClient.Create(context.TODO(), serviceSpec, metaV1.CreateOptions{})
		if serviceErr != nil {
			return serviceErr
		}
		fmt.Printf("Created service %q.\n", service.GetObjectMeta().GetName())
	}

	for i := 0; i < len(s.Jobs); i++ {
		jobSpec := &s.Jobs[i]
		jobClient := client.BatchV1().Jobs(s.Namespace)
		job, jobErr := jobClient.Create(context.TODO(), jobSpec, metaV1.CreateOptions{})
		if jobErr != nil {
			return jobErr
		}
		fmt.Printf("Created job %q.\n", job.GetObjectMeta().GetName())
	}

	return nil
}
