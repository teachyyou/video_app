ALTER TABLE videos
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS retry_attempt,
  DROP COLUMN IF EXISTS failure_reason,
  DROP COLUMN IF EXISTS processing_started_at,
  DROP COLUMN IF EXISTS hls_ready_at;

DROP INDEX IF EXISTS videos_status_idx;
