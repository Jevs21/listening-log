# Phase 18 — Metabase + Postgres Read Replica

Builds on phase 16 (dockerize & postgres).

## Goal

Add a Postgres streaming read replica and a Metabase instance to the Docker Compose stack. Metabase connects to the replica so analytical queries don't compete with the app for connections or locks on the primary. Metabase is accessible only on localhost for internal use.

## Scope

### In scope

- Configure the primary Postgres for WAL-based streaming replication
- Add a read replica Postgres container
- Add a Metabase container pointing at the replica
- Metabase uses its default embedded SQLite for its own metadata

### Out of scope

- Embedding dashboards/questions into the web app (future phase)
- Metabase user management beyond the initial admin setup
- Exposing Metabase through Cloudflare tunnel or to external users
- Custom Metabase dashboards or questions (that's manual work after deploy)

## Infrastructure changes

### Primary Postgres (`db`) — replication config

Add custom config to enable WAL streaming. Create `postgres/primary/postgresql.conf` overrides:

```
wal_level = replica
max_wal_senders = 3
max_replication_slots = 1
hot_standby = on
```

Add a replication user via an init script at `postgres/primary/init-replication.sh`:

```sql
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'replicator';
```

Mount both into the `db` service:

```yaml
volumes:
  - pgdata:/var/lib/postgresql/data
  - ./postgres/primary/postgresql.conf:/etc/postgresql/postgresql.conf
  - ./postgres/primary/init-replication.sh:/docker-entrypoint-initdb.d/init-replication.sh
```

Add `command: postgres -c config_file=/etc/postgresql/postgresql.conf` to use the custom config.

Update `pg_hba.conf` to allow replication connections from the replica container. Create `postgres/primary/pg_hba.conf`:

```
# Default entries
local   all             all                                     trust
host    all             all             0.0.0.0/0               md5

# Replication
host    replication     replicator      0.0.0.0/0               md5
```

Mount it: `./postgres/primary/pg_hba.conf:/etc/postgresql/pg_hba.conf`

And add to the `command`: `postgres -c config_file=/etc/postgresql/postgresql.conf -c hba_file=/etc/postgresql/pg_hba.conf`

### Read replica (`db-replica`)

The replica needs a custom entrypoint that performs a `pg_basebackup` from the primary on first boot, then starts Postgres in standby mode.

Create `postgres/replica/entrypoint.sh`:

```bash
#!/bin/bash
set -e

# Only do base backup if data dir is empty (first boot)
if [ -z "$(ls -A /var/lib/postgresql/data)" ]; then
  pg_basebackup -h db -U replicator -D /var/lib/postgresql/data -Fp -Xs -R -P
fi

exec postgres
```

The `-R` flag in `pg_basebackup` automatically creates `standby.signal` and configures `primary_conninfo` in `postgresql.auto.conf`.

```yaml
db-replica:
  image: postgres:17-alpine
  environment:
    PGPASSWORD: replicator
  volumes:
    - pgdata-replica:/var/lib/postgresql/data
    - ./postgres/replica/entrypoint.sh:/entrypoint.sh
  entrypoint: ["/bin/bash", "/entrypoint.sh"]
  user: postgres
  depends_on:
    db:
      condition: service_healthy
  healthcheck:
    test: ["CMD", "pg_isready", "-U", "listening_log"]
    interval: 5s
    timeout: 3s
    retries: 10
```

### Metabase (`metabase`)

```yaml
metabase:
  image: metabase/metabase:latest
  ports:
    - "3000:3000"
  environment:
    MB_DB_TYPE: h2  # default SQLite-like embedded DB
    MB_JETTY_HOST: 0.0.0.0
    MB_JETTY_PORT: 3000
  volumes:
    - metabase-data:/metabase-data
  depends_on:
    db-replica:
      condition: service_healthy
```

Metabase stores its own data (dashboards, questions, settings) in an embedded H2 database at `/metabase-data/metabase.db`, persisted via the `metabase-data` volume.

On first launch, the admin configures the Postgres data source through the Metabase setup wizard, pointing at `db-replica:5432` with the `listening_log` credentials (read-only by nature of being a replica).

### Volumes

Add to `volumes:`:

```yaml
volumes:
  pgdata:
  pgdata-replica:
  metabase-data:
```

## File structure

```
postgres/
  primary/
    postgresql.conf
    pg_hba.conf
    init-replication.sh
  replica/
    entrypoint.sh
```

## Definition of done

- [ ] `docker compose up` starts all four services: `db`, `db-replica`, `app`, `metabase`
- [ ] Primary Postgres accepts replication connections
- [ ] Replica streams from primary and contains the same data (verify with a `SELECT count(*)` on a table from both)
- [ ] Replica rejects write queries (`INSERT` returns an error)
- [ ] Metabase is accessible at `http://localhost:3000`
- [ ] Metabase can be configured via setup wizard to connect to `db-replica:5432` and query app data
- [ ] App on port 8080 continues to work normally with no performance change
- [ ] Writes to primary appear on replica within seconds
