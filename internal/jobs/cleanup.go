package jobs

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/openclaw/relay-server-go/internal/repository"
)

type CleanupJob struct {
	adminSessionRepo     repository.AdminSessionRepository
	portalSessionRepo    repository.PortalSessionRepository
	portalAccessCodeRepo repository.PortalAccessCodeRepository
	pairingCodeRepo      repository.PairingCodeRepository
	inboundMsgRepo       repository.InboundMessageRepository
	sessionRepo          repository.SessionRepository
	interval             time.Duration
	done                 chan struct{}
}

func NewCleanupJob(
	adminSessionRepo repository.AdminSessionRepository,
	portalSessionRepo repository.PortalSessionRepository,
	portalAccessCodeRepo repository.PortalAccessCodeRepository,
	pairingCodeRepo repository.PairingCodeRepository,
	inboundMsgRepo repository.InboundMessageRepository,
	sessionRepo repository.SessionRepository,
	interval time.Duration,
) *CleanupJob {
	return &CleanupJob{
		adminSessionRepo:     adminSessionRepo,
		portalSessionRepo:    portalSessionRepo,
		portalAccessCodeRepo: portalAccessCodeRepo,
		pairingCodeRepo:      pairingCodeRepo,
		inboundMsgRepo:       inboundMsgRepo,
		sessionRepo:          sessionRepo,
		interval:             interval,
		done:                 make(chan struct{}),
	}
}

func (j *CleanupJob) Start() {
	go j.run()
	log.Info().Dur("interval", j.interval).Msg("cleanup job started")
}

func (j *CleanupJob) Stop() {
	close(j.done)
	log.Info().Msg("cleanup job stopped")
}

func (j *CleanupJob) run() {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	j.cleanup()

	for {
		select {
		case <-j.done:
			return
		case <-ticker.C:
			j.cleanup()
		}
	}
}

func (j *CleanupJob) cleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	j.runCleanup(ctx, "admin sessions", j.adminSessionRepo.DeleteExpired)
	j.runCleanup(ctx, "portal sessions", j.portalSessionRepo.DeleteExpired)
	j.runCleanup(ctx, "portal access codes", j.portalAccessCodeRepo.DeleteExpired)
	j.runCleanup(ctx, "pairing codes", j.pairingCodeRepo.DeleteExpired)
	j.runCleanup(ctx, "inbound messages", j.inboundMsgRepo.MarkExpired)
	if j.sessionRepo != nil {
		j.runCleanup(ctx, "sessions", j.sessionRepo.DeleteExpired)
	}
}

func (j *CleanupJob) runCleanup(ctx context.Context, name string, fn func(context.Context) (int64, error)) {
	count, err := fn(ctx)
	if err != nil {
		log.Error().Err(err).Msgf("failed to cleanup %s", name)
	} else if count > 0 {
		log.Info().Int64("count", count).Msgf("cleaned up %s", name)
	}
}
