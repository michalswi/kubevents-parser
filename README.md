### Simple web server to monitor K8s events 

You can run it either on k8s or locally.

#### Prerequisites
```
go build -a -ldflags '-w -s' -installsuffix cgo -o .
docker build -t local/kubevents:0.0.1 .
```

#### Run on kubernetes cluster
Displays events by default **only** from `default` namespace.  
```sh
$ kubectl apply -f deploy/rbac.yml
$ kubectl apply -f deploy/pod.yml
$ kubectl apply -f deploy/svc.yml
```
If other namespace than `default` it should be specify as INITNAMESPACE [here](./deploy/pod.yml). 
```sh
$ kubectl apply -f deploy/rbac.yml -n <initnamespace>
$ kubectl apply -f deploy/pod.yml -n <initnamespace>
$ kubectl apply -f deploy/svc.yml -n <initnamespace>
```
If you don't provide `-n <initnamespace>` you want have access to provided namespace, error in logs.  

#### Run locally
```sh
# default namespace
$ ./kubevents-parser --run-outside-k-cluster true
Kubevents 2021/07/16 09:44:39 kubevents.go:68: Starting server on port :5000

# any namespace
$ ./kubevents-parser --ns=mynamespace --run-outside-k-cluster true
```

#### Check locally
```sh
# web server considers only events which appeared after the script was run
$ curl localhost:5000/api/v1/log | jq
{
  "data": null,
  "error": "null",
  "namespace": "default",
  "status": "running"
}

# run some app
$ kubectl run hello-app --image=nginxdemos/hello --port=80

# check
$ curl localhost:5000/api/v1/log | jq
{
  "data": [
    {
      "id": 1,
      "name": "hello-app.15713668420d4728",
      "reason": "ScalingReplicaSet",
      "timeup": "00:00:43"
    },
    {
      "id": 2,
      "name": "hello-app-5c7477d7b7.15713668437e7582",
      "reason": "SuccessfulCreate",
      "timeup": "00:00:43"
    },
    ...
  ],
  "error": "null",
  "namespace": "default",
  "status": "running"
}

$ kubectl delete pod hello-app
```
