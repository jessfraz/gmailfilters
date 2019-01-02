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

	"github.com/genuinetools/pkg/cli"
	"github.com/jessfraz/gmailfilters/version"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

const (
	tokenFile = "/tmp/token.json"
	gmailUser = "me"
)

var (
	credsFile  string
	jsonFile   string
	filterFile string

	api *gmail.Service

	debug           bool
	downloadFilters bool
)

func main() {
	// Create a new cli program.
	p := cli.NewProgram()
	p.Name = "gmailfilters"
	p.Description = "A tool to sync Gmail filters from a config file to your account"
	// Set the GitCommit and Version.
	p.GitCommit = version.GITCOMMIT
	p.Version = version.VERSION

	// Setup the global flags.
	p.FlagSet = flag.NewFlagSet("gmailfilters", flag.ExitOnError)
	p.FlagSet.BoolVar(&debug, "v", false, "enable debug logging")
	p.FlagSet.BoolVar(&debug, "debug", false, "enable debug logging")
	p.FlagSet.BoolVar(&downloadFilters, "d", false, "download existing filters to toml and json. (Best effort conversion)")
	p.FlagSet.BoolVar(&downloadFilters, "download", false, "download existing filters to toml")

	p.FlagSet.StringVar(&credsFile, "creds-file", os.Getenv("GMAIL_CREDENTIAL_FILE"), "Gmail credential file (or env var GMAIL_CREDENTIAL_FILE)")
	p.FlagSet.StringVar(&credsFile, "c", os.Getenv("GMAIL_CREDENTIAL_FILE"), "Gmail credential file (or env var GMAIL_CREDENTIAL_FILE)")
	p.FlagSet.StringVar(&filterFile, "f", os.Getenv("FILTER_FILE"), "Filter toml file (or env var FILTER_FILE)")

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

		if downloadFilters != false {
			fmt.Println("Downloading existing filters.")
			err := downloadExistingFilters(&labels)
			if err != nil {
				return err
			}
		}

		if len(filterFile) >= 1 {
			fmt.Printf("Decoding filters from file %s\n", filterFile)
			filters, err := decodeFile(filterFile)
			if err != nil {
				return err
			}

			// Delete our existing filters.
			fmt.Println("Deleting existing filters")
			if err := deleteExistingFilters(); err != nil {
				return err
			}

			// Convert our filters into gmail filters and add them.
			fmt.Printf("Updating %d filters, this might take a bit...\n", len(filters))
			for _, f := range filters {
				if err := f.addFilter(&labels); err != nil {
					return err
				}
			}
			fmt.Printf("Successfully updated %d filters\n", len(filters))
		} else {
			fmt.Printf("No filter file specified. Will not updating or deleting filters.")
		}
		return nil
	}

	// Run our program.
	p.Run()
}
