# grepShot

A tool for extracting text from screenshots and searching through them using fzf.

## Features

- Extracts text from screenshot images using OCR (Tesseract)
- Stores image paths and extracted text in a JSON file
- Provides a script to search through screenshots based on text content using fzf

## Requirements

- Go
- Tesseract OCR
- jq (for the search script)
- fzf (for the search interface)
- catimg or chafa (optional, for image previews in the search interface)

## Usage

1. Build and run the main program to extract text from screenshots:
   ```
   go build
   ./grepShot
   ```

2. Make the search script executable:
   ```
   chmod +x grepshot.sh
   ```

3. Run the search script to find screenshots by text content:
   ```
   ./grepshot.sh
   ```

4. Type any text to filter screenshots by their content. Press Enter to open the selected screenshot.

## Notes

- By default, the program scans up to 20 images in `/home/neel/Pictures/Screenshots`
- The extracted text data is stored in `~/.grepshot_data.json`