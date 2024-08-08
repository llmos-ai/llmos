#!/bin/sh
set -e
set -o noglob

# Usage:
#   curl ... | ENV_VAR=... sh - [args]
#       or
#   ENV_VAR=... ./install.sh [args]
#
# Example:
#   Installing a cluster init-role with bootstrap password:
#     curl ... | sh -s - --cluster-init --bootstrap-password=xxx
#   Installing an agent to join the cluster:
#     curl ... | LLMOS_TOKEN=xxx LLMOS_URL=https://server-url:6443 sh -
#
#
# Environment variables:
#   - LLMOS_*
#     Environment variables which begin with LLMOS_ will be preserved for the
#     systemd service to use. Setting LLMOS_URL without explicitly setting
#     a systemd exec command will default the command to "agent", and we
#     enforce that LLMOS_TOKEN is also set.
#
#   - INSTALL_LLMOS_SKIP_DOWNLOAD
#     If set to true will not download llmos hash or binary.
#
#   - INSTALL_LLMOS_FORCE_RESTART
#     If set to true will always restart the llmos service
#
#   - INSTALL_LLMOS_SKIP_ENABLE
#     If set to true will not enable or start llmos service.
#
#   - INSTALL_LLMOS_SKIP_START
#     If set to true will not start llmos service.
#
#   - INSTALL_LLMOS_VERSION
#     Version of llmos to download from github. Will attempt to download from the
#     stable channel if not specified.
#
#   - INSTALL_LLMOS_BIN_DIR
#     Directory to install llmos binary, links, and uninstall script to, or use
#     /usr/local/bin as the default
#
#   - INSTALL_LLMOS_SYSTEMD_DIR
#     Directory to install systemd service and environment files to, or use
#     /etc/systemd/system as the default
#
LLMOS_RELEASE_URL=https://github.com/llmos-ai/llmos/releases
DOWNLOADER=

# --- helper functions for logs ---
info()
{
	echo '[INFO] ' "$@"
}
warn()
{
	echo '[WARN] ' "$@" >&2
}
fatal()
{
	echo '[ERROR] ' "$@" >&2
	exit 1
}

# --- fatal if no systemd ---
verify_system() {
	if [ -x /bin/systemctl ] || type systemctl > /dev/null 2>&1; then
		HAS_SYSTEMD=true
		return
	fi
	if [ -x /sbin/openrc-run ]; then
		HAS_OPENRC=true
		return
	fi

	fatal 'Can not find systemd or openrc to use as a process supervisor for LLMOS'
}

# --- add quotes to command arguments ---
quote() {
    for arg in "$@"; do
		printf '%s\n' "$arg" | sed "s/'/'\\\\''/g;1s/^/'/;\$s/\$/'/"
    done
}

# --- add indentation and trailing slash to quoted args ---
quote_indent() {
    printf ' \\\n'
    for arg in "$@"; do
        printf '\t%s \\\n' "$(quote "$arg")"
    done
}

# --- escape most punctuation characters, except quotes, forward slash, and space ---
escape() {
    printf '%s' "$@" | sed -e 's/\([][!#$%&()*;<=>?\_`{|}]\)/\\\1/g;'
}

# --- escape double quotes ---
escape_dq() {
    printf '%s' "$@" | sed -e 's/"/\\"/g'
}

# --- use sudo if we are not already root ---
check_sudo() {
    if [ $(id -u) -ne 0 ]; then
		if command -v sudo >/dev/null 2>&1; then
			info "running as non-root, will use sudo for installation."
			SUDO=sudo
	  	else
			fatal "This script must be run as root. Please use sudo or run as root."
		fi
    else
		SUDO=
    fi
}

# --- define needed environment variables ---
setup_env() {
    SYSTEM_NAME=llmos

    CMD_LLMOS_EXEC="$(quote_indent "$@")"

    # --- check for invalid characters in system name ---
    valid_chars=$(printf '%s' "${SYSTEM_NAME}" | sed -e 's/[][!#$%&()*;<=>?\_`{|}/[:space:]]/^/g;' )
    if [ "${SYSTEM_NAME}" != "${valid_chars}"  ]; then
        invalid_chars=$(printf '%s' "${valid_chars}" | sed -e 's/[^^]/ /g')
        fatal "Invalid characters for system name:
            ${SYSTEM_NAME}
            ${invalid_chars}"
    fi

    # --- use binary install directory if defined or create default ---
    if [ -n "${INSTALL_LLMOS_BIN_DIR}" ]; then
        BIN_DIR=${INSTALL_LLMOS_BIN_DIR}
    else
        # --- use /usr/local/bin if root can write to it, otherwise use /opt/bin if it exists
        BIN_DIR=/usr/local/bin
        if ! $SUDO sh -c "touch ${BIN_DIR}/llmos-ro-test && rm -rf ${BIN_DIR}/llmos-ro-test"; then
            if [ -d /opt/bin ]; then
                BIN_DIR=/opt/bin
            fi
        fi
    fi

    # --- use systemd directory if defined or create default ---
    if [ -n "${INSTALL_LLMOS_SYSTEMD_DIR}" ]; then
        SYSTEMD_DIR="${INSTALL_LLMOS_SYSTEMD_DIR}"
    else
        SYSTEMD_DIR=/etc/systemd/system
    fi

    # --- set related files from system name ---
    SERVICE_LLMOS=${SYSTEM_NAME}.service
    UNINSTALL_LLMOS_SH=${UNINSTALL_LLMOS_SH:-${BIN_DIR}/${SYSTEM_NAME}-uninstall.sh}

    # --- use service or environment location depending on systemd/openrc ---
	if [ "${HAS_SYSTEMD}" = true ]; then
		FILE_LLMOS_SERVICE=${SYSTEMD_DIR}/${SERVICE_LLMOS}
		FILE_LLMOS_ENV=${SYSTEMD_DIR}/${SERVICE_LLMOS}.env
    elif [ "${HAS_OPENRC}" = true ]; then
		$SUDO mkdir -p /etc/llmos
		FILE_LLMOS_SERVICE=/etc/init.d/${SYSTEM_NAME}
		FILE_LLMOS_ENV=/etc/llmos/${SYSTEM_NAME}.env
    fi

    # --- get hash of config & exec for currently installed llmos ---
    PRE_INSTALL_HASHES=$(get_installed_hashes)
}

# --- check if skip download environment variable set ---
can_skip_download() {
    if [ "${INSTALL_LLMOS_SKIP_DOWNLOAD}" != true ]; then
        return 1
    fi
}

# --- verify an executable llmos binary is installed ---
verify_llmos_is_executable() {
    if [ ! -x "${BIN_DIR}"/llmos ]; then
        fatal "Executable llmos binary not found at ${BIN_DIR}/llmos"
    fi
}

# --- set arch and suffix, fatal if architecture not supported ---
setup_verify_os_arch() {
    if [ -z "$OS_NAME" ]; then
        OS_NAME=$(uname -s)
    fi
    case $OS_NAME in
        Darwin)
            OS_NAME=darwin
            SUFFIX=_${OS_NAME}
            ;;
        Linux)
            OS_NAME=linux
            SUFFIX=_${OS_NAME}
            ;;
        *)
            fatal "Unsupported OS $OS_NAME"
    esac

    if [ -z "$ARCH" ]; then
        ARCH=$(uname -m)
    fi
    case $ARCH in
        amd64)
            ARCH=amd64
            SUFFIX="${SUFFIX}_${ARCH}"
            ;;
        x86_64)
            ARCH=amd64
            SUFFIX="${SUFFIX}_amd64"
            ;;
        arm64)
            ARCH=arm64
            SUFFIX="${SUFFIX}_arm64"
            ;;
        aarch64)
            ARCH=arm64
            SUFFIX="${SUFFIX}_${ARCH}"
            ;;
        *)
            fatal "Unsupported architecture $ARCH"
    esac

    info "Detected OS:${OS_NAME} ARCH:${ARCH}"
}

# --- verify existence of network utility command ---
verify_command() {
    # Return failure if it doesn't exist or is no executable
    [ -x "$(command -v "$1")" ] || return 1
	DOWNLOADER=$1
    return 0
}

# --- create temporary directory and cleanup when done ---
setup_tmp() {
    TMP_DIR=$(mktemp -d -t llmos-install.XXXXXXXXXX)
    TMP_HASH=${TMP_DIR}/llmos.hash
    TMP_BIN=${TMP_DIR}/llmos.bin
    cleanup() {
        code=$?
        set +e
        trap - EXIT
        rm -rf "${TMP_DIR}"
        exit $code
    }
    trap cleanup INT EXIT
}

# --- use desired llmos version if defined or find version from github releases ---
get_release_version() {
    if [ -n "${INSTALL_LLMOS_VERSION}" ]; then
		VERSION_LLMOS=${INSTALL_LLMOS_VERSION}
    else
		version_url="${LLMOS_RELEASE_URL}/latest"
		info "Finding release for LLMOS via ${version_url}"
		if [ "$DOWNLOADER" = "curl" ]; then
			VERSION_LLMOS=$(curl -Ls -o /dev/null -w %{url_effective} "${version_url}" | sed 's|.*/||')
		elif [ "$DOWNLOADER" = "wget" ]; then
			VERSION_LLMOS=$(wget --server-response -O /dev/null "${version_url}" 2>&1 | awk '/^  Location: / {print $2}' | sed 's|.*/||')
		else
			fatal "Incorrect downloader executable '$DOWNLOADER'"
		fi
    fi

	if expr "$VERSION_LLMOS" : '^v' >/dev/null; then
		info "Using ${VERSION_LLMOS} as release"
	else
		fatal "Invalid llmos version: ${VERSION_LLMOS}"
	fi
}

# --- download from github url ---
download() {
    [ $# -eq 2 ] || fatal 'download needs exactly 2 arguments'

    case $DOWNLOADER in
        curl)
            curl -o "$1" -sfL "$2"
            ;;
        wget)
            wget -qO "$1" "$2"
            ;;
        *)
            fatal "Incorrect executable '$DOWNLOADER'"
            ;;
    esac

    # Abort if download command failed
    [ $? -eq 0 ] || fatal 'Download failed'
}

# --- download checksums file from github url ---
download_checksums() {
    CHECKSUM_URL=${LLMOS_RELEASE_URL}/download/${VERSION_LLMOS}/checksums.txt
    info "Downloading checksums file ${CHECKSUM_URL}"
    download "${TMP_HASH}" "${CHECKSUM_URL}"
    HASH_EXPECTED=$(grep " llmos${SUFFIX}$" "${TMP_HASH}" | awk '{print $1}')
    HASH_EXPECTED=${HASH_EXPECTED%%[[:blank:]]*}
    info "Expected hash ${HASH_EXPECTED}"
}

# --- check hash against installed version ---
installed_hash_matches() {
    if [ -x ${BIN_DIR}/llmos ]; then
        HASH_INSTALLED=$(sha256sum ${BIN_DIR}/llmos)
        HASH_INSTALLED=${HASH_INSTALLED%%[[:blank:]]*}
        if [ "${HASH_EXPECTED}" = "${HASH_INSTALLED}" ]; then
            return
        fi
    fi
    return 1
}

# --- download binary from git ---
download_binary() {
    BIN_URL=${LLMOS_RELEASE_URL}/download/${VERSION_LLMOS}/llmos${SUFFIX}
    info "Downloading binary ${BIN_URL}"
    download ${TMP_BIN} ${BIN_URL}
}

# --- verify downloaded binary hash ---
verify_binary() {
    info "Verifying binary download"
    HASH_BIN=$(sha256sum ${TMP_BIN})
    HASH_BIN=${HASH_BIN%%[[:blank:]]*}
    if [ "${HASH_EXPECTED}" != "${HASH_BIN}" ]; then
        fatal "Download sha256 does not match ${HASH_EXPECTED}, got ${HASH_BIN}"
    fi
}

# --- setup permissions and move binary to system directory ---
setup_binary() {
    chmod 755 ${TMP_BIN}
    info "Installing llmos to ${BIN_DIR}/llmos"
    $SUDO chown root:root ${TMP_BIN}
    $SUDO mv -f ${TMP_BIN} ${BIN_DIR}/llmos
}

# --- download and verify llmos ---
download_and_verify() {
    if can_skip_download; then
       info 'Skipping llmos download and verify'
       verify_llmos_is_executable
       return
    fi

    setup_verify_os_arch
    verify_command curl || verify_command wget || fatal 'Can not find curl or wget for downloading files'
    setup_tmp
    get_release_version
    download_checksums

    if installed_hash_matches; then
        info 'Skipping binary downloaded, installed llmos matches hash'
        return
    fi

    download_binary
    verify_binary
    setup_binary
}

# --- create uninstall script ---
create_uninstall() {
    info "Creating uninstall script ${UNINSTALL_LLMOS_SH}"
    $SUDO tee ${UNINSTALL_LLMOS_SH} >/dev/null << EOF
#!/bin/sh
set -x
[ \$(id -u) -eq 0 ] || exec sudo \$0 \$@

LLMOS_DATA_DIR=\${LLMOS_DATA_DIR:-/var/lib/llmos}
KUBE_UNINSTALL=

if command -v systemctl; then
    systemctl disable ${SYSTEM_NAME}
    systemctl reset-failed ${SYSTEM_NAME}
    systemctl daemon-reload
fi
if command -v rc-update; then
    rc-update delete ${SYSTEM_NAME} default
fi

rm -f ${FILE_LLMOS_SERVICE}
rm -f ${FILE_LLMOS_ENV}

if [ -f /etc/systemd/system/k3s.service ]; then
	KUBE_UNINSTALL=k3s-uninstall.sh
elif [ -f /etc/systemd/system/k3s-agent.service ]; then
	KUBE_UNINSTALL=k3s-agent-uninstall.sh
elif [ -f /etc/systemd/system/rke2.service ]; then
	KUBE_UNINSTALL=rke2-uninstall.sh
elif [ -f /etc/systemd/system/rke2-agent.service ]; then
	KUBE_UNINSTALL=rke2-agent-uninstall.sh
else
	warn "Kubernetes runtime not found, skipping uninstall k8s runtime."
fi

if [ -n "${KUBE_UNINSTALL}" ]; then
	info "Uninstalling k8s runtime by ${KUBE_UNINSTALL}"
	$SUDO ${KUBE_UNINSTALL}

	$SUDO rm -rf /var/lib/rancher /etc/rancher
fi

remove_uninstall() {
    rm -f ${UNINSTALL_LLMOS_SH}
}
trap remove_uninstall EXIT

rm -rf \${LLMOS_DATA_DIR} /var/lib/rook/*llmos*
rm -f ${BIN_DIR}/llmos

EOF
    $SUDO chmod 755 ${UNINSTALL_LLMOS_SH}
    $SUDO chown root:root ${UNINSTALL_LLMOS_SH}
}

# --- disable current service if loaded --
systemd_disable() {
    $SUDO systemctl disable ${SYSTEM_NAME} >/dev/null 2>&1 || true
    $SUDO rm -f /etc/systemd/system/${SERVICE_LLMOS} || true
    $SUDO rm -f /etc/systemd/system/${SERVICE_LLMOS}.env || true
}

# --- capture current env and create file containing LLMOS_ variables ---
create_env_file() {
    info "env: Creating environment file ${FILE_LLMOS_ENV}"
    $SUDO touch "${FILE_LLMOS_ENV}"
    $SUDO chmod 0600 "${FILE_LLMOS_ENV}"
    env | grep '^LLMOS_' | $SUDO tee "${FILE_LLMOS_ENV}" >/dev/null
    env | grep -Ei '^(NO|HTTP|HTTPS)_PROXY' | $SUDO tee -a "${FILE_LLMOS_ENV}" >/dev/null
}

# --- write systemd or openrc service file ---
create_service_file() {
    [ "${HAS_SYSTEMD}" = true ] && create_systemd_service_file
    [ "${HAS_OPENRC}" = true ] && create_openrc_service_file
    return 0
}

# --- write systemd service file ---
create_systemd_service_file() {
    info "systemd: Creating service file ${FILE_LLMOS_SERVICE}"
    $SUDO tee ${FILE_LLMOS_SERVICE} >/dev/null << EOF
[Unit]
Description=LLMOS Bootstrap
Documentation=https://github.com/llmos-ai/llmos
Wants=network-online.target
After=network-online.target

[Install]
WantedBy=multi-user.target

[Service]
Type=oneshot
EnvironmentFile=-/etc/default/%N
EnvironmentFile=-/etc/sysconfig/%N
EnvironmentFile=-${FILE_LLMOS_ENV}
KillMode=process
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
TimeoutStartSec=0
ExecStart=${BIN_DIR}/llmos bootstrap ${CMD_LLMOS_EXEC}
EOF
}

# --- write openrc service file ---
create_openrc_service_file() {
    LOG_FILE=/var/log/${SYSTEM_NAME}.log

    info "openrc: Creating service file ${FILE_LLMOS_SERVICE}"
    $SUDO tee "${FILE_LLMOS_SERVICE}" >/dev/null << EOF
#!/sbin/openrc-run

depend() {
    after network-online
    want cgroups
}

start_pre() {
    rm -f /tmp/llmos.*
}

supervisor=supervise-daemon
name=${SYSTEM_NAME}
command="${BIN_DIR}/llmos"
command_args="$(escape_dq "${CMD_LLMOS_EXEC}")
    >>${LOG_FILE} 2>&1"

output_log=${LOG_FILE}
error_log=${LOG_FILE}

pidfile="/var/run/${SYSTEM_NAME}.pid"
respawn_delay=5
respawn_max=0

set -o allexport
if [ -f /etc/environment ]; then . /etc/environment; fi
if [ -f ${FILE_LLMOS_ENV} ]; then . ${FILE_LLMOS_ENV}; fi
set +o allexport
EOF
    $SUDO chmod 0755 "${FILE_LLMOS_SERVICE}"

    $SUDO tee /etc/logrotate.d/${SYSTEM_NAME} >/dev/null << EOF
${LOG_FILE} {
	missingok
	notifempty
	copytruncate
}
EOF
}

# --- get hashes of the current llmos bin and service files
get_installed_hashes() {
    $SUDO sha256sum ${BIN_DIR}/llmos ${FILE_LLMOS_SERVICE} ${FILE_LLMOS_ENV} 2>&1 || true
}

# --- enable and start systemd service ---
systemd_enable() {
    info "systemd: Enabling ${SYSTEM_NAME} unit"
    $SUDO systemctl enable ${FILE_LLMOS_SERVICE} >/dev/null
    $SUDO systemctl daemon-reload >/dev/null
}

systemd_start() {
    info "systemd: Starting ${SYSTEM_NAME}"
    $SUDO systemctl restart --no-block ${SYSTEM_NAME}
    info "Run \"journalctl -u ${SYSTEM_NAME} -f\" to watch logs"
}

# --- enable and start openrc service ---
openrc_enable() {
    info "openrc: Enabling ${SYSTEM_NAME} service for default runlevel"
    $SUDO rc-update add ${SYSTEM_NAME} default >/dev/null
}

openrc_start() {
    info "openrc: Starting ${SYSTEM_NAME}"
    $SUDO "${FILE_LLMOS_SERVICE}" restart
}

# --- startup systemd or openrc service ---
service_enable_and_start() {
    [ "${INSTALL_LLMOS_SKIP_ENABLE}" = true ] && return

    [ "${HAS_SYSTEMD}" = true ] && systemd_enable
	[ "${HAS_OPENRC}" = true ] && openrc_enable

    [ "${INSTALL_LLMOS_SKIP_START}" = true ] && return

    POST_INSTALL_HASHES=$(get_installed_hashes)
    if [ "${PRE_INSTALL_HASHES}" = "${POST_INSTALL_HASHES}" ] && [ "${INSTALL_LLMOS_FORCE_RESTART}" != true ]; then
        info 'No change detected so skipping service start'
        return
    fi

    [ "${HAS_SYSTEMD}" = true ] && systemd_start
	[ "${HAS_OPENRC}" = true ] && openrc_start
    return 0
}

# --- re-evaluate args to include env command ---
eval set -- $(escape "${INSTALL_LLMOS_EXEC}") $(quote "$@")

# --- run the install process --
{
    verify_system
    check_sudo
    setup_env "$@"
    download_and_verify
    create_uninstall
    systemd_disable
    create_env_file
    create_service_file
    service_enable_and_start
}