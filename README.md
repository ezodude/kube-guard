# Kube Guard

Kube Guard is a POC k8s API. For now, it enumerates a subject's privileges e.g. roles.

To run the API, please clone the repo to the correct location in the `GOPATH`.

Once cloned, please navigate to the project's root directory.

## Finding a subject's privileges (roles)

- You can search for more than one `subject` at once.

- You can retrieve results as `JSON` or `YAML`.

- You can use a RegExp to enable a wild card search.

- All listed privileges are ordered by `subject` alphabetically.

## Minikube

Kube Guard assumes you have `MiniKube` installed locally and bootstrapped with users and roles.

Follow the steps below to get up and running with a bootstrapped `Minikube`.

## Minikube Installation

If Minikube is not installed, then please use [instructions here to install it](https://kubernetes.io/docs/tasks/tools/install-minikube/).

## RBAC, setup a user

We're looking to create a `developer` user bound to a `pod-reader` role.

This user can only `["get", "watch", "list"]` the `pods` resource.

### Create user's client certificate

Create a directory where to save the certificates

```shell
mkdir cert && cd cert
```

Generate a key using `OpenSSL`

```shell
openssl genrsa -out developer.key 2048
```

Generate a Client Sign Request (CSR)

```shell
openssl req -new \
  -key developer.key \
  -out developer.csr \
  -subj "/CN=developer/O=group1"
```

Generate the certificate (CRT)

```shell
openssl x509 -req \
  -in developer.csr \
  -CA ~/.minikube/ca.crt \
  -CAkey ~/.minikube/ca.key \
  -CAcreateserial \
  -out developer.crt \
  -days 500
```

### Create developer user

Set a user entry in kubeconfig

```shell
kubectl config set-credentials developer \
  --client-certificate=developer.crt \
  --client-key=developer.key
```

Set a context entry in kubeconfig

```shell
kubectl config set-context developer-context \
  --cluster=minikube \
  --namespace=default \
  --user=developer
```

You can check that it is successfully added to kubeconfig:

```shell
kubectl config view
```

## RBAC, grant a role to the user

### Deploy both role.yaml and role-binding.yaml to k8s

Please use the provided files `role.yaml` and `role-binding.yaml`.

- `role.yaml`, creates a `pod-reader` role.
- `role-binding.yaml`, creates a role binding between our `developer` user and the `pod-reader` role.

Go back to the project dir.

```shell
cd ..
```

Ensure you're using the `minikube` context

```shell
kubectl config use-context minikube
```

Apply the role

```shell
kubectl apply -f role.yaml
```

Apply the role binding

```shell
kubectl apply -f role-binding.yaml
```

### Check deployed roles and role binding

```shell
kubectl get roles
kubectl get rolebindings
```

## Up and running with Kube Guard

First clone the repo to the correct location in the `GOPATH`.

Once cloned, please navigate to the project's root directory.

### Running the API server

```shell
cd cmd/api
go build
./api
```

### Sample Requests

Retrieving data as `JSON`

```shell
curl -XGET http://localhost:8080/api/v0.1/privilege/search \
  -d '{"subjects":["developer"],"format":"JSON"}' \
  -H 'Content-Type:application/Shutting down the API server
```

Retrieving data as `YAML`

```shell
curl -XGET http://localhost:8080/api/v0.1/privilege/search \
  -d '{"subjects":["developer"],"format":"YAML"}' \
  -H 'Content-Type:application/json'
```

Using RegExp wildcards

```shell
curl -XGET http://localhost:8080/api/v0.1/privilege/search \
  -d '{"subjects":["developer", "deve*"],"format":"JSON"}' \
  -H 'Content-Type:application/json'
```

### Shutting down the API server

Simply `CTRL-C` on the terminal window where the server is running.
