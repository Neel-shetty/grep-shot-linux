package app

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/neel/grepShot/internal/db"
)

type Config struct {
	Pattern       string
	Directory     string
	CaseSensitive bool
	WholeWord     bool
	LineNumbers   bool
	Count         bool
	Recursive     bool
	FilePattern   string
	Context       int
	MaxDepth      int
	Exclude       string
	Interactive   bool
	LogLevel      string
	Output        string
}

func Run() error {
	config := parseFlags()

	if err := setupLogging(config.LogLevel); err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}

	if config.Interactive {
		return runInteractiveMode(config)
	}

	return runSingleSearch(config)
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.Pattern, "p", "", "Pattern to search for (required)")
	flag.StringVar(&config.Directory, "d", ".", "Directory to search in")
	flag.BoolVar(&config.CaseSensitive, "c", false, "Case sensitive search")
	flag.BoolVar(&config.WholeWord, "w", false, "Match whole words only")
	flag.BoolVar(&config.LineNumbers, "n", false, "Show line numbers")
	flag.BoolVar(&config.Count, "count", false, "Show only count of matches")
	flag.BoolVar(&config.Recursive, "r", true, "Recursive search")
	flag.StringVar(&config.FilePattern, "f", "", "File pattern to match")
	flag.IntVar(&config.Context, "A", 0, "Show N lines after match")
	flag.IntVar(&config.MaxDepth, "depth", -1, "Maximum directory depth")
	flag.StringVar(&config.Exclude, "exclude", "", "Exclude pattern")
	flag.BoolVar(&config.Interactive, "i", false, "Interactive mode")
	flag.StringVar(&config.LogLevel, "log", "info", "Log level (debug, info, warn, error)")
	flag.StringVar(&config.Output, "o", "", "Output file")

	flag.Parse()

	if !config.Interactive && config.Pattern == "" {
		fmt.Println("Pattern is required. Use -p flag or -i for interactive mode.")
		flag.Usage()
		os.Exit(1)
	}

	return config
}

func setupLogging(level string) error {
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

func runInteractiveMode(config *Config) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("grepShot> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "exit" || input == "quit" {
			break
		}

		if input == "" {
			continue
		}

		config.Pattern = input
		if err := runSingleSearch(config); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}

func runSingleSearch(config *Config) error {
	matches, err := performSearch(config)
	if err != nil {
		return err
	}

	if err := db.SaveSearchHistory(config.Pattern, config.Directory, len(matches)); err != nil {
		log.Printf("Failed to save search history: %v", err)
	}

	return displayResults(matches, config)
}

func performSearch(config *Config) ([]SearchResult, error) {
	var results []SearchResult

	err := filepath.Walk(config.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if shouldSkipDirectory(path, config) {
				return filepath.SkipDir
			}
			return nil
		}

		if !shouldSearchFile(path, config) {
			return nil
		}

		matches, err := searchInFile(path, config)
		if err != nil {
			log.Printf("Error searching file %s: %v", path, err)
			return nil
		}

		results = append(results, matches...)
		return nil
	})

	return results, err
}

func shouldSkipDirectory(path string, config *Config) bool {
	if config.MaxDepth >= 0 {
		depth := strings.Count(strings.TrimPrefix(path, config.Directory), string(os.PathSeparator))
		if depth > config.MaxDepth {
			return true
		}
	}

	if config.Exclude != "" {
		matched, _ := regexp.MatchString(config.Exclude, filepath.Base(path))
		return matched
	}

	return false
}

func shouldSearchFile(path string, config *Config) bool {
	if config.FilePattern != "" {
		matched, _ := regexp.MatchString(config.FilePattern, filepath.Base(path))
		if !matched {
			return false
		}
	}

	return true
}

func searchInFile(filePath string, config *Config) ([]SearchResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []SearchResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	pattern := config.Pattern
	if !config.CaseSensitive {
		pattern = "(?i)" + pattern
	}
	if config.WholeWord {
		pattern = "\\b" + pattern + "\\b"
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if regex.MatchString(line) {
			result := SearchResult{
				File:    filePath,
				Line:    lineNum,
				Content: line,
				Match:   regex.FindString(line),
			}
			results = append(results, result)
		}
	}

	return results, scanner.Err()
}

func displayResults(results []SearchResult, config *Config) error {
	if config.Count {
		fmt.Printf("Total matches: %d\n", len(results))
		return nil
	}

	var output strings.Builder

	for _, result := range results {
		if config.LineNumbers {
			output.WriteString(fmt.Sprintf("%s:%d:%s\n", result.File, result.Line, result.Content))
		} else {
			output.WriteString(fmt.Sprintf("%s:%s\n", result.File, result.Content))
		}
	}

	if config.Output != "" {
		return os.WriteFile(config.Output, []byte(output.String()), 0644)
	}

	fmt.Print(output.String())
	return nil
}

type SearchResult struct {
	File    string
	Line    int
	Content string
	Match   string
}
