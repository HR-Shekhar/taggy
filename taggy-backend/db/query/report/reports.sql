-- =========================================
-- REPORTS
-- =========================================

-- name: CreateReport :one
INSERT INTO report (
    reporter_id,
    target_type,
    target_id,
    reason,
    status
)
VALUES ($1, $2, $3, $4, 'OPEN')
RETURNING
    id,
    reporter_id,
    target_type,
    target_id,
    reason,
    status,
    resolved_at,
    resolved_by,
    created_at,
    updated_at;


-- name: GetOpenReportByReporterAndTarget :one
SELECT
    id,
    reporter_id,
    target_type,
    target_id,
    reason,
    status,
    resolved_at,
    resolved_by,
    created_at,
    updated_at
FROM report
WHERE reporter_id = $1
  AND target_type = $2
  AND target_id = $3
  AND status = 'OPEN'
LIMIT 1;


-- name: ListReportsByReporterID :many
SELECT
    id,
    reporter_id,
    target_type,
    target_id,
    reason,
    status,
    resolved_at,
    resolved_by,
    created_at,
    updated_at
FROM report
WHERE reporter_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2;


-- name: GetReportByIDAndReporterID :one
SELECT
    id,
    reporter_id,
    target_type,
    target_id,
    reason,
    status,
    resolved_at,
    resolved_by,
    created_at,
    updated_at
FROM report
WHERE id = $1
  AND reporter_id = $2;
