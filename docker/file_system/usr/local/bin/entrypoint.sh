#!/usr/bin/env bash
set -e

APP_NAME="caspbx"
APP_BIN="/usr/local/bin/${APP_NAME}"

export TZ="${TZ:-America/New_York}"
export CONFIG_DIR="${CONFIG_DIR:-/config/${APP_NAME}}"
export DATA_DIR="${DATA_DIR:-/data/${APP_NAME}}"

log() {
    echo "[entrypoint] $(date '+%Y-%m-%d %H:%M:%S') $*"
}

setup_timezone() {
    if [ -n "${TZ:-}" ] && [ -f "/usr/share/zoneinfo/${TZ}" ]; then
        ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime
        echo "${TZ}" > /etc/timezone
    fi
}

initialize_aio_postgres() {
    if ! command -v initdb >/dev/null 2>&1; then
        return
    fi

    mkdir -p /data/db/postgres /data/db/valkey /data/log/postgres /run/postgresql /run/valkey
    chown -R postgres:postgres /data/db/postgres /data/log/postgres /run/postgresql
    chmod 700 /data/db/postgres
    chmod 755 /run/valkey

    if [ ! -f /data/db/postgres/PG_VERSION ]; then
        log "Initializing PostgreSQL database..."
        su - postgres -c "initdb -D /data/db/postgres"
        cp /config/postgres/postgresql.conf /data/db/postgres/postgresql.conf
        su - postgres -c "pg_ctl -D /data/db/postgres -l /data/log/postgres/init.log start"
        sleep 3
        su - postgres -c "psql -c \"CREATE USER ${DB_USER:-caspbx} WITH PASSWORD '${DB_PASSWORD:-caspbx}';\"" || true
        su - postgres -c "psql -c \"CREATE DATABASE ${DB_NAME:-caspbx} OWNER ${DB_USER:-caspbx};\"" || true
        su - postgres -c "psql -c \"GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME:-caspbx} TO ${DB_USER:-caspbx};\"" || true
        su - postgres -c "pg_ctl -D /data/db/postgres stop"
    fi
}

setup_timezone

if [ "${AIO:-false}" = "true" ]; then
    initialize_aio_postgres
    export TOR_ENABLED="${TOR_ENABLED:-false}"
    log "Starting ${APP_NAME} all-in-one services..."
    exec /usr/bin/supervisord -c /etc/supervisor/supervisord.conf
fi

log "Starting ${APP_NAME}..."

FLAGS="--address ${ADDRESS:-0.0.0.0} --port ${PORT:-80}"
[ "${DEBUG:-false}" = "true" ] && FLAGS="${FLAGS} --debug"

exec ${APP_BIN} ${FLAGS} "$@"
