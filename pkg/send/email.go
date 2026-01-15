package send

import (
	"os"
	"time"

	"github.com/pPrecel/PKUP/pkg/config"
	"github.com/pterm/pterm"
)

type Sender interface {
	ForConfig(*config.Config, string) error
}

type sender struct {
	logger *pterm.Logger
}

func New(logger *pterm.Logger) Sender {
	return &sender{
		logger: logger,
	}
}

func (s *sender) ForConfig(config *config.Config, zipPrefix string) error {
	body := ""
	if config.Send.HTMLBodyPath != "" {
		s.logger.Debug("reading body", s.logger.Args("path", config.Send.HTMLBodyPath))
		bodyBytes, err := os.ReadFile(config.Send.HTMLBodyPath)
		if err != nil {
			return err
		}

		body = string(bodyBytes)
	}

	if config.Send.PerDial == 0 {
		// send at least one message before delay
		config.Send.PerDial = 1
	}

	dialer, err := NewDialer(s.logger, config.Send.ServerAddress, config.Send.ServerPort, config.Send.Username, config.Send.Password)
	if err != nil {
		return err
	}
	defer dialer.Close()

	zipper := NewZipper(s.logger)

	var j int
	for i := 0; i < len(config.Reports); i += config.Send.PerDial {
		j += config.Send.PerDial
		if j > len(config.Reports) {
			j = len(config.Reports)
		}

		messages := make([]*message, j-i)
		for iter, report := range config.Reports[i:j] {
			s.logger.Debug("zipping report", s.logger.Args("dir", report.OutputDir, "prefix", zipPrefix))
			reportFile, err := zipper.Do(report.OutputDir, zipPrefix)
			if err != nil {
				return err
			}

			s.logger.Info("building email message", s.logger.Args("from", config.Send.From, "to", report.Email, "attachmentPath", reportFile))
			messages[iter] = &message{
				from:           config.Send.From,
				to:             report.Email,
				subject:        config.Send.Subject,
				body:           body,
				attachmentPath: reportFile,
			}
		}

		s.logger.Info("sending built messages", s.logger.Args("len", len(messages)))
		err = dialer.SendMail(messages...)
		if err != nil {
			return err
		}

		if i+config.Send.PerDial < len(config.Reports) {
			s.logger.Debug("waiting...", s.logger.Args("delay", config.Send.DialDelay.String()))
			time.Sleep(config.Send.DialDelay)
		}
	}

	return nil
}
