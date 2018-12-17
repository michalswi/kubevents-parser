
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
# it will load and display already existing events from 'default' namespace
$ go run kubevents.go
Start..
2018/12/17 19:44:49 Event added, name: hello-app-5c7477d7b7-94brw.1571326091adb1c9, reason: Scheduled, timestamp: 2018-12-17 19:32:17 +0100 CET
2018/12/17 19:44:49 Event added, name: hello-app-5c7477d7b7-94brw.15713260a3891339, reason: SuccessfulMountVolume, timestamp: 2018-12-17 19:32:17 +0100 CET
```

**Check**:
```sh
# "timeup": "00:00:00" - set if event alread exists, each new will have the proper time when up
$ curl localhost:5000/api/v1/log | jq
{
  "data": [
    {
      "id": 1,
      "name": "hello-app-5c7477d7b7-94brw.1571326091adb1c9",
      "reason": "Scheduled",
      "timeup": "00:00:00"
    },
    {
      "id": 2,
      "name": "hello-app-5c7477d7b7-94brw.15713260a3891339",
      "reason": "SuccessfulMountVolume",
      "timeup": "00:00:00"
    }
  ],
  "error": "null",
  "status": "running"
}
```