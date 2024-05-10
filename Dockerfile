ARG ARCH="amd64"
ARG OS="linux"
FROM ${ARCH}/debian:12
LABEL maintainer="Trey Dockendorf <treydock@gmail.com>"
ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/infiniband_exporter /infiniband_exporter
RUN apt update && apt install -y infiniband-diags && apt clean && rm -rf /var/lib/{apt,dpkg,cache,log}/
EXPOSE 9315
ENTRYPOINT ["/infiniband_exporter"]
