###############################################################
####                           ARGS                        ####
###############################################################
# TARGETARCH is used to determine the architecture of the image, it is already set by Docker so it doesn't need to be defined here
ARG FAMILY=opensuse
ARG ELEMENTAL_TOOLKIT
ARG LLMOS_CLI_REPO
ARG VERSION
ARG GOLANG_ARCH
ARG INSTALL_K3S_VERSION=v1.29.3+k3s1

#FROM alpine AS container-images
#ARG GOLANG_ARCH
#ARG INSTALL_K3S_VERSION
#RUN apk add --no-cache curl zstd
## prepare air-gap env images
#RUN mkdir -p /images/k3s
#RUN curl -sfL -o /images/k3s/k3s-airgap-images-amd64.tar.zst "https://github.com/k3s-io/k3s/releases/download/${INSTALL_K3S_VERSION}/k3s-airgap-images-${GOLANG_ARCH}.tar.zst"
#RUN ls -la /images/k3s

###############################################################
####                        Base OS                        ####
###############################################################
# OS base image of our flavor choice
FROM registry.opensuse.org/opensuse/leap:15.5 AS baseos
ARG ARCH

RUN zypper ar -G https://download.opensuse.org/repositories/utilities/15.5/utilities.repo || true && \
    zypper ref
# The image should provide at least grub, systemd, dracut, a kernel and an initrd
RUN ARCH=$(uname -m); \
    [[ "${ARCH}" == "aarch64" ]] && ARCH="arm64"; \
    zypper --non-interactive install --no-recommends -- \
      kernel-default \
      device-mapper \
      dracut \
      grub2 \
      grub2-${ARCH}-efi \
      bash-completion \
      btrfsprogs \
      btrfsmaintenance \
      conntrack-tools \
      coreutils \
      curl \
      dhcp-client \
      dosfstools \
      e2fsprogs \
      fail2ban \
      findutils \
      gawk \
      growpart \
      gptfdisk \
      gzip \
      haveged \
      htop \
      iproute2 \
      iptables \
      iputils \
      jq \
      logrotate \
      lsscsi \
      less \
      lvm2 \
      mtools \
      nano \
      nohang \
      NetworkManager \
      open-iscsi \
      openssh-server \
      openssh-clients \
      parted \
      rsync \
      shim \
      squashfs \
      strace \
      sudo \
      snapper \
      systemd \
      sed \
      timezone \
      tar \
      vim \
      which \
      wget2 \
      xorriso \
      zstd \
      && zypper clean --all


FROM ${ELEMENTAL_TOOLKIT} AS TOOLKIT

FROM ${LLMOS_CLI_REPO}:${VERSION}-${GOLANG_ARCH} AS LLMOS_CLI

FROM ollama/ollama:0.1.29 as ollama

###############################################################
####                        LLMOS                          ####
###############################################################
FROM baseos as llmos-opensuse
ARG REPO
ARG VERSION
ARG ARCH
ARG GOLANG_ARCH
ENV VERSION=${VERSION}
# Add system files first
COPY iso/framework/files /

# Add LLMOS installer cli
COPY --from=LLMOS_CLI /llmos /usr/bin/llmos
RUN chmod +x /usr/bin/llmos

# Add the elemental cli
COPY --from=TOOLKIT /usr/bin/elemental /usr/bin/elemental

# Add the ollama bin
COPY --from=ollama /usr/bin/ollama /usr/bin/ollama

# Copy the ollama model
RUN mkdir -p /var/lib/llmos/ollama
COPY build-models/models /var/lib/llmos/ollama

# Enable essential services
RUN systemctl enable NetworkManager.service

# Enable /tmp to be on tmpfs
RUN cp /usr/share/systemd/tmp.mount /etc/systemd/system


ENV INSTALL_K3S_BIN_DIR="/usr/bin"
ENV INSTALL_K3S_SELINUX_WARN=true
ENV INSTALL_K3S_SKIP_START="true"
ENV INSTALL_K3S_SKIP_ENABLE="true"
ENV INSTALL_K3S_SKIP_SELINUX_RPM="true"
ENV INSTALL_K3S_VERSION=v1.29.3+k3s1

RUN mkdir -p /var/lib/rancher/k3s/agent/images
RUN curl -sfL -o /var/lib/rancher/k3s/agent/images/k3s-airgap-images-amd64.tar.zst "https://github.com/k3s-io/k3s/releases/download/${INSTALL_K3S_VERSION}/k3s-airgap-images-${GOLANG_ARCH}.tar.zst"
COPY webui.tar /var/lib/rancher/k3s/agent/images/webui.tar

RUN curl -sfL https://get.k3s.io > installer.sh
RUN bash installer.sh
RUN bash installer.sh agent
RUN rm -rf installer.sh
RUN chmod +x /usr/bin/k3s

# Generate initrd with required elemental services
RUN elemental init immutable-rootfs,grub-config,grub-default-bootargs,elemental-setup,dracut-config --debug --force

# Update os-release file with some metadata
RUN echo NAME=\"LLMOS-openSUSE Leap\" >> /etc/os-release && \
    echo VERSION=\"${VERSION}\"       >> /etc/os-release && \
    echo ID=\"llmos-opensuse-leap\"   >> /etc/os-release && \
    echo IMAGE_REPO=\"${REPO}\"       >> /etc/os-release && \
    echo IMAGE_TAG=\"${VERSION}\"     >> /etc/os-release && \
    echo IMAGE=\"${REPO}:${VERSION}\" >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"LLMOS\" >> /etc/os-release

RUN rm /etc/machine-id || true
RUN rm /var/lib/dbus/machine-id || true
RUN rm /etc/hostname || true

# Good for validation after the build
CMD /bin/bash
