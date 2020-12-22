package service

import (
	"github.com/AdeAttwood/service-broker/pkg/kube"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
)

type Service interface {
	Definition() osb.Service
	GetHost(instanceId string, namespace string) string
	GetProvisionSpec(options ServiceOptions) *kube.Spec
	GetDeprovisionSpec(options ServiceOptions) *kube.Spec
	GetBindSpec(options BindOptions) *kube.Spec
	GetDebindSpec(options BindOptions) *kube.Spec
}

type ServiceOptions struct {
	ID        string
	PlanID    string
	Namespace string
}

type BindOptions struct {
	ID         string
	InstanceID string
	Namespace  string
}

func int32Ptr(i int32) *int32 { return &i }
func int64Ptr(i int64) *int64 { return &i }

func truePtr() *bool {
	b := true
	return &b
}
