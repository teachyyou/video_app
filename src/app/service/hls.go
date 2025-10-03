package service

import (
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/hls"
	"awesomeProject/src/app/repository"
	"awesomeProject/src/util"
	"context"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Верхнеуровневая конвертация
type Converter interface {
	Start(ctx context.Context)
	Enqueue(slug string)
}

type ConversionService struct {
	config   *config.Config
	packager hls.Packager
	repo     *repository.VideoRepository
	log      *zap.Logger

	jobs          chan string
	parallelLimit int

	runCtx context.Context
	cancel context.CancelFunc
}

func NewConversionService(cfg *config.Config, repo *repository.VideoRepository, pkg hls.Packager, logger *zap.Logger) *ConversionService {
	return &ConversionService{
		config:        cfg,
		packager:      pkg,
		repo:          repo,
		log:           logger,
		jobs:          make(chan string, 256),
		parallelLimit: cfg.Conv.Parallel,
	}
}

func (svc *ConversionService) Start() {
	var semChan = make(chan struct{}, svc.parallelLimit)
	svc.log.Info("converter started", zap.Int("parallel", svc.parallelLimit))
	go func() {
		for {
			select {
			case <-svc.runCtx.Done():
				svc.log.Info("converted stopped")
				return
			case slug := <-svc.jobs:
				svc.log.Info("extracted from chan, waiting for semaphore", zap.String("slug", slug))
				semChan <- struct{}{}
				go func(sl string) {
					defer func() { <-semChan }()
					err := svc.handleJob(svc.runCtx, sl)
					svc.log.Info("called job handler", zap.String("slug", slug))
					if err != nil {
						svc.log.Error("handle job failed", zap.Error(err))
						return
					}
					svc.log.Debug("job finished", zap.String("slug", sl))
				}(slug)
			}
		}

	}()
}

func (svc *ConversionService) Enqueue(slug string) {
	svc.log.Info("enqueued worker for ", zap.String("slug", slug))
	svc.jobs <- slug
}

func (svc *ConversionService) handleJob(ctx context.Context, slug string) error {
	video, err := svc.repo.GetBySlug(ctx, slug)
	svc.log.Info("handling job for ", zap.String("slug", slug))
	if err != nil {
		return err
	}
	defer os.RemoveAll(filepath.Join(svc.config.Conv.TmpDir, slug))

	if err := svc.repo.SetProcessing(ctx, slug, time.Now()); err != nil {
		return err
	}

	datePath := video.CreatedAt.Format("2006/01/02")

	inPath := filepath.Join(svc.config.Data.RawDir, datePath, slug, "source.mp4")
	outDir := filepath.Join(svc.config.Conv.TmpDir, slug, "hls")

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		svc.log.Error("create output dir failed", zap.Error(err), zap.String("slug", slug))
		_ = svc.repo.SetInterrupted(ctx, slug, err)
		return err
	}
	if err := svc.packager.PackageHLS(ctx, inPath, outDir); err != nil {
		svc.log.Error("packaging failed", zap.Error(err), zap.String("slug", slug))
		_ = svc.repo.SetInterrupted(ctx, slug, err)
		return err
	}

	destPath := filepath.Join(svc.config.Conv.ConvDir, datePath, slug)

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		svc.log.Error("make final parent dir failed", zap.Error(err), zap.String("slug", slug))
		_ = svc.repo.SetInterrupted(ctx, slug, err)
		return err
	}
	if err := util.MoveDir(outDir, destPath); err != nil {
		svc.log.Error("move artifacts failed",
			zap.Error(err),
			zap.String("from", outDir),
			zap.String("to", destPath),
			zap.String("slug", slug),
		)
		_ = svc.repo.SetInterrupted(ctx, slug, err)
		return err
	}
	// опционально подчистим корень tmp для этого slug

	if err := svc.repo.SetReady(ctx, slug, time.Now()); err != nil {
		return err
	}
	svc.log.Info("converting succeeded for video", zap.String("slug", slug))

	return nil
}

var ConvServiceModule = fx.Module("conversion_service",
	fx.Provide(NewConversionService),
	fx.Invoke(func(lc fx.Lifecycle, cs *ConversionService) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				cs.runCtx, cs.cancel = context.WithCancel(context.Background())
				cs.Start()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if cs.cancel != nil {
					cs.cancel()
				}
				return nil
			},
		})
	}),
)
