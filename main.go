package main

import (
 	"log/slog"
 	"time"
	"os"
	"os/signal"
	"context"
	"syscall"
	"api-students/app/repository"
	"api-students/config"
	"api-students/database"
	"api-students/app/service"
)



func main() {
	config.LoadEnv()

	logger := config.NewLogger()

	pool, err := database.NewPool(context.Background()) 
	if err != nil { 
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1) 
	} 
	defer pool.Close() 
	
	// 3. Perakitan: pool -> repository -> handler 
	studentRepository := repository.NewStudentRepository(pool) 
	studentService := service.NewStudentService(studentRepository) 
	
	app := config.NewApp(logger, pool, studentService) 
	port := config.GetEnv("APP_PORT", "3000")

 		go func() { 
		if err := app.Listen(":" + port); err != nil { 
			logger.Error("server berhenti", slog.String("error", err.Error())) 
			os.Exit(1) 
		} 
	}() 
	
	logger.Info("server berjalan", slog.String("port", port)) 
	
	// 5. Graceful shutdown: tunggu Ctrl+C, lalu beri waktu request 
	// yang sedang berjalan untuk selesai. 
	quit := make(chan os.Signal, 1) 
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) 
	<-quit 
	
	logger.Info("sinyal berhenti diterima, menutup server") 
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) 
	defer cancel() 
	
	if err := app.ShutdownWithContext(ctx); err != nil { 
		logger.Error("gagal menutup server dengan rapi", 
			slog.String("error", err.Error())) 
	} 
	
	logger.Info("server berhenti dengan rapi")
}
