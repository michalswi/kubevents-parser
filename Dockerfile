# [local] 
# docker run -it -p 5000:5000 local/kubevents:0.0.1
FROM golang:1.15

# [local]
# RUN mkdir -p /root/.kube
# COPY ./config /root/.kube/config

COPY kubevents-parser .
ENTRYPOINT [ "./kubevents-parser" ]