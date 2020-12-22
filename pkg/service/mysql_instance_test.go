package service

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

var spec = NewMysqlInstance().GetProvisionSpec(ServiceOptions{
	ID:     "test-id",
	PlanID: "86064792-7ea2-467b-af93-ac9694d96d5b",
})

func TestInjectLables(t *testing.T) {
	spec.InjectLabels(spec.Lables)
	serviceIdLabel := spec.Deployments[0].Labels["service-instance-id"]
	if serviceIdLabel != "test-id" {
		t.Errorf("Invalid service ID label '%s'", serviceIdLabel)
	}
}

func TestMySqlImage(t *testing.T) {
	mysqlDeployment := spec.Deployments[0]
	image := mysqlDeployment.Spec.Template.Spec.Containers[0].Image
	if image != "mysql:5.7" {
		t.Errorf("Invalid deployment image '%s'", image)
	}
}

func TestApply(t *testing.T) {
	client := fake.NewSimpleClientset()

	err := spec.Create(client)
	if err != nil {
		t.Fatalf("error injecting pod add: %v", err)
	}
}
