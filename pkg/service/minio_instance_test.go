package service

import (
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

var minioTestSpec = NewMinioInstance().GetProvisionSpec(ServiceOptions{
	ID:     "test-id",
	PlanID: "2f931eba-c3cc-4d41-8702-e63cd5ee9a5c",
})

func TestApplyMinioInstanceProvistion(t *testing.T) {
	client := fake.NewSimpleClientset()
	err := minioTestSpec.Create(client)
	if err != nil {
		t.Fatalf("error injecting pod add: %v", err)
	}
}
