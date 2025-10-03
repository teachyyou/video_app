package util

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strconv"
	"time"
)

type ffprobeFormat struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func ProbeDuration(ctx context.Context, path string) (time.Duration, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "json",
		path,
	)

	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("ffprobe: %w", err)
	}

	var p ffprobeFormat
	if err := json.Unmarshal(out, &p); err != nil {
		return 0, fmt.Errorf("ffprobe json: %w", err)
	}
	if p.Format.Duration == "" {
		return 0, fmt.Errorf("ffprobe: empty duration")
	}

	sec, err := strconv.ParseFloat(p.Format.Duration, 64)
	if err != nil {
		return 0, fmt.Errorf("ffprobe parse: %w", err)
	}

	return time.Duration(math.Round(sec)) * time.Second, nil
}
