#!/usr/bin/env bash

# Path to the data file
DATA_FILE="$HOME/.grepshot_data.json"

# Check if the data file exists
if [ ! -f "$DATA_FILE" ]; then
    echo "Data file not found. Please run grepShot first."
    exit 1
fi

# Use jq to format the JSON data for fzf
# Format: "image_path    |    extracted_text"
selected=$(jq -r 'to_entries | map("\(.key)    |    \(.value)") | .[]' "$DATA_FILE" | 
    fzf --delimiter='    |    ' \
        --with-nth=2 \
        --preview="if command -v catimg >/dev/null 2>&1; then catimg -w 100 {1}; elif command -v chafa >/dev/null 2>&1; then chafa {1}; else echo 'Install catimg or chafa for image preview'; fi" \
        --preview-window=right:60%)

# If user selected something, extract the image path and open it
if [ -n "$selected" ]; then
    image_path=$(echo "$selected" | cut -d'|' -f1 | xargs)
    if [ -f "$image_path" ]; then
        # Try to open the image with the default image viewer
        if command -v xdg-open >/dev/null 2>&1; then
            xdg-open "$image_path" &
        elif command -v open >/dev/null 2>&1; then
            open "$image_path" &
        else
            echo "Could not find a suitable program to open the image."
        fi
    else
        echo "Image file not found: $image_path"
    fi
fi