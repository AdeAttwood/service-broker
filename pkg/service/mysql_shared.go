package service

import (
	"fmt"
	"strings"

	"github.com/AdeAttwood/service-broker/pkg/kube"
	osb "github.com/pmorie/go-open-service-broker-client/v2"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SharedMysqlConfig struct {
	Name     string `yaml:"name"`
	ID       string `yaml:"id"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

func NewSharedMysql(config SharedMysqlConfig) *SharedMysql {
	return &SharedMysql{
		name:     config.Name,
		id:       config.ID,
		user:     config.User,
		password: config.Password,
		port:     config.Port,
		host:     config.Host,
	}
}

var mysqlDatabaseBindScript = `
set -ex

export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"

until mysql -u "$MYSQL_USER" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e ";" > /dev/null 2>&1; do
	echo "Waiting for host '$MYSQL_HOST'"
	sleep 5
done

echo "Creating databse '$DB_NAME' and granting privileges to '$DB_USER'"
mysql -u "$MYSQL_USER" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "CREATE SCHEMA IF NOT EXISTS $DB_NAME;"
mysql -u "$MYSQL_USER" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "CREATE USER IF NOT EXISTS '$DB_USER'@'%' IDENTIFIED BY '$DB_PASSWORD';"
mysql -u "$MYSQL_USER" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'%';"
mysql -u "$MYSQL_USER" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "FLUSH PRIVILEGES;"
`

var mysqlDatabaseDebindScript = `
set -ex

export MYSQL_PWD="$MYSQL_ROOT_PASSWORD"

until mysql -u "$MYSQL_USER" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e ";" > /dev/null 2>&1; do
	echo "Waiting for host '$MYSQL_HOST'"
	sleep 5
done

echo "Removing user '$DB_USER'"
mysql -u "$MYSQL_USER" -h "$MYSQL_HOST" -P "$MYSQL_PORT" -e "DROP USER '$DB_USER';"
`

type SharedMysql struct {
	name     string `yaml:"name"`
	id       string `yaml:"id"`
	user     string `yaml:"user"`
	password string `yaml:"password"`
	host     string `yaml:"host"`
	port     string `yaml:"port"`
}

func (s *SharedMysql) Definition() osb.Service {
	return osb.Service{
		Name:        fmt.Sprintf("mysql-shared-%s", s.name),
		ID:          s.id,
		Description: "A database on a shared mysql instance",
		Bindable:    true,
		Metadata: map[string]interface{}{
			"displayName": "Shared Mysql Database",
			"imageUrl":    "htps://avatars2.githubusercontent.com/u/19862012?s=200&v=4",
		},
		Plans: []osb.Plan{
			{
				Name:        "default",
				ID:          s.id,
				Description: "The default plan",
				Free:        truePtr(),
			},
		},
	}
}

func (s *SharedMysql) GetHost(instanceID string, namespace string) string {
	return s.host
}

func (s *SharedMysql) GetDebindSpec(options BindOptions) *kube.Spec {
	secretName := fmt.Sprintf("mysql-shared-%s-secret", s.name)
	bindingSecretName := fmt.Sprintf("binding-secret-%s", options.ID)

	return &kube.Spec{
		Namespace: options.GlobalNamespace,
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
									Command: []string{"bash", "-c", mysqlDatabaseDebindScript},
									Env: []coreV1.EnvVar{
										kube.EnvSecret("MYSQL_ROOT_PASSWORD", secretName, "password"),
										kube.EnvSecret("MYSQL_USER", secretName, "user"),
										kube.EnvSecret("MYSQL_HOST", secretName, "host"),
										kube.EnvSecret("MYSQL_PORT", secretName, "port"),

										kube.EnvSecret("DB_USER", bindingSecretName, "user"),
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

func (s *SharedMysql) GetBindSpec(options BindOptions) *kube.Spec {
	secretName := fmt.Sprintf("mysql-shared-%s-secret", s.name)
	bindingSecretName := fmt.Sprintf("binding-secret-%s", options.ID)
	databaseName := strings.Replace(fmt.Sprintf("%s_%s", options.Namespace, options.ID[0:8]), "-", "_", -1)

	return &kube.Spec{
		Namespace: options.GlobalNamespace,
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
					"port":     []byte(s.port),
					"database": []byte(databaseName),
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
									Command: []string{"bash", "-c", mysqlDatabaseBindScript},
									Env: []coreV1.EnvVar{
										kube.EnvSecret("MYSQL_ROOT_PASSWORD", secretName, "password"),
										kube.EnvSecret("MYSQL_USER", secretName, "user"),
										kube.EnvSecret("MYSQL_HOST", secretName, "host"),
										kube.EnvSecret("MYSQL_PORT", secretName, "port"),

										kube.EnvSecret("DB_NAME", bindingSecretName, "database"),
										kube.EnvSecret("DB_USER", bindingSecretName, "user"),
										kube.EnvSecret("DB_PASSWORD", bindingSecretName, "password"),
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

func (s *SharedMysql) GetDeprovisionSpec(options ServiceOptions) *kube.Spec {
	return &kube.Spec{Namespace: options.Namespace}
}

func (s *SharedMysql) GetProvisionSpec(options ServiceOptions) *kube.Spec {
	secretName := fmt.Sprintf("mysql-shared-%s-secret", s.name)

	return &kube.Spec{
		Namespace: options.GlobalNamespace,
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
					"user":     []byte(s.user),
					"password": []byte(s.password),
					"host":     []byte(s.host),
					"port":     []byte(s.port),
				},
			},
		},
	}
}
