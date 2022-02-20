#!/usr/bin/env bash

set -euo pipefail

version="${1:-}"
if [[ -z "$version" ]]; then
	# jq -r '.[0].tag_name'
	version="$(curl -s -H 'Accept: application/vnd.github.v3+json' \
			'https://api.github.com/repos/libusb/libusb/releases?per_page=1' \
			| sed -n 's/ *"tag_name": *"v\([^"]*\)",$/\1/p'
	)"
fi

if [[ "$version" = v* ]]; then
	version="${version#v}"
fi

archive=libusb-${version}.tar.bz2

curl -L -o "$archive" "https://github.com/libusb/libusb/releases/download/v$version/$archive"

rm -Rf libusb-${version}
tar xjf "$archive"

rm -Rf libusb.orig
if [[ -d libusb ]]; then
	mv libusb libusb.orig
fi
mv libusb-${version} libusb
rm -Rf libusb.orig "$archive"
