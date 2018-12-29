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
	"path/filepath"
	"syscall"

	"github.com/genuinetools/pkg/cli"
	"github.com/jessfraz/gmailfilterb0t/version"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

var (
	gmailKeyfile string
	credsDir     string

	gmailClient *gmail.Service

	debug bool
)

func main() {
	// Get home directory.
	home, err := getHome()
	if err != nil {
		logrus.Fatal(err)
	}
	credsDir = filepath.Join(home, ".gmailfilterb0t")

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

	p.FlagSet.StringVar(&gmailKeyfile, "gmail-keyfile", filepath.Join(credsDir, "gmail.json"), "Path to Gmail keyfile")

	// Set the before function.
	p.Before = func(ctx context.Context) error {
		// Set the log level.
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		// Create the Google calendar API client.
		gmailData, err := ioutil.ReadFile(gmailKeyfile)
		if err != nil {
			return fmt.Errorf("reading file %s failed: %v", gmailKeyfile, err)
		}
		gmailTokenSource, err := google.JWTConfigFromJSON(gmailData, gmail.GmailSettingsBasicScope)
		if err != nil {
			return fmt.Errorf("creating gmail token source from file %s failed: %v", gmailKeyfile, err)
		}

		// Create the Gmail client.
		gmailClient, err = gmail.New(gmailTokenSource.Client(ctx))
		if err != nil {
			return fmt.Errorf("creating gmail client failed: %v", err)
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
