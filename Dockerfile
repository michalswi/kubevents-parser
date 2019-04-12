# [local] docker run -it -p 5000:5000 local/kubevents:0.0.1
FROM golang:1.12

# [local]
# RUN mkdir -p /root/.kube
# COPY ./config /root/.kube/config

COPY kubevents .

ENTRYPOINT [ "./kubevents" ]