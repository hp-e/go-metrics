package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func countLines(path string, wg *sync.WaitGroup, result chan<- string) {
	defer wg.Done()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		result <- fmt.Sprintf("Error reading file %s: %v", path, err)
		return
	}

	lines := strings.Split(string(content), "\n")
	codeLines, commentLines, functionCount := 0, 0, 0

	inMultiLineComment := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			commentLines++
			continue
		}

		if strings.HasPrefix(line, "/*") {
			commentLines++
			inMultiLineComment = true
			continue
		}

		if strings.HasSuffix(line, "*/") {
			commentLines++
			inMultiLineComment = false
			continue
		}

		if inMultiLineComment {
			commentLines++
			continue
		}

		if strings.HasPrefix(line, "func") {
			functionCount++
		}

		if line != "" {
			codeLines++
		}
	}

	result <- fmt.Sprintf("%s: %d lines of code, %d lines of comments, %d functions", path, codeLines, commentLines, functionCount)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s directory\n", os.Args[0])
		os.Exit(1)
	}

	dir := os.Args[1]

	var wg sync.WaitGroup
	result := make(chan string)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error accessing file %s: %v\n", path, err)
			return nil
		}

		if info.IsDir() || strings.Contains(path, "node_modules") ||
			strings.Contains(path, "obj") || strings.Contains(path, "bin") || strings.Contains(path, "nuget") {
			// if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		wg.Add(1)
		go countLines(path, &wg, result)

		return nil
	})

	go func() {
		wg.Wait()
		close(result)
	}()

	for res := range result {
		fmt.Println(res)
	}
}
