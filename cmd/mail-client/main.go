package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kakeetopius/net-tools/internal/util"
	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/wneessen/go-mail"
)

type MailOptions struct {
	Sender      string
	Receiver    string
	Password    string
	Username    string
	Subject     string
	Body        string
	interactive bool
}

func main() {
	mailOptions, err := getOptions()
	if err != nil {
		if err != pflag.ErrHelp {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return
	}

	err = sendMail(&mailOptions)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}
}

func getOptions() (MailOptions, error) {
	flagSet := pflag.NewFlagSet("mail-client", pflag.ContinueOnError)

	flagSet.SortFlags = false
	flagSet.StringP("from", "f", "", "The address of the sender.")
	flagSet.StringP("to", "t", "", "Email Address to send mail To.")
	flagSet.StringP("username", "u", "", "The username to use when sending the mail.")
	flagSet.StringP("password", "p", "", "The password (app password for smarthost) to use.")
	flagSet.StringP("subject", "s", "", "The Subject of the Mail")
	flagSet.StringP("message", "m", "", "The message to send in e-mail")

	configFile := flagSet.StringP("config", "c", "", "The config file to use. (default is a file called mail.toml in the user's config directory)")
	interactive := flagSet.BoolP("interactive", "i", false, "Prompt the user for the mail options")

	flagSet.Usage = util.UsageFunc("mail-client", "", flagSet.FlagUsages(), "Send e-mails directly from your terminal.")
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		return MailOptions{}, err
	}

	config, err := NewConfig(*configFile)
	if err != nil {
		return MailOptions{}, err
	}

	config.BindPFlags(flagSet)

	opts := MailOptions{
		Sender:      config.GetString("from"),
		Receiver:    config.GetString("to"),
		Password:    config.GetString("password"),
		Username:    config.GetString("username"),
		Subject:     config.GetString("subject"),
		Body:        config.GetString("message"),
		interactive: config.GetBool("interactive"),
	}

	if *interactive {
		err := opts.AddMissingFieldsInteractively()
		if err != nil {
			return MailOptions{}, err
		}
	}

	return opts, nil
}

func sendMail(mailOptions *MailOptions) error {
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
	body := strings.ReplaceAll(mailOptions.Body, "\\n", "\n") // Replace all new line escape sequences given in the string with actual new line character.

	message.SetBodyString(mail.TypeTextPlain, body)

	client, err := mail.NewClient("smtp.gmail.com", mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(mailOptions.Sender), mail.WithPassword(mailOptions.Password))
	if err != nil {
		return fmt.Errorf("failed to create mail client %v", err)
	}

	if mailOptions.interactive {
		confirm, err := pterm.DefaultInteractiveConfirm.Show("Proceed to send mail")
		if err != nil {
			return err
		}

		if !confirm {
			return nil
		}

		spinner, err := pterm.DefaultSpinner.Start("Sending mail")
		if err != nil {
			return err
		}
		defer spinner.Stop()
	}
	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %v", err)
	}

	return nil
}

func NewConfig(configFile string) (*viper.Viper, error) {
	config := viper.New()
	if configFile != "" {
		config.SetConfigFile(configFile)
	} else {
		config.SetConfigName("mail")
		config.SetConfigType("toml")
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		config.AddConfigPath(configDir)
		config.AddConfigPath(".")
	}

	confErr := config.ReadInConfig()
	if confErr != nil {
		if _, ok := confErr.(viper.ConfigFileNotFoundError); !ok {
			// only return error if an error occured that is not file not found
			return nil, confErr
		}
	}
	return config, nil
}

func (o *MailOptions) AddMissingFieldsInteractively() error {
	if o.Sender == "" {
		result, err := pterm.DefaultInteractiveTextInput.Show("Enter the sender mail address")
		if err != nil {
			return err
		}
		o.Sender = result
	}

	if o.Receiver == "" {
		result, err := pterm.DefaultInteractiveTextInput.Show("Enter the receiver mail address")
		if err != nil {
			return err
		}
		o.Receiver = result
	}

	if o.Username == "" {
		result, err := pterm.DefaultInteractiveTextInput.Show("Enter the username(Leave blank for default)")
		if err != nil {
			return err
		}
		o.Username = result
	}

	if o.Password == "" {
		result, err := pterm.DefaultInteractiveTextInput.WithMask("*").Show("Enter the app password")
		if err != nil {
			return err
		}
		o.Password = result
	}

	if o.Subject == "" {
		subj, err := pterm.DefaultInteractiveTextInput.Show("Type the Mail Subject")
		if err != nil {
			return err
		}
		o.Subject = subj
	}

	if o.Body == "" {
		body, err := pterm.DefaultInteractiveTextInput.WithMultiLine().Show("Type the message")
		if err != nil {
			return err
		}
		o.Body = body
	}

	return nil
}
