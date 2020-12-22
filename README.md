# Kubernetes Service Broker

[Open service broker implementation](https://github.com/openservicebrokerapi/servicebroker) for kubernetes, bootstrapped from the
[pmorie/osb-starter-pack](https://github.com/pmorie/osb-starter-pack) and
inspired by
[kubernetes-sigs/minibroker](https://github.com/kubernetes-sigs/minibroker).

Available Service:

- **MySql Instance:** A mysql instance deployed into your namespace

## Who should use this project?

You should use this project if you're looking for a quick way to implement an
Open Service Broker and start iterating on it.

## Prerequisites

You'll need:

- A running [Kubernetes](https://github.com/kubernetes/kubernetes) cluster
- [Helm 3](https://helm.sh)
- The [service-catalog](https://github.com/kubernetes-incubator/service-catalog)
  [installed](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md)
  in that cluster

```bash
# Installing the service catalog
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
kubectl create ns catalog
helm install catalog svc-cat/catalog --namespace catalog --set asyncBindingOperationsEnabled=true
```

## Installation

**TODO: Better installation of cart. This is installing it from the repo**

```bash
kubectl create ns service-broker
helm install service-broker --namespace service-broker charts/service-broker
```

## Getting started

**Create a test namespace**

```bash
kubectl create ns test-ns
```

**Provision a mysql service instance**

This will spin up a mysql instance in the `test-ns`

```bash
kubectl apply -f manifests/mysql/instance-service-instance.yaml
```

**Create a binding to the service**

This will create a kubernetes job to create a new user on the instance and
generate the corresponding secrets. This way each application get their own
database user. An example of useing secrets in a deployment can be found in the
[Wordpress example](manifests/mysql/wordpress.yml)

```bash
kubectl apply -f manifests/mysql/instance-service-binding.yaml
```

Access the binding secret

```bash
echo User: $(kubectl get secret mysql-instance-service-binding -n test-ns -o jsonpath="{.data.user}" | base64 --decode)
echo Password: $(kubectl get secret mysql-instance-service-binding -n test-ns -o jsonpath="{.data.password}" | base64 --decode)
echo Database: $(kubectl get secret mysql-instance-service-binding -n test-ns -o jsonpath="{.data.database}" | base64 --decode)
echo Host: $(kubectl get secret mysql-instance-service-binding -n test-ns -o jsonpath="{.data.host}" | base64 --decode)
```

Remove the service binding

```bash
kubectl delete -f manifests/mysql-instance-service-binding.yaml
```

Remove the service instance

```bash
kubectl delete -f manifests/mysql-instance-service-instance.yaml
```

## TODO

- [ ] Allow parameter to the mysql binding to allow there to be multiple
  databases on the instance
- [ ] Allow to provision different size instance probably through a plan
- [ ] Remove user when binding is deleted
- [ ] Add more services
  - [ ] Minio
  - [ ] Shared MySql

## Known Issues

- [ ] Things get a bit stuck if you try and remove the namespace and not delete
  the manifests the services were provisioned with
- [ ] Pods that run the binding jobs dont get deleted
