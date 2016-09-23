#!/bin/bash
set -eu

#Fingerprints, versions and SHAs must not be base64 encoded for them to be compatible with the gocli
#This script will convert any such occurrences of base64 encoded to their unencoded form.

#Arguments: the directory of the release you want to convert.

type base64 #Script assumes that base64 is on the $PATH
release_directory="$1"
cd "$release_directory"

#find all files with base64'd sha, fingerprint, or version
grep -Ril 'binary |-' . | while read -r fileToConvert; do
#find the base64'd string and decode it, then replace the encoded string in the file
#only acts on lines that contain string '!binary |-'
  awk -f <(cat <<EOF
  /!binary \|-/ {
    split(\$0, key, ":")
    getline
    base64var=\$0
    gsub(/ /, "", base64var)
    "echo " base64var " | base64 -D" | getline decodedVar
    print key[1]": "decodedVar
    next
  }
  {
    print \$0
  }
EOF) "$fileToConvert" > "$fileToConvert.bak"

  mv "$fileToConvert.bak" "$fileToConvert"
done



