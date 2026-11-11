#!/bin/bash
set -e

SYS_INSTALLED_FLAG="/usr/local/.system_dependencies_installed"
PY_INSTALLED_FLAG="/usr/local/lib/python3.9/site-packages/.python_dependencies_installed"
REQ_HASH_FILE="/usr/local/lib/python3.9/site-packages/.requirements_hash"


DEFAULT_SYS_DEPS="libjpeg62-turbo zlib1g libpq5 tzdata ca-certificates curl libgomp1 fonts-dejavu-core"

log() {
    echo "[Entrypoint] $(date '+%Y-%m-%d %H:%M:%S') $1"
}

if [ -n "$DEBIAN_MIRROR" ]; then
    log "Updating Debian sources to $DEBIAN_MIRROR..."
    sed -i "s@http://deb.debian.org@$DEBIAN_MIRROR@g" /etc/apt/sources.list
    sed -i "s@http://security.debian.org@$DEBIAN_MIRROR@g" /etc/apt/sources.list
fi

install_system_deps() {
    local PACKAGES="$DEFAULT_SYS_DEPS $APT_PACKAGES"

    if [ ! -f "$SYS_INSTALLED_FLAG" ]; then
        log "First run detected. Installing system dependencies: $PACKAGES"

        apt-get update
        apt-get install -y --no-install-recommends $PACKAGES

        if [ -n "$TZ" ]; then
            ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
        fi

        rm -rf /var/lib/apt/lists/*

        touch "$SYS_INSTALLED_FLAG"
        log "System dependencies installed."
    else
        log "System dependencies already installed."
    fi
}

install_python_deps() {
    CURRENT_REQ_HASH=$(md5sum requirements.txt | awk '{print $1}')
    STORED_REQ_HASH=""
    [ -f "$REQ_HASH_FILE" ] && STORED_REQ_HASH=$(cat "$REQ_HASH_FILE")

    if [ ! -f "$PY_INSTALLED_FLAG" ] || [ "$CURRENT_REQ_HASH" != "$STORED_REQ_HASH" ]; then
        log "Installing/Updating Python dependencies..."

        if [ -n "$PIP_INDEX_URL" ]; then
            pip config set global.index-url "$PIP_INDEX_URL"
        fi

        pip install --no-cache-dir --compile -r requirements.txt

        touch "$PY_INSTALLED_FLAG"
        echo "$CURRENT_REQ_HASH" > "$REQ_HASH_FILE"
        log "Python dependencies installed."
    else
        log "Python dependencies are up-to-date."
    fi
}

install_system_deps
install_python_deps

mkdir -p logs data

log "Starting application: *"
exec "$@"