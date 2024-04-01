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

func (s *sender) ForConfig(config *config.Config, zipSuffix string) error {
	body := ""
	if config.Send.HTMLBodyPath != "" {
		s.logger.Debug("reading body", s.logger.Args("path", config.Send.HTMLBodyPath))
		bodyBytes, err := os.ReadFile(config.Send.HTMLBodyPath)
		if err != nil {
			return err
		}

		body = string(bodyBytes)
	}

	dialer, err := NewDialer(s.logger, config.Send.ServerAddress, config.Send.ServerPort, config.Send.Username, config.Send.Password)
	if err != nil {
		return err
	}
	defer dialer.Close()

	zipper := NewZipper(s.logger)

	for i, report := range config.Reports {
		s.logger.Debug("zipping report", s.logger.Args("dir", report.OutputDir, "suffix", zipSuffix))
		reportFile, err := zipper.Do(report.OutputDir, zipSuffix)
		if err != nil {
			return err
		}

		s.logger.Info("sending report", s.logger.Args("from", config.Send.From, "to", report.Email, "attachmentPath", reportFile))
		err = dialer.SendMail(config.Send.From, config.Send.Subject, report.Email, body, reportFile)
		if err != nil {
			return err
		}

		if config.Send.Delay != nil && i < len(config.Reports)-1 {
			s.logger.Debug("waiting...", s.logger.Args("delay", config.Send.Delay.String()))
			time.Sleep(*config.Send.Delay)
		}
	}

	return nil
}
