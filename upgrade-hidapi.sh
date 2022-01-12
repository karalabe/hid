#!/usr/bin/env bash

set -euo pipefail

if [[ -z "${1:-}" ]]; then
	echo "usage: $0 <version>"
	echo "See https://github.com/libusb/hidapi/releases"
	exit 1
fi
version="$1"
if [[ "$version" = v* ]]; then
	version="${version#v}"
fi

archive=hidapi-${version}.zip
dir=hidapi-hidapi-${version}

curl -L -o "$archive" "https://github.com/libusb/hidapi/archive/refs/tags/$archive"

rm -Rf $dir
unzip "$archive"

rm -Rf hidapi.orig
if [[ -d hidapi ]]; then
	mv hidapi hidapi.orig
fi
mv $dir hidapi
rm -Rf hidapi.orig "$archive"
