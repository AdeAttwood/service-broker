package broker

import (
	"net/http/httptest"
	"testing"

	"github.com/pmorie/osb-broker-lib/pkg/broker"
	"k8s.io/client-go/kubernetes/fake"
)

var logic, _ = NewBusinessLogic(Options{
	Async:            true,
	ServiceNamespace: "service-broker",
	K8sClient:        fake.NewSimpleClientset(),
})

func mocRequest() *broker.RequestContext {
	return &broker.RequestContext{
		Writer:  httptest.NewRecorder(),
		Request: httptest.NewRequest("GET", "http://test.com", nil),
	}
}

func TestGetCatalog(t *testing.T) {
	res, err := logic.GetCatalog(mocRequest())
	if err != nil {
		t.Error(err.Error())
	}

	if res.CatalogResponse.Services[0].Name != "mysql-instance" {
		t.Errorf("Invalid service name '%s'", res.Services[0].Name)
	}
}
