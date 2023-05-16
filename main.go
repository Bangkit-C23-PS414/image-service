package main

import (
	"context"
	"image-service/adapter/handler"
	"image-service/adapter/repository"
	"image-service/core/service"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx := context.Background()
	store, err := repository.NewImageRepository(ctx)
	if err != nil {
		log.Fatalf("error initialize NewImageRepository with error %v", err)
	}
	imageService := service.NewImageService(store)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go handler.InitHttpServer(*imageService)
	<-done
}
