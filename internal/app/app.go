package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/otiai10/gosseract/v2"
	// tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	ScreenshotsDir string
	OutputFile     string
}

func Run() error {
	if err := setupLogging(); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	config := parseFlags()
	return runOCRProcessing(config)
}

func parseFlags() *Config {
	var config Config

	// Define command line flags
	flag.StringVar(&config.ScreenshotsDir, "dir", "/home/neel/Pictures/Screenshots", "Path to screenshots directory")
	flag.StringVar(&config.ScreenshotsDir, "d", "/home/neel/Pictures/Screenshots", "Path to screenshots directory (short)")

	defaultOutput := filepath.Join(os.Getenv("HOME"), ".grepshot_data.json")
	flag.StringVar(&config.OutputFile, "output", defaultOutput, "Output file path")
	flag.StringVar(&config.OutputFile, "o", defaultOutput, "Output file path (short)")

	flag.Parse()

	// If output file is just a filename (no path), place it in HOME directory
	if !strings.Contains(config.OutputFile, "/") {
		config.OutputFile = filepath.Join(os.Getenv("HOME"), config.OutputFile)
	}

	return &config
}

func setupLogging() error {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("grepshot_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}

func runOCRProcessing(config *Config) error {
	// Check if directory exists
	info, err := os.Stat(config.ScreenshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Directory does not exist: %s\n", config.ScreenshotsDir)
			return err
		}
		log.Printf("Error accessing directory: %s\n", err)
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", config.ScreenshotsDir)
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
	filepath.WalkDir(config.ScreenshotsDir, func(path string, d fs.DirEntry, err error) error {
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

	log.Printf("\nFound %d image files in %s\n", count, config.ScreenshotsDir)

	// Display the number of images with extracted text
	log.Printf("\nSuccessfully extracted text from %d images\n", len(imageTextMap))

	err = writeImageTextToFile(imageTextMap, config.OutputFile)
	if err != nil {
		log.Printf("Error writing to file: %v\n", err)
		return err
	} else {
		log.Printf("Image text data written to %s\n", config.OutputFile)
	}

	return nil
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
