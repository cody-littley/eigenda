#!/usr/bin/env bash

# Starts the container and runs the light node.

# Expected argument 1: data directory
DATA_DIRECTORY=$1
# Expected argument 2: path to configuration file
CONFIG_FILE=$2

# Create data directory
mkdir -p $DATA_DIRECTORY
# Copy configuration file into data directory
cp "${CONFIG_FILE}" "${DATA_DIRECTORY}/config.json"

# TODO it would be neat if the compiled go binary did this parsing for us to avoid additional dependencies
PORT=$(jq .Port "${CONFIG_FILE}")
echo "Starting node on port $PORT"

# Example commands
# ./docker/run.sh ~/ws/tmp/source ~/ws/tmp/source.json
# ./docker/run.sh ~/ws/tmp/destination ~/ws/tmp/destination.json

docker run \
  --rm \
  --mount "type=bind,source=${DATA_DIRECTORY},target=/home/lnode/data" \
  --expose $PORT \
  --publish $PORT:$PORT \
  lnode \
  /home/lnode/lnode data/config.json
