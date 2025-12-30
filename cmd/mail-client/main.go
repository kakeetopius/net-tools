package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/wneessen/go-mail"
)

type Options struct {
	Sender   string
	Receiver string
	Password string
	Username string
	Subject  string
	Body     string
}

func main() {
	mailOptions, err := getParams()
	if err != nil {
		if err != pflag.ErrHelp {
			fmt.Println(err)
		}
		return
	}

	err = sendMail(mailOptions)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func getParams() (*Options, error) {
	flagSet := pflag.NewFlagSet("mail-client", pflag.ContinueOnError)

	sender := flagSet.StringP("from", "f", "", "The address of the sender.")
	passwd := flagSet.StringP("pass", "p", "", "The password (app password for smarthost) to use.")
	uname := flagSet.StringP("uname", "u", "", "The username to use when sending the mail.")
	receipient := flagSet.StringP("to", "t", "", "Email Address to send mail To.")
	subject := flagSet.StringP("subject", "s", "", "The Subject of the Mail")
	body := flagSet.StringP("message", "m", "", "The message to send in e-mail")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return nil, err
	}

	// Replace all escaped new lines with actual new line character.
	*body = strings.ReplaceAll(*body, "\\n", "\n")

	return &Options{
		Sender:   *sender,
		Receiver: *receipient,
		Username: *uname,
		Password: *passwd,
		Subject:  *subject,
		Body:     *body,
	}, nil
}

func sendMail(mailOptions *Options) error {
	message := mail.NewMsg()

	if err := message.To(mailOptions.Receiver); err != nil {
		return err
	}
	if mailOptions.Username != "" {
		if err := message.FromFormat(mailOptions.Username, mailOptions.Sender); err != nil {
			return err
		}
	} else if err := message.From(mailOptions.Sender); err != nil {
		return err
	}

	message.Subject(mailOptions.Subject)
	message.SetBodyString(mail.TypeTextPlain, mailOptions.Body)

	client, err := mail.NewClient("smtp.gmail.com", mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(mailOptions.Sender), mail.WithPassword(mailOptions.Password))
	if err != nil {
		return fmt.Errorf("failed to create mail client %v", err)
	}

	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %v", err)
	}

	return nil
}
