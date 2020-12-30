package service

import (
	"fmt"

	osb "github.com/pmorie/go-open-service-broker-client/v2"

	"github.com/AdeAttwood/service-broker/pkg/kube"

	appsV1 "k8s.io/api/apps/v1"
	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewMinioInstance() *MinioInstance {
	return &MinioInstance{}
}

type MinioInstance struct{}

// Get the service definition of the minio instance
func (s *MinioInstance) Definition() osb.Service {
	return osb.Service{
		Name:        "minio-instance",
		ID:          "2a661d27-20a0-40f1-9320-15ea144a694c",
		Description: "A minio instance deployment",
		Bindable:    true,
		Metadata: map[string]interface{}{
			"displayName": "Minio Instance",
			"imageUrl":    "htps://avatars2.githubusercontent.com/u/19862012?s=200&v=4",
		},
		Plans: []osb.Plan{
			{
				Name:        "default",
				ID:          "2f931eba-c3cc-4d41-8702-e63cd5ee9a5c",
				Description: "The default plan",
				Free:        truePtr(),
			},
		},
	}
}

func (s *MinioInstance) GetHost(instanceID string, namespace string) string {
	return fmt.Sprintf("minio-instance-%s.%s.svc.cluster.local", instanceID, namespace)
}

func (s *MinioInstance) GetDebindSpec(options BindOptions) *kube.Spec {
	deploymentName := fmt.Sprintf("minio-instance-%s", options.InstanceID)
	adminSecretName := fmt.Sprintf("%s-admin-secret", deploymentName)
	bindingSecretName := fmt.Sprintf("binding-secret-%s", options.ID)

	return &kube.Spec{
		Namespace: options.Namespace,
		Lables: map[string]string{
			"service-binding-id":  options.ID,
			"service-instance-id": options.InstanceID,
			"service-id":          s.Definition().ID,
			"service-name":        s.Definition().Name,
		},
		Jobs: []batchV1.Job{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: fmt.Sprintf("debinding-job-%s", options.ID),
				},
				Spec: batchV1.JobSpec{
					Template: coreV1.PodTemplateSpec{
						Spec: coreV1.PodSpec{
							RestartPolicy:         coreV1.RestartPolicyOnFailure,
							ActiveDeadlineSeconds: int64Ptr(120),
							Containers: []coreV1.Container{
								{
									Name:    "mc",
									Image:   "minio/mc:latest",
									Command: []string{"bash", "/tmp/debind.bash"},
									Env: []coreV1.EnvVar{
										kube.EnvSecret("MC_HOST_myminio", adminSecretName, "minioalias"),
										kube.EnvSecret("MINIO_USER", bindingSecretName, "user"),
									},
									VolumeMounts: []coreV1.VolumeMount{
										{
											Name:      "config-volume",
											MountPath: "/tmp/debind.bash",
											ReadOnly:  true,
											SubPath:   "debind.bash",
										},
									},
								},
							},
							Volumes: []coreV1.Volume{
								{
									Name: "config-volume",
									VolumeSource: coreV1.VolumeSource{
										ConfigMap: &coreV1.ConfigMapVolumeSource{
											LocalObjectReference: coreV1.LocalObjectReference{
												Name: deploymentName,
											},
											DefaultMode: int32Ptr(500),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (s *MinioInstance) GetBindSpec(options BindOptions) *kube.Spec {
	deploymentHost := s.GetHost(options.InstanceID, options.Namespace)
	deploymentName := fmt.Sprintf("minio-instance-%s", options.InstanceID)
	adminSecretName := fmt.Sprintf("%s-admin-secret", deploymentName)
	bindingSecretName := fmt.Sprintf("binding-secret-%s", options.ID)

	user := fmt.Sprintf("minio-%s", options.ID)
	password := kube.RandStringBytes(32)

	return &kube.Spec{
		Namespace: options.Namespace,
		Lables: map[string]string{
			"service-binding-id":  options.ID,
			"service-instance-id": options.InstanceID,
			"service-id":          s.Definition().ID,
			"service-name":        s.Definition().Name,
		},
		Secrets: []coreV1.Secret{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: bindingSecretName,
				},
				Type: "Opaque",
				Data: map[string][]byte{
					"user":       []byte(user),
					"password":   []byte(password),
					"host":       []byte(deploymentHost),
					"bucket":     []byte("my-bucket"),
					"minioalias": []byte(fmt.Sprintf("http://%s:%s@%s:9000", user, password, deploymentHost)),
				},
			},
		},
		Jobs: []batchV1.Job{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: fmt.Sprintf("binding-job-%s", options.ID),
				},
				Spec: batchV1.JobSpec{
					Template: coreV1.PodTemplateSpec{
						Spec: coreV1.PodSpec{
							RestartPolicy:         coreV1.RestartPolicyOnFailure,
							ActiveDeadlineSeconds: int64Ptr(120),
							Containers: []coreV1.Container{
								{
									Name:    "mc",
									Image:   "minio/mc:latest",
									Command: []string{"bash", "/tmp/bind.bash"},
									Env: []coreV1.EnvVar{
										kube.EnvSecret("MC_HOST_myminio", adminSecretName, "minioalias"),
										kube.EnvSecret("MINIO_USER", bindingSecretName, "user"),
										kube.EnvSecret("MINIO_PASSWORD", bindingSecretName, "password"),
										kube.EnvSecret("MINIO_BUCKET", bindingSecretName, "bucket"),
									},
									VolumeMounts: []coreV1.VolumeMount{
										{
											Name:      "config-volume",
											MountPath: "/tmp/bind.bash",
											ReadOnly:  true,
											SubPath:   "bind.bash",
										},
									},
								},
							},
							Volumes: []coreV1.Volume{
								{
									Name: "config-volume",
									VolumeSource: coreV1.VolumeSource{
										ConfigMap: &coreV1.ConfigMapVolumeSource{
											LocalObjectReference: coreV1.LocalObjectReference{
												Name: deploymentName,
											},
											DefaultMode: int32Ptr(500),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (s *MinioInstance) GetDeprovisionSpec(options ServiceOptions) *kube.Spec {
	return &kube.Spec{Namespace: options.Namespace}
}

func (s *MinioInstance) GetProvisionSpec(options ServiceOptions) *kube.Spec {
	deploymentName := fmt.Sprintf("minio-instance-%s", options.ID)
	secretName := fmt.Sprintf("%s-admin-secret", deploymentName)
	pvcName := fmt.Sprintf("%s-pvc", deploymentName)

	user := fmt.Sprintf("minio-%s", options.ID)
	password := kube.RandStringBytes(32)

	return &kube.Spec{
		Namespace: options.Namespace,
		Lables: map[string]string{
			"service-instance-id": options.ID,
			"service-id":          s.Definition().ID,
			"service-name":        s.Definition().Name,
			"service-plan":        options.PlanID,
		},
		Secrets: []coreV1.Secret{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: secretName,
				},
				Type: "Opaque",
				Data: map[string][]byte{
					"user":       []byte(user),
					"password":   []byte(password),
					"minioalias": []byte(fmt.Sprintf("http://%s:%s@%s:9000", user, password, s.GetHost(options.ID, options.Namespace))),
				},
			},
		},
		ConfigMaps: []coreV1.ConfigMap{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: deploymentName,
				},
				Data: map[string]string{
					"bind.bash": `
set -e

until mc ls myminio > /dev/null 2>&1; do
    echo "Waiting for minio"
done

cat > /tmp/policy.json <<EOF
{
 "Version": "2012-10-17",
 "Statement": [
  {
   "Effect": "Allow",
   "Action": [
    "s3:*"
   ],
   "Resource": [
     "arn:aws:s3:::$MINIO_BUCKET/*"
   ]
  }
 ]
}
EOF

mc mb "myminio/$MINIO_BUCKET" || true
mc admin user add myminio "$MINIO_USER" "$MINIO_PASSWORD"
mc admin policy add myminio "policy-$MINIO_USER" /tmp/policy.json
mc admin policy set myminio "policy-$MINIO_USER" user=$MINIO_USER
`,
					"debind.bash": `
set -e

until mc ls myminio > /dev/null 2>&1; do
    echo "Waiting for minio"
done

mc admin user remove myminio "$MINIO_USER"
mc admin policy remove myminio "policy-$MINIO_USER"
`,
				},
			},
		},
		PVCS: []coreV1.PersistentVolumeClaim{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: pvcName,
				},
				Spec: coreV1.PersistentVolumeClaimSpec{
					AccessModes: []coreV1.PersistentVolumeAccessMode{
						"ReadWriteOnce",
					},
					Resources: coreV1.ResourceRequirements{
						Requests: coreV1.ResourceList{
							"storage": resource.MustParse("2Gi"),
						},
					},
				},
			},
		},
		Deployments: []appsV1.Deployment{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: deploymentName,
				},
				Spec: appsV1.DeploymentSpec{
					Replicas: int32Ptr(1),
					Selector: &metaV1.LabelSelector{
						MatchLabels: map[string]string{
							"app": deploymentName,
						},
					},
					Template: coreV1.PodTemplateSpec{
						ObjectMeta: metaV1.ObjectMeta{
							Labels: map[string]string{
								"app": deploymentName,
							},
						},
						Spec: coreV1.PodSpec{
							Containers: []coreV1.Container{
								{
									Name:    "minio",
									Image:   "minio/minio:latest",
									Command: []string{"minio", "server", "/data"},
									Ports: []coreV1.ContainerPort{
										{
											Name:          "tpc",
											Protocol:      coreV1.ProtocolTCP,
											ContainerPort: 9000,
										},
									},
									Env: []coreV1.EnvVar{
										kube.EnvSecret("MINIO_ACCESS_KEY", secretName, "user"),
										kube.EnvSecret("MINIO_SECRET_KEY", secretName, "password"),
									},
									ReadinessProbe: &coreV1.Probe{
										Handler: coreV1.Handler{
											TCPSocket: &coreV1.TCPSocketAction{
												Port: intstr.IntOrString{
													Type:   intstr.Int,
													IntVal: 9000,
												},
											},
										},
										FailureThreshold:    1,
										SuccessThreshold:    1,
										TimeoutSeconds:      2,
										InitialDelaySeconds: 10,
										PeriodSeconds:       10,
									},
									VolumeMounts: []coreV1.VolumeMount{
										{
											Name:      pvcName,
											MountPath: "/data",
										},
									},
								},
							},
							Volumes: []coreV1.Volume{
								{
									Name: pvcName,
									VolumeSource: coreV1.VolumeSource{
										PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
											ClaimName: pvcName,
										},
									},
								},
							},
						},
					},
				},
			},
		},
		Services: []coreV1.Service{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: deploymentName,
				},
				Spec: coreV1.ServiceSpec{
					Selector: map[string]string{
						"app": deploymentName,
					},
					Type: "LoadBalancer",
					Ports: []coreV1.ServicePort{
						{
							Port: 9000,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 9000,
							},
						},
					},
				},
			},
		},
	}
}
