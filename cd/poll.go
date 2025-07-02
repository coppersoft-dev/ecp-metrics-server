package cd

import (
	"context"
	"time"
)

func (s *Service) StartPollLoop(ctx context.Context, errCh chan<- error) {
	timer := time.NewTimer(0)
	var prevErr bool

	for {
		select {
		case <-ctx.Done():
			s.log.Info("shutting down CD reader")
			close(s.shutdownCh)
			return
		case <-timer.C:
			cdContent, err := s.GetCDContent()
			if err != nil {
				s.log.Error("failed reading CD content from database", "error", err)
				errCh <- err
				prevErr = true
				timer.Reset(3 * time.Second)
				continue
			}
			componentsCur, err := ParseCDContent(cdContent)
			if err != nil {
				s.log.Error("failed parsing CD content", "error", err)
				errCh <- err
				prevErr = true
				timer.Reset(3 * time.Second)
				continue
			}

			logF := s.log.Debug
			if prevErr {
				prevErr = false
				logF = s.log.Info
			}
			logF("Parsed CD content")

			s.cs.SetComponents(componentsCur)

			timer.Reset(60 * time.Second)
		}
	}
}
