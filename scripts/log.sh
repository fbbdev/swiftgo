#!/bin/sh
# Copyright (c) 2024 Fabio Massaioli
# Use of this source code is governed by the MIT license
# that can be found in the LICENSE file.
#
# A go toolexec script that logs tool invocations

set -e

lockfile="$(dirname "$0")/log.lock"
logfile="$(dirname "$0")/toolexec.log"

while ! shlock -f "$lockfile" -p "$$"; do sleep 0.1; done

echo "Package:" "$TOOLEXEC_IMPORTPATH" >> "$logfile"
echo "Tool:" "'""$1""'" >> "$logfile"
echo "$@" >> "$logfile"
echo "---------------" >> "$logfile"

rm "$lockfile"

exec "$@"
