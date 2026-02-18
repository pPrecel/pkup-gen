package send

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
	"gopkg.in/mail.v2"
)

type dialer struct {
	logger *pterm.Logger
	sender mail.SendCloser
}

func NewDialer(logger *pterm.Logger, address string, port int, username, password string) (*dialer, error) {
	d := mail.NewDialer(address, port, username, password)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	logger.Trace("dialing to server", logger.Args("address", address, "port", port, "username", username, "isPassword", password != ""))
	sendCloser, err := d.Dial()
	if err != nil {
		return nil, err
	}

	err = d.DialAndSend()
	if err != nil {
		return nil, err
	}

	return &dialer{
		logger: logger,
		sender: sendCloser,
	}, nil
}

func (d *dialer) Close() {
	d.logger.Trace("closing dialer")
	d.sender.Close()
}

type message struct {
	from           string
	to             string
	subject        string
	body           string
	attachmentPath string
}

func (d *dialer) SendMails(messages []*message) error {
	mailMessages := make([]*mail.Message, len(messages))
	for i, m := range messages {
		mailMessage := mail.NewMessage()
		mailMessage.SetHeader("From", m.from)
		mailMessage.SetHeader("To", m.to)
		mailMessage.SetHeader("Subject", m.subject)
		mailMessage.SetBody("text/html", m.body)
		mailMessage.Attach(m.attachmentPath)

		mailMessages[i] = mailMessage
	}

	for _, message := range mailMessages {
		err := d.sendMailWithRetry(message)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *dialer) sendMailWithRetry(messages *mail.Message) error {
	var err error
	iterations := 5
	for i := 0; i < iterations; i++ {
		err = mail.Send(d.sender, messages)
		if err == nil {
			return nil
		}

		delay := time.Minute
		d.logger.Warn("failed to send mail, retrying", d.logger.Args("iteration", fmt.Sprintf("%d/%d", i+1, iterations), "delay", delay.String(), "error", err.Error()))
		time.Sleep(delay)
	}

	return err
}
