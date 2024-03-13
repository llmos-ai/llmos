ARG ELEMENTAL_TOOLKIT
ARG LLMOS_CLI_REPO
ARG VERSION
ARG GOLANG_ARCH
FROM ${ELEMENTAL_TOOLKIT} AS TOOLKIT

FROM ${LLMOS_CLI_REPO}:${VERSION}-${GOLANG_ARCH} AS LLMOS_CLI

# OS base image of our flavor choice
FROM registry.opensuse.org/opensuse/leap:15.5 as OS
ARG REPO
ARG VERSION
ENV VERSION=${VERSION}

# install kernel, systemd, dracut, grub2 and other required tools
RUN ARCH=$(uname -m); \
    [[ "${ARCH}" == "aarch64" ]] && ARCH="arm64"; \
    zypper --non-interactive install --no-recommends -- \
      kernel-default \
      device-mapper \
      dracut \
      grub2 \
      grub2-${ARCH}-efi \
      shim \
      haveged \
      systemd \
      NetworkManager \
      openssh-server \
      openssh-clients \
      timezone \
      parted \
      e2fsprogs \
      dosfstools \
      mtools \
      xorriso \
      findutils \
      gptfdisk \
      rsync \
      squashfs \
      lvm2 \
      tar \
      gzip \
      vim \
      which \
      less \
      sudo \
      curl \
      iproute2 \
      podman \
      sed \
      btrfsprogs \
      btrfsmaintenance \
      snapper && \
    zypper clean --all

# Add system files first
COPY iso/framework/files /
#COPY iso/framework/boot /

# Add LLMOS installer cli
COPY --from=LLMOS_CLI /llmos /usr/sbin/llmos
RUN chmod +x /usr/sbin/llmos

# Just add the elemental cli
COPY --from=TOOLKIT /usr/bin/elemental /usr/bin/elemental

# Enable essential services
RUN systemctl enable NetworkManager.service

# Enable /tmp to be on tmpfs
RUN cp /usr/share/systemd/tmp.mount /etc/systemd/system

# Generate initrd with required elemental services
RUN elemental init --debug --force

# Update os-release file with some metadata
RUN echo IMAGE_REPO=\"${REPO}\"         >> /etc/os-release && \
    echo IMAGE_TAG=\"${VERSION}\"           >> /etc/os-release && \
    echo IMAGE=\"${REPO}:${VERSION}\" >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"LLMOS\" >> /etc/os-release

# Good for validation after the build
CMD /bin/bash
