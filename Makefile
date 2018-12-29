# Setup name variables for the package/tool
NAME := gmailfilters
PKG := github.com/jessfraz/$(NAME)

CGO_ENABLED := 0

# Set any default go build tags.
BUILDTAGS :=

include basic.mk

.PHONY: prebuild
prebuild:
