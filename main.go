package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: scan <domain>")
		os.Exit(1)
	}

	domain := os.Args[1]
	outputDir := filepath.Join(".", domain)

	// Create the output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Println("Error creating output directory:", err)
		os.Exit(1)
	}

	// Subfinder
	subfinderCmd := exec.Command("subfinder", "-d", domain, "-silent")
	subfinderCmd.Dir = outputDir
	subfinderOutput, _ := subfinderCmd.Output()
	writeToFile(filepath.Join(outputDir, "Subfinder.txt"), subfinderOutput)

	// Assetfinder
	assetfinderCmd := exec.Command("assetfinder", "-subs-only", domain)
	assetfinderCmd.Dir = outputDir
	assetfinderOutput, _ := assetfinderCmd.Output()
	writeToFile(filepath.Join(outputDir, "Assetfinder.txt"), assetfinderOutput)

	// Amass
	amassCmd := exec.Command("amass", "enum", "-passive", "-d", domain)
	amassCmd.Dir = outputDir
	amassOutput, _ := amassCmd.Output()
	writeToFile(filepath.Join(outputDir, "amass.txt"), amassOutput)

	// Combine and deduplicate results
	combineAndDeduplicate(outputDir)

	// Read input from "combined.txt" file
	combinedFile := filepath.Join(outputDir, "combined.txt")
	hosts, err := readLines(combinedFile)
	if err != nil {
		fmt.Println("Error reading combined file:", err)
		os.Exit(1)
	}

	// Httpx
	httpxCmd := exec.Command("httpx", "-silent")
	httpxCmd.Dir = outputDir

	// Append hosts as arguments to the httpx command
	httpxCmd.Args = append(httpxCmd.Args, hosts...)

	// Execute httpx command
	httpxOutput, err := httpxCmd.Output()
	if err != nil {
		fmt.Println("Error running httpx command:", err)
		os.Exit(1)
	}

	// Write httpx output to a file
	writeToFile(filepath.Join(outputDir, "alive.txt"), httpxOutput)

	fmt.Println("Scan completed. Results saved to alive.txt.")
}

// readLines reads a file into a slice of strings, one string per line.
func readLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeToFile writes data to a file.
func writeToFile(filename string, data []byte) error {
	return ioutil.WriteFile(filename, data, 0644)
}

func combineAndDeduplicate(outputDir string) {
	files, err := filepath.Glob(filepath.Join(outputDir, "*.txt"))
	if err != nil {
		fmt.Println("Error listing files:", err)
		os.Exit(1)
	}

	var combinedData []byte
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(1)
		}
		combinedData = append(combinedData, data...)
	}

	// Remove duplicates
	uniqueLines := removeDuplicates(strings.Split(string(combinedData), "\n"))

	// Write to a new file
	writeToFile(filepath.Join(outputDir, "combined.txt"), []byte(strings.Join(uniqueLines, "\n")))
}

func removeDuplicates(lines []string) []string {
	uniqueLines := make(map[string]struct{})
	for _, line := range lines {
		if line != "" {
			uniqueLines[line] = struct{}{}
		}
	}

	var result []string
	for line := range uniqueLines {
		result = append(result, line)
	}

	return result
}
