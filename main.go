package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"syscall"

	"github.com/genuinetools/pkg/cli"
	"github.com/jessfraz/gmailfilterb0t/version"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	tokenFile = "/tmp/token.json"
)

var (
	credsFile string

	api *gmail.Service

	debug bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "gmailfilterb0t"
	p.Description = "A bot to sync gmail filters from a config file to your account"
	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("gmailfilterb0t", flag.ExitOnError)
	p.FlagSet.BoolVar(&debug, "d", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")

	p.FlagSet.StringVar(&credsFile, "creds-file", os.Getenv("GMAIL_CREDENTIAL_FILE"), "Gmail credential file (or env var GMAIL_CREDENTIAL_FILE)")
	p.FlagSet.StringVar(&credsFile, "f", os.Getenv("GMAIL_CREDENTIAL_FILE"), "Gmail credential file (or env var GMAIL_CREDENTIAL_FILE)")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if len(credsFile) < 1 {
			return errors.New("Gmail credential file cannot be empty")
		}

		// Make sure the file exists.
		if _, err := os.Stat(credsFile); os.IsNotExist(err) {
			return fmt.Errorf("Credential file %s does not exist", credsFile)
		}

		// Read the credentials file.
		b, err := ioutil.ReadFile(credsFile)
		if err != nil {
			logrus.Fatalf("reading client secret file %s failed: %v", credsFile, err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b,
			// Read, modify, and manage your settings.
			gmail.GmailSettingsBasicScope)
		if err != nil {
			logrus.Fatalf("parsing client secret file to config failed: %v", err)
		}

		// Get the client from the config.
		client, err := getClient(ctx, config)
		if err != nil {
			logrus.Fatalf("creating client failed: %v", err)
		}

		// Create the service for the Gmail client.
		api, err = gmail.New(client)
		if err != nil {
			logrus.Fatalf("creating Gmail client failed: %v", err)
		}

		return nil
	}

	p.Action = func(ctx context.Context, args []string) error {
		if len(args) < 1 {
			return errors.New("must pass a path to a gmail filter configuration file")
		}

		// On ^C, or SIGTERM handle exit.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		go func() {
			for sig := range c {
				logrus.Infof("Received %s, exiting.", sig.String())
				os.Exit(0)
			}
		}()

		return nil
	}

	// Run our program.
	p.Run()
}

func getHome() (string, error) {
	home := os.Getenv(homeKey)
	if home != "" {
		return home, nil
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.HomeDir, nil
}
