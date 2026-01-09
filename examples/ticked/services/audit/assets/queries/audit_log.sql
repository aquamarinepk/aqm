-- name: InsertAuditRecord :exec
INSERT INTO audit_log (id, event_type, item_id, user_id, payload, source, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAuditRecords :many
SELECT id, event_type, COALESCE(item_id, '') AS item_id, user_id, payload, source, created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT $1;

-- name: GetAuditRecordByID :one
SELECT id, event_type, COALESCE(item_id, '') AS item_id, user_id, payload, source, created_at
FROM audit_log
WHERE id = $1;

-- name: ListAuditRecordsByUserID :many
SELECT id, event_type, COALESCE(item_id, '') AS item_id, user_id, payload, source, created_at
FROM audit_log
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListAuditRecordsByEventType :many
SELECT id, event_type, COALESCE(item_id, '') AS item_id, user_id, payload, source, created_at
FROM audit_log
WHERE event_type = $1
ORDER BY created_at DESC
LIMIT $2;
