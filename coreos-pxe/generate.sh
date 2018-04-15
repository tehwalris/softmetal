#!/bin/sh
set -e

ct -in-file ./config/container-linux-config -out-file ./config/ignition.json
