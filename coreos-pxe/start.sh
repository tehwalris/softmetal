#!/bin/sh
set -e

sudo pixiecore boot \
  --dhcp-no-bind \
  coreos_production_pxe.vmlinuz \
  coreos_production_pxe_image.cpio.gz \
  --cmdline='coreos.autologin coreos.first_boot=1 coreos.config.url={{ ID "./config/ignition.json" }}'
