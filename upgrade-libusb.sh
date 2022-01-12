#!/usr/bin/env bash

set -euo pipefail

if [[ -z "${1:-}" ]]; then
	echo "usage: $0 <version>"
	echo "See https://github.com/libusb/libusb/releases"
	exit 1
fi
version="$1"
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
