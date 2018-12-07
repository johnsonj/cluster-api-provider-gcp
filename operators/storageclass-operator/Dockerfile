FROM ubuntu:latest
WORKDIR /root/
COPY manager manager
COPY channels/ channels/

ARG K8S_VER=v1.12.0
RUN apt update && apt -y install curl 
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$K8S_VER/bin/linux/amd64/kubectl && \
    chmod a+x kubectl && mv kubectl /usr/local/bin

ENTRYPOINT ["./manager"]
