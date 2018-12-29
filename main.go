package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/genuinetools/pkg/cli"
	"github.com/jessfraz/gmailfilterb0t/version"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	tokenFile = "/tmp/token.json"
	gmailUser = "me"
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
	p.Description = "A tool to sync Gmail filters from a config file to your account"
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
			return fmt.Errorf("reading client secret file %s failed: %v", credsFile, err)
		}

		// If modifying these scopes, delete your previously saved token.json.
		config, err := google.ConfigFromJSON(b,
			// Manage labels.
			gmail.GmailLabelsScope,
			// Read, modify, and manage your settings.
			gmail.GmailSettingsBasicScope)
		if err != nil {
			return fmt.Errorf("parsing client secret file to config failed: %v", err)
		}

		// Get the client from the config.
		client, err := getClient(ctx, config)
		if err != nil {
			return fmt.Errorf("creating client failed: %v", err)
		}

		// Create the service for the Gmail client.
		api, err = gmail.New(client)
		if err != nil {
			return fmt.Errorf("creating Gmail client failed: %v", err)
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

		labels, err := getLabelMap()
		if err != nil {
			return err
		}
		fmt.Printf("labels: %#v\n\n", labels)

		filters, err := decodeFile(args[0])
		if err != nil {
			return err
		}
		fmt.Printf("file: %#v\n\n", filters)

		if err := getExistingFilters(); err != nil {
			return err
		}

		return nil
	}

	// Run our program.
	p.Run()
}

func decodeFile(file string) ([]filter, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading filter file %s failed: %v", file, err)
	}

	var ff filterfile
	if _, err := toml.Decode(string(b), &ff); err != nil {
		return nil, fmt.Errorf("decoding toml failed: %v", err)
	}

	return ff.Filters, nil
}

func getExistingFilters() error {
	// Get current filters for the user.
	l, err := api.Users.Settings.Filters.List(gmailUser).Do()
	if err != nil {
		return fmt.Errorf("listing filters failed: %v", err)
	}

	// Iterate over the filters.
	for _, f := range l.Filter {
		fmt.Printf("Action: %#v\n", f.Action)
		fmt.Printf("Criteria: %#v\n\n", f.Criteria)
	}

	return nil
}

func getLabelMap() (map[string]string, error) {
	// Get the labels for the user and map its name to its ID.
	l, err := api.Users.Labels.List(gmailUser).Do()
	if err != nil {
		return nil, fmt.Errorf("listing labels failed: %v", err)
	}

	labels := map[string]string{}
	for _, label := range l.Labels {
		labels[label.Name] = label.Id
	}

	return labels, nil
}
