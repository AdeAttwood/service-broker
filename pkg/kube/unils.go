package kube

import (
	"math/rand"

	coreV1 "k8s.io/api/core/v1"
)

const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Helper function fore generating a kubernetes environment variable from a
// kubernetes secret
func EnvSecret(name string, secretName string, secretKey string) coreV1.EnvVar {
	return coreV1.EnvVar{
		Name: name,
		ValueFrom: &coreV1.EnvVarSource{
			SecretKeyRef: &coreV1.SecretKeySelector{
				LocalObjectReference: coreV1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}
