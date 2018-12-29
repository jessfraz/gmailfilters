# gmailfilterb0t

[![Travis CI](https://img.shields.io/travis/jessfraz/gmailfilterb0t.svg?style=for-the-badge)](https://travis-ci.org/jessfraz/gmailfilterb0t)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/gmailfilterb0t)
[![Github All Releases](https://img.shields.io/github/downloads/jessfraz/gmailfilterb0t/total.svg?style=for-the-badge)](https://github.com/jessfraz/gmailfilterb0t/releases)

{DESCRIPTION}

* [Installation](README.md#installation)
   * [Binaries](README.md#binaries)
   * [Via Go](README.md#via-go)
* [Usage](README.md#usage)
* [Setup](README.md#setup)
   * [Gmail](README.md#gmail)

## Installation

#### Binaries

For installation instructions from binaries please visit the [Releases Page](https://github.com/jessfraz/gmailfilterb0t/releases).

#### Via Go

```console
$ go get github.com/jessfraz/gmailfilterb0t
```

## Usage

```console
$ gmailfilterb0t -h
gmailfilterb0t -  A bot to sync gmail filters from a config file to your account.

Usage: gmailfilterb0t <command>

Flags:

  -d, --debug      enable debug logging (default: false)
  --gmail-keyfile  Path to GMail keyfile (default: /home/jessie/.gmailfilterb0t/gmail.json)

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
