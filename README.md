## Prerequisites
Api server should be runned with the RBAC API group. To enable RBAC, start the api server with `--authorization-mode=RBAC` or `—extra-config=apiserver.Authorization.Mode=RBAC` in case of minikube.
## How to run
Application uses default kubeconfig to connect to k8s cluster. Default config is located at `~/.kube/config`. Add `-kubeconfig=<path to config>` to launch command if you want to use custom config file. 
Also you could use `make` to build and run application with default configuration:
* `make build` - download dependencies and build application as `kuber` executable
* `make run` - run `kuber` with default configuration (default `kubeconfig` and `bootstrap` files)


## bootstrap.json
Application uses json to prepare k8s cluster for tests. You can see example json with test data in `bootstrap.json`.
Bootstrap file consists of several base types:
* namespaces  
* users
* roles
* cluster roles
* role bindings
* cluster role bindings

All types are very close to the k8s API analogues. User type also contains additional array `tests` which contains authorization tests for k8s api. This paths are executing after resources initialization.
Each test contains path and expected http status code.     