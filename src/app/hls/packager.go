package hls

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"go.uber.org/fx"
)

// Низкоуровневая конвертация
type Packager interface {
	PackageHLS(ctx context.Context, inPath string, outDir string) error
}

type FFmpegPackager struct {
}

func NewFFmpegPackager() Packager {
	return &FFmpegPackager{}
}

func (*FFmpegPackager) PackageHLS(ctx context.Context, inPath string, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// Параметры — подправь под свой материал
	segmentDur := "6" // секунды
	gop := "360"      // если ~8 fps*6 = 48; для 30fps ставь 180, для 25fps — 150 и т.п.

	args := []string{
		"-y",
		// GPU decode path
		"-hwaccel", "cuda",
		"-hwaccel_output_format", "cuda",
		"-i", inPath,

		// Видео на NVENC
		"-c:v", "h264_nvenc",
		"-preset", "p3", // p1 лучше качество, p7 быстрее
		"-b:v", "5M",
		"-maxrate", "5M",
		"-bufsize", "10M",
		"-g", gop,
		"-keyint_min", gop,
		"-sc_threshold", "0",

		// Аудио
		"-c:a", "aac",
		"-ac", "2",
		"-b:a", "128k",

		// HLS
		"-hls_time", segmentDur,
		"-hls_list_size", "0",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_filename", filepath.Join(outDir, "seg_%06d.ts"),
		filepath.Join(outDir, "index.m3u8"),
	}

	if err := runFFmpeg(ctx, args...); err != nil {
		return err
	}

	thumbPath := filepath.Join(outDir, "preview.png")
	argsThumb := []string{
		"-y",
		"-i", inPath,

		"-vf", "thumbnail,scale=-2:360",
		"-frames:v", "1",
		thumbPath,
	}
	if err := runFFmpeg(ctx, argsThumb...); err != nil {
		return err
	}

	return nil
}

func runFFmpeg(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

var FFmpegPackagerModule = fx.Module("ffmpeg", fx.Provide(NewFFmpegPackager))
