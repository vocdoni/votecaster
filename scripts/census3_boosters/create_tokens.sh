#!/bin/bash

# Check if jq is installed
if ! command -v jq &>/dev/null; then
    echo "Error: jq is not installed. Please install jq to proceed."
    exit 1
fi

# Check if the input file is provided as an argument
if [ -z "$1" ]; then
    echo "Error: No JSON file path provided."
    echo "Usage: $0 /path/to/input.json"
    exit 1
fi

# Input JSON file path
INPUT_FILE="$1"

# Check if the file exists
if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: File '$INPUT_FILE' not found."
    exit 1
fi

# Extract boosters_tokens array length
LENGTH=$(jq '.boosters_tokens | length' $INPUT_FILE)

# Loop through each item in the boosters_tokens list
for ((i = 0; i < $LENGTH; i++)); do
    # Extract each item as a JSON string
    ITEM=$(jq -c ".boosters_tokens[$i]" $INPUT_FILE)

    # Send POST request using wget
    wget --quiet \
        --method POST \
        --header 'Accept: */*' \
        --header 'Content-Type: application/json' \
        --body-data "$ITEM" \
        --output-document \
        - https://census3-votecaster.vocdoni.net/api/tokens
done
