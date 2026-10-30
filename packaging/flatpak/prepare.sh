#!/bin/sh
# Prepares the flatpak build: fills build.sh's @NAME@/@VERSION@ from
# mygo.json and gives the metainfo its <releases>, from CHANGELOG.md with
# each version dated by its tag, into _build/ (gitignored). The manifest
# builds from there, so nothing derived is committed and a release needs
# no step of its own: the changelog a release already writes is all
# there is to update.
#
#	Run from the repository root:
#	  packaging/flatpak/prepare.sh
#	  cd packaging/flatpak && flatpak-builder --user --install-deps-from=flathub \
#	    --force-clean --repo=repo --install build dev.egoist.cloudmusic.yaml
set -e
cd "$(dirname "$0")/../.."

go run ./packaging/flatpak/gen packaging/flatpak
