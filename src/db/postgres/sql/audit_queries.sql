-- name: InsertRequestLog :one
INSERT INTO request_logs (
  user_id, api_key_id, model, endpoint, messages, parameters, client_ip
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: InsertResponseLog :one
INSERT INTO response_logs (
  request_id, response, latency_ms, input_tokens, output_tokens, total_tokens
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: InsertFirewallEvent :one
INSERT INTO firewall_events (
  request_id, firewall_id, firewall_type, blocked, blocked_reason, risk_score
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: InsertAuditArchive :one
INSERT INTO audit_archives (
  request_id, s3_path, archive_hash
)
VALUES ($1, $2, $3)
RETURNING *;

-- name: MarkRequestArchived :exec
UPDATE request_logs
SET archived = TRUE
WHERE request_id = $1;

-- name: GetUnarchivedRequests :many
SELECT * FROM request_logs
WHERE archived = FALSE
AND received_at < now() - interval '10 minutes';

-- name: GetRequestFullTrace :many
SELECT rl.*, res.response, res.latency_ms, pe.*
FROM request_logs rl
LEFT JOIN response_logs res ON rl.request_id = res.request_id
LEFT JOIN firewall_events pe ON rl.request_id = pe.request_id
WHERE rl.request_id = $1;
