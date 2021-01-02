package broker

import (
	"context"
	"fmt"
	"sync"

	"github.com/pmorie/osb-broker-lib/pkg/broker"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/AdeAttwood/service-broker/pkg/service"

	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

// NewBusinessLogic is a hook that is called with the Options the program is run
// with. NewBusinessLogic is the place where you will initialize your
// BusinessLogic the parameters passed in.
func NewBusinessLogic(o Options) (*BusinessLogic, error) {
	// For example, if your BusinessLogic requires a parameter from the command
	// line, you would unpack it from the Options and set it on the
	// BusinessLogic here.
	return &BusinessLogic{
		async:     o.Async,
		k8sClient: o.K8sClient,
		namespace: o.ServiceNamespace,
		services: map[string]service.Service{
			"4f6e6cf6-ffdd-425f-a2c7-3c9258ad246a": service.NewMysqlInstance(),
			"2a661d27-20a0-40f1-9320-15ea144a694c": service.NewMinioInstance(),
		},
	}, nil
}

// BusinessLogic provides an implementation of the broker.BusinessLogic
// interface.
type BusinessLogic struct {
	// Indicates if the broker should handle the requests asynchronously.
	async bool
	// Synchronize go routines.
	sync.RWMutex
	// The available services in this service broker
	services map[string]service.Service
	// The kubernetes client that will be used to create all of the service in
	// the cluster
	k8sClient kubernetes.Interface
	// The namespace that all of the global services will be created in
	namespace string
}

var _ broker.Interface = &BusinessLogic{}

func truePtr() *bool {
	b := true
	return &b
}

func (b *BusinessLogic) GetCatalog(c *broker.RequestContext) (*broker.CatalogResponse, error) {
	response := &broker.CatalogResponse{}

	services := make([]osb.Service, 0)
	for _, s := range b.services {
		services = append(services, s.Definition())
	}

	osbResponse := &osb.CatalogResponse{Services: services}
	response.CatalogResponse = *osbResponse

	return response, nil
}

func (b *BusinessLogic) Provision(request *osb.ProvisionRequest, c *broker.RequestContext) (*broker.ProvisionResponse, error) {
	requestedService := b.services[request.ServiceID]

	// Get the namespace to provision this resource in with a fallback to the
	// default service namespace
	namespace := b.namespace
	if request.Context["namespace"] != nil {
		namespace = request.Context["namespace"].(string)
	}

	spec := requestedService.GetProvisionSpec(service.ServiceOptions{
		ID:        request.InstanceID,
		PlanID:    request.PlanID,
		Namespace: namespace,
	})

	b.Lock()
	defer b.Unlock()

	response := broker.ProvisionResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
		go spec.Create(b.k8sClient)
	} else {
		spec.Create(b.k8sClient)
	}

	return &response, nil
}

func (b *BusinessLogic) Deprovision(request *osb.DeprovisionRequest, c *broker.RequestContext) (*broker.DeprovisionResponse, error) {
	requestedService := b.services[request.ServiceID]

	// Get the service instance resource from the cluster. This is done to test
	// if that instance exists and to get the namespace that the instance was
	// provisioned in
	list, _ := b.k8sClient.CoreV1().Secrets(v1.NamespaceAll).List(context.TODO(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("service-instance-id=%s", request.InstanceID),
	})

	// If there are no resources in the list with the requested service instance
	// id then just skip deprivation. This is because the resources have been
	// deleted by something else and there is nothing to deprivation
	if len(list.Items) == 0 {
		return &broker.DeprovisionResponse{}, nil
	}

	specOptions := service.ServiceOptions{
		ID:        request.InstanceID,
		PlanID:    request.PlanID,
		Namespace: list.Items[0].Namespace,
	}

	spec := requestedService.GetProvisionSpec(specOptions)
	deprovisionSpec := requestedService.GetDeprovisionSpec(specOptions)

	b.Lock()
	defer b.Unlock()

	response := broker.DeprovisionResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
		go func() {
			deprovisionSpec.Create(b.k8sClient)
			spec.Delete(b.k8sClient)
		}()
	} else {
		deprovisionSpec.Create(b.k8sClient)
		spec.Delete(b.k8sClient)
	}

	return &response, nil
}

func (b *BusinessLogic) LastOperation(request *osb.LastOperationRequest, c *broker.RequestContext) (*broker.LastOperationResponse, error) {
	// Your last-operation business logic goes here

	return nil, nil
}

func (b *BusinessLogic) Bind(request *osb.BindRequest, c *broker.RequestContext) (*broker.BindResponse, error) {
	requestedService := b.services[request.ServiceID]

	// Get the namespace to create this bindind in
	namespace := b.namespace
	if request.Context["namespace"] != nil {
		namespace = request.Context["namespace"].(string)
	}

	spec := requestedService.GetBindSpec(service.BindOptions{
		ID:         request.BindingID,
		InstanceID: request.InstanceID,
		Namespace:  namespace,
	})

	b.Lock()
	defer b.Unlock()

	response := broker.BindResponse{
		BindResponse: osb.BindResponse{
			Credentials: map[string]interface{}{},
		},
	}

	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	// Always throw the binding request into the background to ensure this
	// request dose not timeout
	go spec.Create(b.k8sClient)

	// Add all of the values from the bind spec into the response secret
	for k, v := range spec.Secrets[0].Data {
		response.BindResponse.Credentials[k] = string(v)
	}

	return &response, nil
}

func (b *BusinessLogic) Unbind(request *osb.UnbindRequest, c *broker.RequestContext) (*broker.UnbindResponse, error) {
	requestedService := b.services[request.ServiceID]
	namespace := b.namespace

	// Try to get the service id from another resource in the cluster if it has
	// not been passed in with the request. This is an optional paramiter and
	// can't guaranty it will be there
	if requestedService == nil {
		list, _ := b.k8sClient.CoreV1().Secrets(v1.NamespaceAll).List(context.TODO(), v1.ListOptions{
			LabelSelector: fmt.Sprintf("service-binding-id=%s", request.BindingID),
		})

		// If there is no resources with this service binding id then just
		// return. This is because the resources have been deleted by something
		// / someone else and the rest of the unbinding will fail because there
		// are no resources to delete
		if len(list.Items) == 0 {
			return &broker.UnbindResponse{}, nil
		}

		request.ServiceID = list.Items[0].Labels["service-id"]
		requestedService = b.services[request.ServiceID]
		namespace = list.Items[0].Namespace
	}

	if requestedService == nil {
		errorMessage := fmt.Sprintf("Invalid service '%s'", request.ServiceID)
		return &broker.UnbindResponse{}, osb.HTTPStatusCodeError{
			StatusCode:   400,
			ErrorMessage: &errorMessage,
		}
	}

	bindingOptions := service.BindOptions{
		ID:         request.BindingID,
		InstanceID: request.InstanceID,
		Namespace:  namespace,
	}
	bindSpec := requestedService.GetBindSpec(bindingOptions)
	debindSpec := requestedService.GetDebindSpec(bindingOptions)

	b.Lock()
	defer b.Unlock()

	debindSpec.Create(b.k8sClient)
	bindSpec.Delete(b.k8sClient)

	return &broker.UnbindResponse{}, nil
}

func (b *BusinessLogic) Update(request *osb.UpdateInstanceRequest, c *broker.RequestContext) (*broker.UpdateInstanceResponse, error) {
	// Your logic for updating a service goes here.
	response := broker.UpdateInstanceResponse{}
	if request.AcceptsIncomplete {
		response.Async = b.async
	}

	return &response, nil
}

func (b *BusinessLogic) ValidateBrokerAPIVersion(version string) error {
	return nil
}
