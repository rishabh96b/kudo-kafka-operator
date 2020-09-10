#!/bin/bash
set -exu

kafka_repo_root="$(realpath "$(dirname "$0")")"

cd ${kafka_repo_root}/kuttl-tests
./run.sh
