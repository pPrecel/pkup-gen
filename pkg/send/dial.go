package send

import (
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

	logger.Trace("dialing to server", logger.Args("address", address, "port", port, "username", username))
	sendCloser, err := d.Dial()
	d.DialAndSend()
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

func (d *dialer) SendMail(from, subject, destination, body, attachmentPath string) error {
	message := mail.NewMessage()
	message.SetHeader("From", from)
	message.SetHeader("To", destination)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	message.Attach(attachmentPath)

	return mail.Send(d.sender, message)
}
