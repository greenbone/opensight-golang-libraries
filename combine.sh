#!/bin/bash

# Directory to search for .go files
SEARCH_DIR="/Users/johanneshollerer/GolandProjects/greenbone/opensight-golang-libraries/pkg/query/"

# File to store the combined output
OUTPUT_FILE="combined.go"

# Check if the output file already exists and remove it
if [ -f "$OUTPUT_FILE" ]; then
    rm "$OUTPUT_FILE"
fi

# Find all .go files and append them to the output file
find "$SEARCH_DIR" -name '*.go' -exec cat {} + >> "$OUTPUT_FILE"

echo "All .go files have been combined into $OUTPUT_FILE"
