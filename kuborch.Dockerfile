FROM centos:6
LABEL maintainer="myntra"
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.14.0/bin/linux/amd64/kubectl \
    && chmod +x kubectl \
    && mv kubectl /usr/bin/kubectl \
    && mkdir -p yaml
ADD kuborch/kuborch /usr/bin/kuborch
RUN chmod +x /usr/bin/kuborch
ENTRYPOINT ["/usr/bin/kuborch", "-configPath=ci:/root/.kube/buildconfig", "-configPath=qa:/root/.kube/qaconfig", "-stderrthreshold=INFO"]
