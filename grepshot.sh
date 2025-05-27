#!/usr/bin/env bash

# grepshot-finder.sh - Interactive file finder with image preview using kitty

set -euo pipefail

# Configuration
GREPSHOT_DATA="$HOME/.grepshot_data.json"
PREVIEW_WIDTH="60%"
DELIMITER=" => "

# Check if required tools are available
for cmd in jq fzf kitty bat; do
    if ! command -v "$cmd" &> /dev/null; then
        echo "Error: $cmd is not installed" >&2
        exit 1
    fi
done

# Check if data file exists
if [[ ! -f "$GREPSHOT_DATA" ]]; then
    echo "Error: $GREPSHOT_DATA not found" >&2
    exit 1
fi

# Preview function for fzf
preview_file() {
    local input="$1"
    local file
    
    # Extract filename from the formatted line
    file=$(echo "$input" | awk -F"$DELIMITER" '{print $2}')
    
    # Check if file exists
    if [[ ! -f "$file" ]]; then
        echo "File not found: $file"
        return
    fi
    
    # Check if it's an image and preview accordingly
    if file --mime-type "$file" | grep -qF "image/"; then
        # Use kitty image protocol for images
        kitty icat \
            --clear \
            --transfer-mode=memory \
            --stdin=no \
            --place="${FZF_PREVIEW_COLUMNS}x${FZF_PREVIEW_LINES}@0x0" \
            "$file"
    else
        # Use bat for syntax highlighting, fallback to cat
        bat --color=always "$file" 2>/dev/null || cat "$file"
    fi
}

# Export the function so fzf can use it
export -f preview_file
export DELIMITER

# Main execution
main() {
    local selected_file
    
    # Process the JSON data and run fzf
    selected_file=$(
        cat "$GREPSHOT_DATA" \
        | jq -r 'to_entries[] | (.value | gsub("\n"; " ")) + "... => " + .key' \
        | fzf \
            --delimiter="$DELIMITER" \
            --with-nth=1 \
            --preview 'preview_file {}' \
            --preview-window="right:$PREVIEW_WIDTH" \
        | awk -F"$DELIMITER" '{print $2}'
    )
    
    # Output the selected file path
    if [[ -n "$selected_file" ]]; then
        echo "$selected_file"
    fi
}

# Run main function
main "$@"
