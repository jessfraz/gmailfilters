# gmailfilters

[![Travis CI](https://img.shields.io/travis/jessfraz/gmailfilters.svg?style=for-the-badge)](https://travis-ci.org/jessfraz/gmailfilters)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/gmailfilters)
[![Github All Releases](https://img.shields.io/github/downloads/jessfraz/gmailfilters/total.svg?style=for-the-badge)](https://github.com/jessfraz/gmailfilters/releases)

A tool to sync Gmail filters from a config file to your account.

* [Installation](README.md#installation)
   * [Binaries](README.md#binaries)
   * [Via Go](README.md#via-go)
* [Usage](README.md#usage)
* [Setup](README.md#setup)
   * [Gmail](README.md#gmail)

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/jessfraz/gmailfilters/releases).

#### Via Go

```console
$ go get github.com/jessfraz/gmailfilters
```

## Usage

```console
$ gmailfilters -h
gmailfilters -  A tool to sync Gmail filters from a config file to your account.

Usage: gmailfilters <command>

Flags:

  -d, --debug       enable debug logging (default: false)
  -f, --creds-file  Gmail credential file (or env var GMAIL_CREDENTIAL_FILE) (default: <none>)

Commands:

  version  Show the version information.
```

## Setup

### Gmail

1. Enable the API: To get started using Gmail API, you need to 
    first create a project in the 
    [Google API Console](https://console.developers.google.com),
    enable the API, and create credentials.

    Follow the instructions 
    [for step enabling the API here](https://developers.google.com/gmail/api/quickstart/go).
