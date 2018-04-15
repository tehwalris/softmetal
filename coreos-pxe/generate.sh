#!/bin/sh
set -e

if [ -z "$1" ]; then
  echo "Missing arguments. Usage: ./generate.sh [softmetalHostPort]"
  exit 1
fi

sed "s/{{softmetalHostPort}}/$1/g" ./config/container-linux-config | ct -out-file ./config/ignition.json
