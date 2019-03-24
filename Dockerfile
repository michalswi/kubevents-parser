# docker build -t local/kubevents:0.0.1 .
# docker run -it -p 5000:5000 local/kubevents:0.0.1
# curl localhost:5000/api/v1/log

FROM golang:1.12

# RUN mkdir -p /root/.kube
# COPY ./config /root/.kube/config

# FROM scratch
COPY kubevents .

ENTRYPOINT [ "./kubevents" ]