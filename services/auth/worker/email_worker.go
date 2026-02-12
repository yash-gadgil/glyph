package worker

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/yash-gadgil/glyph/services/auth/db"
	"github.com/yash-gadgil/glyph/services/auth/utils"
	"go.uber.org/zap"
)

//go:embed templates/verification.html
var verificationTemplateHTML string

//go:embed templates/password_reset.html
var passwordResetTemplateHTML string

var (
	verificationTemplate  = template.Must(template.New("verification").Parse(verificationTemplateHTML))
	passwordResetTemplate = template.Must(template.New("passwordReset").Parse(passwordResetTemplateHTML))
)

type EmailData struct {
	URL string
}

var (
	sendEmail      = utils.SendEmail
	retryBaseDelay = 500 * time.Millisecond
)

type EmailJob struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func StartEmailWorker(cache *db.Cache, log *zap.Logger) chan struct{} {
	stopCh := make(chan struct{})

	if cache.Rdb == nil {
		log.Warn("email_worker_disabled", zap.String("reason", "redis_not_configured"))
		return stopCh
	}

	log.Info("starting_email_worker")

	go runQueueWorker(stopCh, log, "verification", func(ctx context.Context) error {
		return processNextEmail(ctx, cache, log)
	})
	go runQueueWorker(stopCh, log, "password_reset", func(ctx context.Context) error {
		return processNextPasswordReset(ctx, cache, log)
	})

	return stopCh
}

func runQueueWorker(stopCh chan struct{}, log *zap.Logger, name string, process func(context.Context) error) {
	ctx := context.Background()
	backoff := 2 * time.Second
	maxBackoff := 60 * time.Second
	consecutiveErrors := 0

	for {
		select {
		case <-stopCh:
			log.Info("email_worker_stopping_gracefully", zap.String("queue", name))
			return
		default:
			if err := process(ctx); err != nil {
				consecutiveErrors++
				if consecutiveErrors%10 == 1 {
					log.Error("email_worker_error", zap.String("queue", name), zap.Error(err),
						zap.Int("consecutive_errors", consecutiveErrors))
				}
				time.Sleep(backoff)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			} else if consecutiveErrors > 0 {
				log.Info("email_worker_recovered", zap.String("queue", name), zap.Int("errors_recovered_from", consecutiveErrors))
				consecutiveErrors = 0
				backoff = 2 * time.Second
			}
		}
	}
}

func processNextEmail(ctx context.Context, cache *db.Cache, log *zap.Logger) error {
	payload, err := cache.DequeueVerificationEmail(ctx)
	if err != nil {
		return fmt.Errorf("failed to dequeue: %w", err)
	}
	if payload == "" {
		return nil
	}

	var job EmailJob
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		log.Error("invalid_job_payload", zap.String("payload", payload), zap.Error(err))
		if dlqErr := cache.MoveVerificationToDLQ(ctx, payload, fmt.Sprintf("invalid JSON: %v", err)); dlqErr != nil {
			log.Error("verification_dlq_move_failed", zap.String("payload", payload), zap.Error(dlqErr))
		}
		return nil
	}

	baseURL := os.Getenv("EMAIL_VERIFICATION_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	verificationURL := fmt.Sprintf("%s/auth/verify?token=%s", baseURL, job.Token)

	var bodyBuf bytes.Buffer
	if err := verificationTemplate.Execute(&bodyBuf, EmailData{URL: verificationURL}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	var sendErr error
	for attempt := 1; attempt <= 4; attempt++ {
		sendErr = sendEmail(job.Email, "Verify your email", bodyBuf.String())
		if sendErr == nil {
			log.Info("email_sent", zap.String("email", job.Email))
			return nil
		}
		log.Warn("email_send_retry", zap.String("email", job.Email), zap.Int("attempt", attempt), zap.Error(sendErr))
		time.Sleep(time.Duration(attempt) * retryBaseDelay)
	}

	log.Error("email_send_failed_dlq", zap.String("email", job.Email), zap.Error(sendErr))
	if dlqErr := cache.MoveVerificationToDLQ(ctx, payload, fmt.Sprintf("send failed: %v", sendErr)); dlqErr != nil {
		log.Error("verification_dlq_move_failed", zap.String("email", job.Email), zap.Error(dlqErr))
	}
	return nil
}

func processNextPasswordReset(ctx context.Context, cache *db.Cache, log *zap.Logger) error {
	payload, err := cache.DequeuePasswordResetEmail(ctx)
	if err != nil {
		return fmt.Errorf("failed to dequeue password reset: %w", err)
	}
	if payload == "" {
		return nil
	}

	var job EmailJob
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		log.Error("invalid_password_reset_payload", zap.String("payload", payload), zap.Error(err))
		return nil
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, job.Token)

	var bodyBuf bytes.Buffer
	if err := passwordResetTemplate.Execute(&bodyBuf, EmailData{URL: resetURL}); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	var sendErr error
	for attempt := 1; attempt <= 4; attempt++ {
		sendErr = sendEmail(job.Email, "Reset your password", bodyBuf.String())
		if sendErr == nil {
			log.Info("password_reset_email_sent", zap.String("email", job.Email))
			return nil
		}
		log.Warn("password_reset_email_retry", zap.String("email", job.Email), zap.Int("attempt", attempt), zap.Error(sendErr))
		time.Sleep(time.Duration(attempt) * retryBaseDelay)
	}

	log.Error("password_reset_email_failed", zap.String("email", job.Email), zap.Error(sendErr))
	return nil
}
