#!/usr/bin/env bash

set -euo pipefail

version="${1:-}"
if [[ -z "$version" ]]; then
	# jq -r '.[0].tag_name'
	version="$(curl -s -H 'Accept: application/vnd.github.v3+json' \
		'https://api.github.com/repos/libusb/hidapi/releases?per_page=1' \
		| sed -n 's/ *"tag_name": *"hidapi-\([^"]*\)",$/\1/p'
	)"
	echo "$version" && exit 1
fi

if [[ "$version" = hidapi-* ]]; then
	version="${version#hidapi-}"
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
