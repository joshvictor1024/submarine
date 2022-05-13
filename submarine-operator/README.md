# submarine-operator

## Versions

- go 1.17.1
- kubernetes v1.21.2
- kubectl v1.21.2
- minikube v1.21.0
- operator-sdk v1.18.1
- istio v1.13.2

## Usage

```shell
minikube start --vm-driver=docker --cpus 8 --memory 8192 --kubernetes-version v1.21.2

istioctl install --set -y

kubectl create namespace submarine
kubectl create namespace submarine-user-test
kubectl label namespace submarine istio-injection=enabled
kubectl label namespace submarine-user-test istio-injection=enabled

go mod vendor
make install # installs crd
make run # runs operator in console
kubectl apply -f config/samples/submarine_v1alpha1_submarine.yaml -n submarine-user-test

kubectl delete submarine/submarine-sample -n submarine-user-test
make uninstall # removes crd

kubectl delete namespace submarine
kubectl delete namespace submarine-user-test

istioctl x uninstall --purge
```