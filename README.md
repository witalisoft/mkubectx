# mkubectx

mkubectx stands as an acronym for multiple kubernetes contexts. Tool allow you to pass command to all or selected kubernetes contexts. To achieve this it's dealing with setup environment variable `KUBECONFIG` for selected contexts and execute command with it. Exec command in each kubernetes context is done in parallel fashion. By default all contexts are taken. 

## Build from source
Project is using Go modules, just build:
```
$ go build
```

## Usage
```
mkubectx [-contexts ctx1,ctx2,...] command [args...]
  contexts arguments can be passed as regular expression
```

Defined command has access to environment variable `KUBECTX` with name of current kubernetes context.

## Examples
```
./mkubectx -contexts "gke_.*(pro|ci).*,minikube" kubectl get namespaces
gke_cluster_pro-0
  NAME              STATUS   AGE
  default           Active   55d
  kube-node-lease   Active   55d
  kube-public       Active   55d
  kube-system       Active   55d
  monitoring        Active   55d
  velero            Active   54d
gke_cluster_pro-1
  NAME              STATUS   AGE
  default           Active   188d
  kube-node-lease   Active   47d
  kube-public       Active   188d
  kube-system       Active   188d
  monitoring        Active   188d
  velero            Active   61d
gke_cluster_ci-0
  NAME              STATUS   AGE
  default           Active   376d
  kube-node-lease   Active   47d
  kube-public       Active   376d
  kube-system       Active   376d
  velero            Active   54d
minikube
  Unable to connect to the server: dial tcp 192.168.64.2:8443: i/o timeout
 error:
  error in executing command, err: exit status 1
```

## Issues
* no Windows support

