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

func (d *dialer) SendMail(messages ...*message) error {
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

	return mail.Send(d.sender, mailMessages...)
}
