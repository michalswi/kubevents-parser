
Simple webserver to monitor k8s events.

**Prerequisites**:  
```sh
$ go get github.com/gorilla/mux
$ go get k8s.io/client-go/...
$ go get k8s.io/api/...
$ go get k8s.io/apimachinery/...
```

**Run**:  
```sh
# it displays already existing events from 'default' namespace
$ go run kubevents.go
Start..
2018/12/17 19:44:49 Event added, name: hello-app-5c7477d7b7-94brw.1571326091adb1c9, reason: Scheduled, timestamp: 2018-12-17 19:32:17 +0100 CET
2018/12/17 19:44:49 Event added, name: hello-app-5c7477d7b7-94brw.15713260a3891339, reason: SuccessfulMountVolume, timestamp: 2018-12-17 19:32:17 +0100 CET
```

**Check**:
```sh
# webserver considers only events which appeared after the script was run
$ curl localhost:5000/api/v1/log | jq
{
  "data": null,
  "error": "null",
  "status": "running"
}

# run some app
$ kubectl run hello-app --image=nginxdemos/hello --port=80 --replicas=1

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
  "status": "running"
}
```