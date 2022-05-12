# submarine-operator

## Versions

- go 1.17.1
- kubernetes v1.21.2
- kubectl v1.21.2
- minikube v1.21.0
- operator-sdk v1.18.1
- istio v1.13.1

## Usage

```shell
minikube start --vm-driver=docker --cpus 8 --memory 4096 --kubernetes-version v1.21.2
go mod vendor
make install # installs crd
make run # runs operator in console
make uninstall # removes crd
```