package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SearchConfig holds the configuration for the file search
type SearchConfig struct {
	RootPath     string
	SearchTerm   string
	MaxWorkers   int
	IgnoreErrors bool
}

// SearchResult represents a found file and any associated error
type SearchResult struct {
	Path  string
	Error error
}

// FileSearcher handles the concurrent file searching logic
type FileSearcher struct {
	config  SearchConfig
	results chan SearchResult
	wg      sync.WaitGroup
}

// NewFileSearcher creates a new FileSearcher instance
func NewFileSearcher(config SearchConfig) *FileSearcher {
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = 10 // Default number of workers
	}
	return &FileSearcher{
		config:  config,
		results: make(chan SearchResult),
	}
}

// Search initiates the concurrent file search
func (fs *FileSearcher) Search(ctx context.Context) (<-chan SearchResult, error) {
	if fs.config.RootPath == "" {
		return nil, fmt.Errorf("root path cannot be empty")
	}

	// Verify the root path exists
	if _, err := os.Stat(fs.config.RootPath); err != nil {
		return nil, fmt.Errorf("invalid root path: %w", err)
	}

	fs.wg.Add(1)
	go fs.searchDir(ctx, fs.config.RootPath)

	// Start a goroutine to close results channel when search is complete
	go func() {
		fs.wg.Wait()
		close(fs.results)
	}()

	return fs.results, nil
}

// searchDir recursively searches directories for matching files
func (fs *FileSearcher) searchDir(ctx context.Context, dir string) {
	defer fs.wg.Done()

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return
	default:
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if !fs.config.IgnoreErrors {
			fs.results <- SearchResult{Error: fmt.Errorf("error reading directory %s: %w", dir, err)}
		}
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		if strings.Contains(entry.Name(), fs.config.SearchTerm) {
			fs.results <- SearchResult{Path: fullPath}
		}

		if entry.IsDir() {
			fs.wg.Add(1)
			go fs.searchDir(ctx, fullPath)
		}
	}
}

// ProcessResults collects and processes search results
func ProcessResults(results <-chan SearchResult) ([]string, []error) {
	var matches []string
	var errors []error

	for result := range results {
		if result.Error != nil {
			errors = append(errors, result.Error)
			continue
		}
		matches = append(matches, result.Path)
	}

	return matches, errors
}

func main() {
	// Create a context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure the search
	config := SearchConfig{
		RootPath:     "/Users/sohambose/Desktop/Edima/Go_prac/",
		SearchTerm:   "README.md",
		MaxWorkers:   5,
		IgnoreErrors: false,
	}

	// Create and start the file searcher
	searcher := NewFileSearcher(config)
	results, err := searcher.Search(ctx)
	if err != nil {
		fmt.Printf("Error starting search: %v\n", err)
		return
	}

	// Process the results
	matches, errors := ProcessResults(results)

	// Print results
	fmt.Printf("Found %d matches:\n", len(matches))
	for _, match := range matches {
		fmt.Printf("- %s\n", match)
	}

	if len(errors) > 0 {
		fmt.Printf("\nEncountered %d errors:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("- %v\n", err)
		}
	}
}
