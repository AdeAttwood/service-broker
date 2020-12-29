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

Installation with helm

```bash
helm repo add ibm https://charts.s3.eu-gb.cloud-object-storage.appdomain.cloud
kubectl create ns service-broker
helm install service-broker ibm/service-broker --namespace service-broker
```

If you have the service catalog cli installed you can verify the installation

```bash
svcat get classes
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

## Cloud Foundry for Kubernetes

Service broker supports [Cloud Foundry for
Kubernetes](https://github.com/cloudfoundry/cf-for-k8s) this allows you to bind
services to your Cloud Foundry deployments. For more info on getting cloud
foundry set up in your cluster you can see there [Getting Started
Guide](https://github.com/cloudfoundry/cf-for-k8s/blob/develop/docs/getting-started-tutorial.md).
To work with Cloud Foundry you will need to install with some variables set.

```bash
helm install service-broker ./charts/service-broker \
    --set tls.enabled=false \
    --set deployClusterServiceBroker=false \
    --namespace service-broker
```

**Link Cloud Foundry**

This will tell Cloud Foundry about Service Broker and list all of the available
services

```bash
cf create-service-broker service-broker user pass \
    http://service-broker-service-broker.service-broker.svc.cluster.local

cf service-access
```

**Enable Services**

You can enable selective services in your deployment with `enable-service-access`

```bash
cf enable-service-access mysql-instance
```

**Create a service instance**

Now all of the admin is done you can create a service instance. This creates a
`mysql-instance` with the `default` plan called `my-mysql-instance`

```bash
cf create-service mysql-instance default my-mysql-instance
```

**Service Binding**

Now we have a service instance we can bind it to a application. This example is
the `test-app` from the [Getting Started
Guide](https://github.com/cloudfoundry/cf-for-k8s/blob/develop/docs/getting-started-tutorial.md).
Once you restart you application the service cerdenals will be avlaidble in the
[VCAP-SERVICES](https://docs.cloudfoundry.org/devguide/deploy-apps/environment-variable.html#VCAP-SERVICES)
environment variable. You can also add a service binding in the [application
manifest](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#services-block)
but you will need to generate the service first as above

```bash
cf bind-service test-app my-mysql-instance
cf restage test-app
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
