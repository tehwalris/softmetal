#!/bin/sh
set -e

sudo pixiecore boot \
  coreos_production_pxe.vmlinuz \
  coreos_production_pxe_image.cpio.gz \
  --cmdline='coreos.autologin coreos.first_boot=1 coreos.config_url={{ ID "../configignition.json" }}'
