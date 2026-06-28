# CLAUDE.md

## Critical Rules

### Docker Volumes — NEVER Delete

**NEVER delete Docker volumes** (`docker volume rm`, `docker volume prune`, `docker-compose down -v`, `docker system prune --volumes`, etc.) under any circumstances.

If a task or fix appears to require deleting a Docker volume, **STOP and alert the user** before proceeding. Explain what you were about to do and why it would involve volume deletion, so they can decide how to proceed.
