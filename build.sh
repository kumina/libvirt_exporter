#!/bin/sh
docker run --rm -v $(pwd):/usr/src/libvirt_exporter -w /usr/src/libvirt_exporter golang:1.14-alpine3.11 dockerscripts/build_static.sh
