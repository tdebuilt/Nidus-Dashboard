#!/bin/sh
# Ensure /data is writable by nidus user when mounted as a Docker volume.
# Volumes created by Docker default to root ownership.
if [ "$(id -u)" = "0" ]; then
  chown -R nidus:nidus /data
  exec su-exec nidus ./nidus "$@"
else
  exec ./nidus "$@"
fi
