package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type TestData struct {
	Filename string `json:"filename"`
	Filepath string `json:"filepath"`
	Tests    []Test `json:"tests"`
}

type Test struct {
	Method string   `json:"method"`
	Line   int      `json:"line"`
	Block  TestCase `json:"block"`
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TestCase struct {
	Given []string `json:"given"`
	When  []string `json:"when"`
	Then  []string `json:"then"`
	Tags  []Tag    `json:"tags"`
}

func main() {
	path := flag.String("path", "", "Directory path to parse Swift files")
	flag.Parse()

	if *path == "" {
		printHelpMenu()
		return
	}

	testData, orphanTestData, err := parseSwiftFiles(*path)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		fmt.Println("Hint: Make sure the specified directory exists and contains valid Swift files.")
		return
	}

	jsonData, err := json.MarshalIndent(testData, "", "  ")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = os.WriteFile("tests.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Println("Test data JSON file generated successfully.")

	orphanJsonData, err := json.MarshalIndent(orphanTestData, "", "  ")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = os.WriteFile("orphan.json", orphanJsonData, 0644)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	fmt.Println("Orphan test data JSON file generated successfully.")
}

func printHelpMenu() {
	fmt.Println("Usage: go run main.go -path <directory_path>")
	fmt.Println("Options:")
	fmt.Println("  -path string")
	fmt.Println("        Directory path to parse Swift files")
}

func parseSwiftFiles(path string) ([]TestData, []TestData, error) {
	var testDataList []TestData
	var orphanTestDataList []TestData

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.Contains(path, "UITests") && strings.HasSuffix(path, ".swift") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			testData := TestData{
				Filename: info.Name(),
				Filepath: path,
			}

			orphanTestData := TestData{
				Filename: info.Name(),
				Filepath: path,
			}

			scanner := bufio.NewScanner(file)
			var commentBlock []string
			var testCase TestCase
			var line int
			var orphan bool
			orphan = true

			for scanner.Scan() {
				text := strings.TrimSpace(scanner.Text())
				line++
				if strings.HasPrefix(text, "/*") {
					commentBlock = []string{text}
				} else if len(commentBlock) > 0 {
					commentBlock = append(commentBlock, text)
					if strings.HasSuffix(text, "*/") {
						testCase = parseCommentBlock(commentBlock)
						orphan = false
						commentBlock = nil
					}
				} else if strings.HasPrefix(text, "}") {
					orphan = true
					commentBlock = nil
				} else if strings.HasPrefix(text, "func test") {
					method := strings.TrimSuffix(text, "{")
					if orphan {
						orphanTestData.Tests = append(orphanTestData.Tests, Test{
							Method: strings.TrimSpace(method),
							Line:   line,
						})
					} else {
						testData.Tests = append(testData.Tests, Test{
							Method: strings.TrimSpace(method),
							Line:   line,
							Block:  testCase,
						})
					}
					orphan = true
					testCase = TestCase{}
				}
			}

			if len(testData.Tests) > 0 {
				testDataList = append(testDataList, testData)
			}

			if len(orphanTestData.Tests) > 0 {
				orphanTestDataList = append(orphanTestDataList, orphanTestData)
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return testDataList, orphanTestDataList, nil
}

func parseCommentBlock(commentBlock []string) TestCase {
	var testCase TestCase
	var currentSection string

	for _, line := range commentBlock {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*/") {
			continue
		}

		if strings.HasPrefix(line, "GIVEN") {
			currentSection = "given"
			continue
		} else if strings.HasPrefix(line, "WHEN") {
			currentSection = "when"
			continue
		} else if strings.HasPrefix(line, "THEN") {
			currentSection = "then"
			continue
		}

		if strings.HasPrefix(line, "#") {
			localTags := strings.Split(line, " ")
			for _, tag := range localTags {
				tagParts := strings.Split(strings.TrimPrefix(tag, "#"), ":")
				if len(tagParts) == 1 {
					testCase.Tags = append(testCase.Tags, Tag{Key: "info", Value: tagParts[0]})
				} else if len(tagParts) == 2 {
					testCase.Tags = append(testCase.Tags, Tag{Key: tagParts[0], Value: tagParts[1]})
				}
			}
			continue
		}

		if strings.HasPrefix(line, "-") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
			switch currentSection {
			case "given":
				testCase.Given = append(testCase.Given, line)
			case "when":
				testCase.When = append(testCase.When, line)
			case "then":
				testCase.Then = append(testCase.Then, line)
			}
		}
	}

	return testCase
}
