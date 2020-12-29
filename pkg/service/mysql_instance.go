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

func NewMysqlInstance() *MysqlInstance {
	return &MysqlInstance{}
}

type MysqlInstance struct {
}

// Get the service definition of the mysql instance
func (s *MysqlInstance) Definition() osb.Service {
	return osb.Service{
		Name:        "mysql-instance",
		ID:          "4f6e6cf6-ffdd-425f-a2c7-3c9258ad246a",
		Description: "A mysql instance deployment",
		Bindable:    true,
		Metadata: map[string]interface{}{
			"displayName": "MySql Instance",
			"imageUrl":    "https://avatars2.githubusercontent.com/u/19862012?s=200&v=4",
		},
		Plans: []osb.Plan{
			{
				Name:        "default",
				ID:          "86064792-7ea2-467b-af93-ac9694d96d5b",
				Description: "The default plan",
				Free:        truePtr(),
			},
		},
	}
}

func (s *MysqlInstance) GetHost(instanceID string, namespace string) string {
	return fmt.Sprintf("mysql-instance-%s.%s.svc.cluster.local", instanceID, namespace)
}

func (s *MysqlInstance) GetDebindSpec(options BindOptions) *kube.Spec {
	deploymentHost := s.GetHost(options.InstanceID, options.Namespace)
	deploymentName := fmt.Sprintf("mysql-instance-%s", options.InstanceID)
	rootSecretName := fmt.Sprintf("%s-root-secret", deploymentName)
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
									Name:    "mysql",
									Image:   "mysql:5.7",
									Command: []string{"bash", "/tmp/debind.bash"},
									Env: []coreV1.EnvVar{
										{
											Name:  "MYSQL_HOST",
											Value: deploymentHost,
										},
										kube.EnvSecret("MYSQL_ROOT_PASSWORD", rootSecretName, "password"),
										kube.EnvSecret("DB_USER", bindingSecretName, "user"),
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

func (s *MysqlInstance) GetBindSpec(options BindOptions) *kube.Spec {
	deploymentHost := s.GetHost(options.InstanceID, options.Namespace)
	deploymentName := fmt.Sprintf("mysql-instance-%s", options.InstanceID)
	rootSecretName := fmt.Sprintf("%s-root-secret", deploymentName)
	bindingSecretName := fmt.Sprintf("binding-secret-%s", options.ID)

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
					"host":     []byte(s.GetHost(options.InstanceID, options.Namespace)),
					"user":     []byte(fmt.Sprintf("user-%s", kube.RandStringBytes(8))),
					"database": []byte("service_database"),
					"password": []byte(kube.RandStringBytes(18)),
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
									Name:    "mysql",
									Image:   "mysql:5.7",
									Command: []string{"bash", "/tmp/bind.bash"},
									Env: []coreV1.EnvVar{
										{
											Name:  "MYSQL_HOST",
											Value: deploymentHost,
										},
										kube.EnvSecret("MYSQL_ROOT_PASSWORD", rootSecretName, "password"),
										kube.EnvSecret("DB_NAME", bindingSecretName, "database"),
										kube.EnvSecret("DB_USER", bindingSecretName, "user"),
										kube.EnvSecret("DB_PASSWORD", bindingSecretName, "password"),
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

func (s *MysqlInstance) GetDeprovisionSpec(options ServiceOptions) *kube.Spec {
	return &kube.Spec{Namespace: options.Namespace}
}

func (s *MysqlInstance) GetProvisionSpec(options ServiceOptions) *kube.Spec {
	deploymentName := fmt.Sprintf("mysql-instance-%s", options.ID)
	secretName := fmt.Sprintf("%s-root-secret", deploymentName)
	pvcName := fmt.Sprintf("%s-pvc", deploymentName)

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
					"password": []byte(kube.RandStringBytes(16)),
				},
			},
		},
		ConfigMaps: []coreV1.ConfigMap{
			{
				ObjectMeta: metaV1.ObjectMeta{
					Name: deploymentName,
				},
				Data: map[string]string{
					"backup.bash": `
set -e

export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"
DBS=$(mysql -uroot -h "$MYSQL_HOST" -e 'show databases' -s --skip-column-names | grep -Ev "(mysql|information_schema|performance_schema)");

BACKUP_DIR=/var/lib/mysql/backups;
test -d "$BACKUP_DIR" || mkdir -p "$BACKUP_DIR"

for db in $DBS; do
	test -d "$BACKUP_DIR/$db" || mkdir -p "$BACKUP_DIR/$db"
	for table in $(mysql -uroot -h "$MYSQL_HOST" -e 'show tables' $db -s --skip-column-names); do
		echo "Backing up '$db/$table'"
		mysqldump -uroot -h "$MYSQL_HOST" $db $table > "$BACKUP_DIR/$db/$table.sql"
	done
done
`,
					"bind.bash": `
set -e

export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"

until mysql -uroot -h "$MYSQL_HOST" -e ";" > /dev/null 2>&1; do
    echo "Waiting for host '$MYSQL_HOST'"
done

echo "Creating databse '$DB_NAME' and granting privileges to '$DB_USER'"
mysql -uroot -h "$MYSQL_HOST" -e "CREATE SCHEMA IF NOT EXISTS $DB_NAME;"
mysql -uroot -h "$MYSQL_HOST" -e "CREATE USER IF NOT EXISTS '$DB_USER'@'%' IDENTIFIED BY '$DB_PASSWORD';"
mysql -uroot -h "$MYSQL_HOST" -e "GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'%';"
mysql -uroot -h "$MYSQL_HOST" -e "FLUSH PRIVILEGES;"
`,
					"debind.bash": `
set -e

export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"

until mysql -uroot -h "$MYSQL_HOST" -e ";" > /dev/null 2>&1; do
    echo "Waiting for host '$MYSQL_HOST'"
done

echo "Removing user '$DB_USER'"
mysql -uroot -h "$MYSQL_HOST" -e "DROP USER '$DB_USER';"
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
									Name:  "mysql",
									Image: "mysql:5.7",
									Ports: []coreV1.ContainerPort{
										{
											Name:          "tpc",
											Protocol:      coreV1.ProtocolTCP,
											ContainerPort: 3306,
										},
									},
									Env: []coreV1.EnvVar{
										{
											Name: "MYSQL_ROOT_PASSWORD",
											ValueFrom: &coreV1.EnvVarSource{
												SecretKeyRef: &coreV1.SecretKeySelector{
													LocalObjectReference: coreV1.LocalObjectReference{Name: secretName},
													Key:                  "password",
												},
											},
										},
									},
									ReadinessProbe: &coreV1.Probe{
										Handler: coreV1.Handler{
											TCPSocket: &coreV1.TCPSocketAction{
												Port: intstr.IntOrString{
													Type:   intstr.Int,
													IntVal: 3306,
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
											MountPath: "/var/lib/mysql",
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
							Port: 3306,
							TargetPort: intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 3306,
							},
						},
					},
				},
			},
		},
	}
}
