package main

import (
	"awesomeProject/src/app/cache"
	"awesomeProject/src/app/config"
	"awesomeProject/src/app/handler"
	"awesomeProject/src/app/hls"
	"awesomeProject/src/app/repository"
	"awesomeProject/src/app/server"
	"awesomeProject/src/app/service"
	"mime"

	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

func init() {
	_ = godotenv.Load()

	mime.AddExtensionType(".m3u8", "application/vnd.apple.mpegurl")
	mime.AddExtensionType(".m3u", "application/vnd.apple.mpegurl")
	mime.AddExtensionType(".ts", "video/mp2t")
	mime.AddExtensionType(".m4s", "video/iso.segment") // LL-HLS
	mime.AddExtensionType(".mp4", "video/mp4")
	mime.AddExtensionType(".key", "application/octet-stream")
}

func main() {

	fx.New(
		server.Module,
		handler.HelloModule,
		handler.VideoModule,
		handler.MediaModule,
		config.Module,
		config.DbModule,
		cache.CacheModule,
		hls.FFmpegPackagerModule,
		service.VideoModule,
		service.ConvServiceModule,
		repository.VideoRepoModule,
	).Run()
}
