ALTER TABLE videos
  ADD COLUMN status text NOT NULL DEFAULT 'uploaded'
    CHECK (status IN ('uploaded','processing','complete','interrupted','archived')),
  ADD COLUMN retry_attempt int NOT NULL DEFAULT 0,
  ADD COLUMN failure_reason text,
  ADD COLUMN processing_started_at timestamptz,
  ADD COLUMN hls_ready_at timestamptz;

-- индекс по статусу для быстрых выборок
CREATE INDEX IF NOT EXISTS videos_status_idx ON videos (status);
