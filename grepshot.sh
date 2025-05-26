#!/usr/bin/env bash

# Directory containing screenshots
SCREENSHOTS_DIR="$HOME/Pictures/Screenshots"

# Check if tesseract is installed
if ! command -v tesseract &> /dev/null; then
    echo "Error: tesseract is not installed" >&2
    exit 1
fi

# Check if screenshots directory exists
if [ ! -d "$SCREENSHOTS_DIR" ]; then
    echo "Error: Screenshots directory does not exist: $SCREENSHOTS_DIR" >&2
    exit 1
fi

# Start JSON output
echo "{"

first=true
# Process each image file in the screenshots directory
for img in "$SCREENSHOTS_DIR"/*.{png,jpg,jpeg,gif,bmp,tiff}; do
    # Skip if no files match the pattern
    [ ! -f "$img" ] && continue
    
    # Add comma separator for all but first entry
    if [ "$first" = true ]; then
        first=false
    else
        echo ","
    fi
    
    # Run tesseract OCR and capture output
    ocr_output=$(tesseract "$img" stdout 2>/dev/null | tr '\n' ' ' | sed 's/"/\\"/g' | sed 's/\t/ /g')
    
    # Output JSON key-value pair
    echo -n "  \"$img\": \"$ocr_output\""
done

# End JSON output
echo ""
echo "}"


