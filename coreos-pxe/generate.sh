#!/bin/sh
set -e

if [ -z "$1" ] || [ -z "$2" ] || [ -z "$3" ]; then
  echo "Missing arguments. Usage: ./generate.sh [softmetalHost] [softmetalHTTPPort] [softmetalGRPCPort]"
  exit 1
fi

sed "s/{{softmetalHost}}/$1/g; s/{{softmetalHTTPPort}}/$2/g; s/{{softmetalGRPCPort}}/$3/g" \
  ./config/container-linux-config | ct -out-file ./config/ignition.json
