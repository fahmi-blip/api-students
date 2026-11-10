package service 
import ( 
	"context" 
	"encoding/json" 
	"errors" 
	"fmt" 
	"log/slog" 
	"github.com/hibiken/asynq" 
	"api-students/app/model" 
	"api-students/app/repository" 
	"api-students/helper" 
) 
// NotificationWorker mengerjakan task "notification:send". 
// // Bagi worker, Asynq adalah framework-nya — sama seperti Fiber bagi API. 
// Karena itu method di sini menerima *asynq.Task, seperti method service 
// lain menerima *fiber.Ctx. 

type NotificationWorker struct { 
	users repository.UserRepository 
	notifications repository.NotificationRepository 
	tx *repository.TxManager 
	mailer helper.Mailer 
	logger *slog.Logger
	courses repository.CourseRepository 
}

func NewNotificationWorker( 
	users repository.UserRepository,
	notifications repository.NotificationRepository, 
	tx *repository.TxManager, 
	mailer helper.Mailer, 
	logger *slog.Logger,
	courses repository.CourseRepository, 
) *NotificationWorker { 
	return &NotificationWorker{users: users, notifications: notifications, 
		tx: tx, mailer: mailer, logger: logger, courses: courses,} 
} 

// HandleSend mengembalikan error untuk meminta Asynq melakukan retry nanti, 
// atau error yang membungkus asynq.SkipRetry untuk menyatakan "jangan di-retry, 
// task ini tidak mungkin berhasil". // 
// Asynq menjamin setiap task dijalankan PALING SEDIKIT satu kali, bukan 
// TEPAT satu kali. Task yang sama bisa datang dua kali, misalnya bila 
// worker mati setelah mengirim email tetapi sebelum sempat melapor ke Redis. 
// Karena itu handler ini harus idempoten: menjalankannya dua kali untuk 
// task yang sama harus berakibat sama dengan menjalankannya sekali. 
func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error { 
	var job model.NotificationJob 
	if err := json.Unmarshal(t.Payload(), &job); err != nil { 
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err) 
	}
	if job.EventType == model.EventCourseQuotaChanged {
		return w.handleQuotaChanged(ctx, job)
	} 
	var event model.EnrollmentEvent 
	if err := json.Unmarshal(job.Payload, &event); err != nil { 
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err) 
	} 
	
	user, err := w.users.FindByID(ctx, event.UserID) 
	if errors.Is(err, repository.ErrNotFound) { 
		// Akunnya sudah dihapus. Melakukan retry tidak akan memunculkannya. 
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	} 
	if err != nil { 
		return err // database sedang bermasalah: layak di-retry 
	} 
	
	msg, err := BuildMessage(job.EventType, event, user) 
	if err != nil { 
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err) 
	} 
	alreadySent := false 
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error { 
		// Claim dan pengiriman berada di dalam SATU transaction: 
		// - pengiriman gagal -> claim ikut dibatalkan -> attempt 
		// berikutnya masih dapat mengirim; 
		// - pengiriman sukses -> claim di-commit -> attempt berikutnya 
		// melihat claim ini dan tidak mengirim lagi. 
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{ 
			OutboxID: job.OutboxID, UserID: user.ID, Channel: "email", 
			Subject: msg.Subject, Body: msg.Body, 
		}) 
		if err != nil { 
			return err 
		} 
		if !claimed { 
			alreadySent = true 
			return nil 
		} 
		return w.mailer.Send(ctx, msg) 
	}) 
	if errors.Is(err, helper.ErrPermanentDelivery) { 
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err) 
	} 
	if err != nil { 
		return err 
	} 
	retry, _ := asynq.GetRetryCount(ctx) 
	w.logger.Info("notification_processed", 
		slog.Int64("outbox_id", job.OutboxID), 
		slog.String("event", job.EventType), 
		slog.Int("retry", retry), slog.Bool("already_sent", alreadySent)) 
	return nil 
}

func (w *NotificationWorker) handleQuotaChanged(ctx context.Context, job model.NotificationJob) error {
	var ev model.QuotaEvent
	if err := json.Unmarshal(job.Payload, &ev); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}
	ids, err := w.courses.EnrolledUserIDs(ctx, ev.CourseID)
	if err != nil {
		return err
	}
	sent, skipped := 0, 0
	for _, uid := range ids {
		user, err := w.users.FindByID(ctx, uid)
		if errors.Is(err, repository.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		msg := BuildQuotaMessage(ev, user)
		already := false
		err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
			claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{
				OutboxID: job.OutboxID, UserID: user.ID, Channel: "email",
				Subject: msg.Subject, Body: msg.Body,
			})
			if err != nil {
				return err
			}
			if !claimed {
				already = true
				return nil
			}
			return w.mailer.Send(ctx, msg)
		})
		if errors.Is(err, helper.ErrPermanentDelivery) {
			w.logger.Warn("quota_recipient_skipped", slog.Int("user_id", uid), slog.String("error", err.Error()))
			continue // satu alamat salah tidak boleh menggagalkan peserta lain
		}
		if err != nil {
			return err // temporary: Asynq me-retry, peserta yang sudah ter-claim dilewati
		}
		if already {
			skipped++
		} else {
			sent++
		}
	}
	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed",
		slog.Int64("outbox_id", job.OutboxID), slog.String("event", job.EventType),
		slog.Int("retry", retry), slog.Bool("already_sent", sent == 0),
		slog.Int("sent", sent), slog.Int("skipped", skipped))
	return nil
}