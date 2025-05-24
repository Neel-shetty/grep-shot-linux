package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/otiai10/gosseract/v2"
)

func main() {
	// Set up logging to both console and file
	logFile, err := setupLogging()
	if err != nil {
		fmt.Printf("Error setting up logging: %v\n", err)
		return
	}
	defer logFile.Close()

	// Define the screenshots directory path
	screenshotsDir := "/home/neel/Pictures/Screenshots"

	// Check if directory exists
	info, err := os.Stat(screenshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Directory does not exist: %s\n", screenshotsDir)
			return
		}
		log.Printf("Error accessing directory: %s\n", err)
		return
	}

	if !info.IsDir() {
		log.Printf("%s is not a directory\n", screenshotsDir)
		return
	}

	// Common image file extensions
	imageExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}

	// Create a thread-safe hashmap to store filepath and extracted text
	var mu sync.Mutex
	imageTextMap := make(map[string]string)

	// Collect image paths first
	var imagePaths []string
	var count int

	// Walk through the directory and collect image file paths
	filepath.WalkDir(screenshotsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if imageExtensions[ext] {
				imagePaths = append(imagePaths, path)
				log.Println(d.Name())
				count++
			}
		}
		return nil
	})

	// Create a worker pool for parallel processing
	numWorkers := runtime.NumCPU() // Use number of available CPU cores
	var wg sync.WaitGroup

	// Process images in parallel
	log.Println("\nExtracting text from images in parallel...")
	log.Printf("Using %d worker goroutines (based on CPU count)\n", numWorkers)

	// Create a channel to distribute work
	jobs := make(chan string, len(imagePaths))

	// Start worker goroutines
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Each worker gets its own Tesseract client
			client := gosseract.NewClient()
			defer client.Close()

			for path := range jobs {
				// Extract text from image using Tesseract
				err := client.SetImage(path)
				if err != nil {
					log.Printf("Worker %d: Error setting image %s: %v\n", workerID, path, err)
					continue
				}

				text, err := client.Text()
				if err != nil {
					log.Printf("Worker %d: Error extracting text from %s: %v\n", workerID, path, err)
				} else {
					// Store the extracted text in the hashmap with mutex protection
					mu.Lock()
					imageTextMap[path] = text
					name := filepath.Base(path)
					mu.Unlock()
					log.Printf("Worker %d: Extracted %d characters of text from %s\n", workerID, len(text), name)
				}
			}
		}(w)
	}

	// Send jobs to the workers
	for _, path := range imagePaths {
		jobs <- path
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()

	log.Printf("\nFound %d image files in %s\n", count, screenshotsDir)

	// Display the number of images with extracted text
	log.Printf("\nSuccessfully extracted text from %d images\n", len(imageTextMap))

	// Write the image path and text to a file
	outputFile := filepath.Join(os.Getenv("HOME"), ".grepshot_data.json")
	err = writeImageTextToFile(imageTextMap, outputFile)
	if err != nil {
		log.Printf("Error writing to file: %v\n", err)
	} else {
		log.Printf("Image text data written to %s\n", outputFile)
	}
}

// setupLogging configures logging to both console and file
func setupLogging() (*os.File, error) {
	// Create logs directory if it doesn't exist
	logsDir := filepath.Join(".", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file with timestamp in name
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join(logsDir, fmt.Sprintf("grepshot_%s.log", timestamp))

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// Set up multi-writer to log to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)

	// Configure log format to include timestamp
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	log.Printf("Log file created at: %s\n", logFilePath)
	return logFile, nil
}

// writeImageTextToFile writes the image path and extracted text to a JSON file
func writeImageTextToFile(imageTextMap map[string]string, outputPath string) error {
	// Create the output file
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the map as JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(imageTextMap)
}
