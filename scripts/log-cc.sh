#!/bin/sh
# Copyright (c) 2024 Fabio Massaioli
# Use of this source code is governed by the MIT license
# that can be found in the LICENSE file.
#
# A clang wrapper script that logs compiler invocations

set -e
exec "$(dirname "$0")/log.sh" clang "$@"
