###############################################################
####                           ARGS                        ####
###############################################################
# TARGETARCH is used to determine the architecture of the image, it is already set by Docker so it doesn't need to be defined here
ARG FAMILY=opensuse
ARG ELEMENTAL_TOOLKIT
ARG CLI_REPO
ARG MODELS_REPO
ARG VERSION
ARG TARGETARCH
ARG K3S_VERSION

###############################################################
####                        Base OS                        ####
###############################################################
# OS base image of our flavor choice
FROM registry.opensuse.org/opensuse/leap:15.5 AS baseos
ARG ARCH

RUN zypper ar -G https://download.opensuse.org/repositories/utilities/15.5/utilities.repo || true && \
    zypper ar -G https://nvidia.github.io/libnvidia-container/stable/rpm/nvidia-container-toolkit.repo || true && \
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
      pciutils \
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
      nvidia-container-toolkit \
      && zypper clean --all


FROM ${ELEMENTAL_TOOLKIT} AS TOOLKIT

FROM ${CLI_REPO}:${VERSION}-${TARGETARCH} AS LLMOS_CLI

FROM ${MODELS_REPO}:${VERSION} AS MODELS

###############################################################
####                        LLMOS                          ####
###############################################################
FROM baseos as llmos-opensuse
ARG REPO
ARG VERSION
ARG TARGETARCH
ARG K3S_VERSION
ARG FLAVOR
ARG REGISTRY
ENV LLMOS_REPO=${REPO}
ENV LLMOS_VERSION=${VERSION}
ENV K3S_VERSION=${K3S_VERSION}
ENV LLMOS_REGISTRY=${REGISTRY}
ENV LLMOS_FLAVOR=${FLAVOR}

# Add system files first
COPY iso/framework/files /

# Add LLMOS installer cli
COPY --from=LLMOS_CLI /llmos /usr/bin/llmos
RUN chmod +x /usr/bin/llmos

# Add the elemental cli
COPY --from=TOOLKIT /usr/bin/elemental /usr/bin/elemental

# Copy the ollama model
RUN mkdir -p /var/lib/llmos/ollama
COPY --from=MODELS /dist/models /var/lib/llmos/ollama

# Enable essential services
RUN systemctl enable NetworkManager.service

# Enable /tmp to be on tmpfs
RUN cp /usr/share/systemd/tmp.mount /etc/systemd/system

###############################################################
####               k3s + airgap assets                     ####
###############################################################
ENV INSTALL_K3S_BIN_DIR="/usr/bin"
ENV INSTALL_K3S_SELINUX_WARN=true
ENV INSTALL_K3S_SKIP_START="true"
ENV INSTALL_K3S_SKIP_ENABLE="true"
ENV INSTALL_K3S_SKIP_SELINUX_RPM="true"
ENV INSTALL_K3S_VERSION=${K3S_VERSION}

## Copy air-gap images
RUN mkdir -p /var/lib/rancher/k3s/agent/images
COPY dist/artifacts/*.tar.zst /var/lib/rancher/k3s/agent/images

RUN curl -sfL https://get.k3s.io > installer.sh
RUN bash installer.sh
RUN bash installer.sh agent
RUN rm -rf installer.sh
RUN chmod +x /usr/bin/k3s

# Save some space
RUN zypper clean --all && \
    rm -rf /var/log/update* && \
    >/var/log/lastlog

# Update os-release file with some metadata
RUN echo NAME=\"LLMOS-${LLMOS_FLAVOR}\" >> /etc/os-release && \
    echo VERSION=\"${LLMOS_VERSION}\"       >> /etc/os-release && \
    echo ID=\"LLMOS-${LLMOS_FLAVOR}\"   >> /etc/os-release && \
    echo IMAGE_REPO=\"https://github.com/llmos-ai/llmos\"       >> /etc/os-release && \
    echo IMAGE_TAG=\"${LLMOS_VERSION}\"     >> /etc/os-release && \
    echo IMAGE=\"${LLMOS_REPO}:${LLMOS_VERSION}\" >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"LLMOS-${LLMOS_FLAVOR}\" >> /etc/os-release

# Generate initrd with required elemental services
RUN elemental init elemental-rootfs,elemental-sysroot,grub-config,grub-default-bootargs,elemental-setup,dracut-config,boot-assessment --debug --force

RUN rm /etc/machine-id || true
RUN rm /var/lib/dbus/machine-id || true
RUN rm /etc/hostname || true

# Good for validation after the build
CMD /bin/bash
