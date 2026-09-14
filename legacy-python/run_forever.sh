#!/bin/bash

PYTHON_SCRIPT_PATH=$1

# TMP="This variable might become useful at some point. Otherwise delete it."

while true
do
    if python3 "$PYTHON_SCRIPT_PATH"; then
        continue
    else
        exit_code=$?
        echo "Script crashed with exit code $exit_code. Restarting..." >&2
        sleep 1
    fi
done
