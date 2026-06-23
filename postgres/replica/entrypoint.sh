#!/bin/bash
set -e

# Only do base backup if data dir is empty (first boot)
if [ -z "$(ls -A /var/lib/postgresql/data)" ]; then
  pg_basebackup -h db -U replicator -D /var/lib/postgresql/data -Fp -Xs -R -P
fi

chmod 0700 /var/lib/postgresql/data

exec postgres
