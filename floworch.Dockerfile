FROM centos:6
LABEL maintainer="myntra"
ENV ENABLE_METRICS=true
WORKDIR /usr/bin
ADD floworch/floworch /usr/bin/floworch
RUN chmod +x /usr/bin/floworch
ADD config.yaml /usr/bin/config.yaml
ENTRYPOINT ["/usr/bin/floworch", "-stderrthreshold=INFO"]
