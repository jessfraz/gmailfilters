# gmailfilters

[![make-all](https://github.com/jessfraz/gmailfilters/workflows/make%20all/badge.svg)](https://github.com/jessfraz/gmailfilters/actions?query=workflow%3A%22make+all%22)
[![make-image](https://github.com/jessfraz/gmailfilters/workflows/make%20image/badge.svg)](https://github.com/jessfraz/gmailfilters/actions?query=workflow%3A%22make+image%22)
[![GoDoc](https://img.shields.io/badge/godoc-reference-5272B4.svg?style=for-the-badge)](https://godoc.org/github.com/jessfraz/gmailfilters)
[![Github All Releases](https://img.shields.io/github/downloads/jessfraz/gmailfilters/total.svg?style=for-the-badge)](https://github.com/jessfraz/gmailfilters/releases)

A tool to sync Gmail filters from a config file to your account.

> **NOTE:** This makes it so the single configuration file is the only way to
   add filters to your account, meaning if you add a filter via the UI and do not
   also add it in your config file, the next time you run this tool on your
   outdated config, the filter you added _only_ in the UI will be deleted.

**Table of Contents**

<!-- toc -->

- [Installation](#installation)
    + [Binaries](#binaries)
    + [Via Go](#via-go)
- [Usage](#usage)
- [Example Filter File](#example-filter-file)
- [Setup](#setup)
  * [Gmail](#gmail)

<!-- tocstop -->

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
  -e, --export      export existing filters (default: false)
  -f, --creds-file  Gmail credential file (or env var GMAIL_CREDENTIAL_FILE) (default: <none>)
  -t, --token-file  Gmail oauth token file (default: /tmp/token.json)

Commands:

  version  Show the version information.
```

## Example Filter File

```toml
[[filter]]
query = "to:your_activity@noreply.github.com"
archive = true
read = true

[[filter]]
query = "from:notifications@github.com LGTM"
label = "github/LGTM"

[[filter]]
query = """
(-to:team_mention@noreply.github.com \
(from:(notifications@github.com) AND (@jfrazelle OR @jessfraz OR to:mention@noreply.github.com OR to:author@noreply.github.com OR to:assign@noreply.github.com)))
"""
label = "github/mentions"

[[filter]]
query = """
to:team_mention@noreply.github.com \
-to:mention@noreply.github.com \
-to:author@noreply.github.com \
-to:assign@noreply.github.com
"""
label = "github/team-mention"

[[filter]]
query = """
from:notifications@github.com \
-to:team_mention@noreply.github.com \
-to:mention@noreply.github.com \
-to:author@noreply.github.com \
-to:assign@noreply.github.com
"""
archive = true

[[filter]]
query = "(from:me AND to:reply@reply.github.com)"
label = "github/mentions"

[[filter]]
query = "(from:notifications@github.com)"
label = "github"

[[filter]]
queryOr = [
"to:plans@tripit.com",
"to:receipts@concur.com",
"to:plans@concur.com",
"to:receipts@expensify.com"
]
delete = true

[[filter]]
queryOr = [
"from:notifications@docker.com",
"from:noreply@github.com",
"from:builds@travis-ci.org"
]
label = "to-be-deleted"

[[filter]]
query = "drive-shares-noreply@google.com OR (subject:\"Invitation to comment\" AND from:me ) OR from:(*@docs.google.com)"
label = "to-be-deleted"

[[filter]]
query = "(from:(-me) {filename:vcs filename:ics} has:attachment) OR (subject:(\"invitation\" OR \"accepted\" OR \"tentatively accepted\" OR \"rejected\" OR \"updated\" OR \"canceled event\" OR \"declined\") when where calendar who organizer)"
label = "to-be-deleted"

[[filter]]
query = "list:coreos-dev@googlegroups.com"
label = "Mailing Lists/coreos-dev"
archiveUnlessToMe = true

[[filter]]
queryOr = [
"list:xdg-app@lists.freedesktop.org",
"list:flatpak@lists.freedesktop.org"
]
label = "Mailing Lists/xdg-apps"
archiveUnlessToMe = true
```

## Setup

### Gmail

1. Enable the API: To get started using Gmail API, you need to 
    first create a project in the 
    [Google API Console](https://console.developers.google.com),
    enable the API, and create credentials.

    Follow the instructions 
    [for step enabling the API here](https://developers.google.com/gmail/api/quickstart/go).