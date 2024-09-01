ARG ARCH="amd64"
ARG OS="linux"
FROM ${ARCH}/debian:12
LABEL maintainer="Trey Dockendorf <treydock@gmail.com>"
ARG ARCH="amd64"
ARG OS="linux"
COPY .build/${OS}-${ARCH}/infiniband_exporter /infiniband_exporter
RUN apt update && \
    apt install -y infiniband-diags && \
    apt clean && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
RUN apt update && \
    apt install -y curl && \
    curl -L -O https://www.mellanox.com/downloads/MFT/mft-4.29.0-131-x86_64-deb.tgz && \
    tar xf mft-4.29.0-131-x86_64-deb.tgz && \
    ./mft-4.29.0-131-x86_64-deb/install.sh --without-kernel --without-autocomplete || cat /tmp/mft*/*.log && \
    rm -rf mft-4.29.0-131-x86_64-deb* && \
    curl -o /usr/bin/ibswinfo https://raw.githubusercontent.com/stanford-rc/ibswinfo/main/ibswinfo.sh && \
    chmod +x /usr/bin/ibswinfo && \
    apt purge -y curl && \
    apt autoremove -y && \
    apt clean && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
EXPOSE 9315
ENTRYPOINT ["/infiniband_exporter"]
