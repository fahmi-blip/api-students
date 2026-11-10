package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"api-students/app/model"
	"api-students/config"
	"api-students/database"
)

func main() {
	config.LoadEnv()

	outboxID := flag.Int64("outbox", 0, "ID outbox yang akan dikirim ulang")
	copies := flag.Int("copies", 1, "Jumlah salinan yang akan di-enqueue ke Redis")
	flag.Parse()

	if *outboxID <= 0 {
		log.Fatal("flag -outbox harus diisi dengan ID valid")
	}
	if *copies <= 0 {
		log.Fatal("flag -copies harus minimal 1")
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	defer pool.Close()

	var eventType string
	var payload []byte
	err = pool.QueryRow(ctx, "SELECT event_type, payload FROM outbox WHERE id = $1", *outboxID).Scan(&eventType, &payload)
	if err != nil {
		log.Fatalf("mengambil outbox %d: %v", *outboxID, err)
	}

	job := model.NotificationJob{
		OutboxID:  *outboxID,
		EventType: eventType,
		Payload:   payload,
	}
	jobPayload, err := json.Marshal(job)
	if err != nil {
		log.Fatalf("marshal job: %v", err)
	}

	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()

	for i := 1; i <= *copies; i++ {
		task := asynq.NewTask(model.TaskSendNotification, jobPayload,
			asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second),
			asynq.Retention(24*time.Hour),
		)
		_, err := client.EnqueueContext(ctx, task)
		if err != nil {
			log.Fatalf("gagal enqueue salinan ke-%d: %v", i, err)
		}
	}

	fmt.Printf("Berhasil memasukkan %d salinan task outbox %d ke Redis queue %s\n", *copies, *outboxID, model.QueueNotifications)
}
