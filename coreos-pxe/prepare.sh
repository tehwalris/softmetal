#!/bin/sh
set -e

ct -in-file ./config/container-linux-config -out-file ./config/ignition.json
wget https://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe.vmlinuz -O coreos_production_pxe.vmlinuz
wget https://alpha.release.core-os.net/amd64-usr/current/coreos_production_pxe_image.cpio.gz -O coreos_production_pxe_image.cpio.gz
