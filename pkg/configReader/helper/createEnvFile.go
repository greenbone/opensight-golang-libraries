// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	// Find all Go files in the current directory and its subdirectories
	var envVarsGoFile []string
	var envVars []string
	err := filepath.Walk("internal/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != "." && strings.HasPrefix(info.Name(), ".") {
			// Skip hidden directories (e.g. ".git")
			return filepath.SkipDir
		}

		if filepath.Ext(path) != ".go" || strings.Contains(path, "_test.go") {
			return nil
		}

		// Parse the Go file and look for structs with the "viperEnv" tag
		fmt.Printf("Looking for %s\n", path)
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		var structDef bool
		var structFields []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			// Check for the start of a struct definition
			if regexp.MustCompile(`type\s+\w+\s+struct\s*{`).MatchString(line) {
				structDef = true
				continue
			}

			// Check for the end of a struct definition
			if structDef && strings.TrimSpace(line) == "}" {
				if len(structFields) > 0 {
					// Found a struct definition with fields, look for fields with "viperEnv" tags
					for _, field := range structFields {
						if strings.Contains(field, "`") {
							tagStart := strings.Index(field, "`")
							tagEnd := strings.LastIndex(field, "`")
							tag := field[tagStart : tagEnd+1]
							if strings.Contains(tag, "viperEnv:") {
								envStart := strings.Index(tag, "viperEnv:\"") + len("viperEnv:\"")
								envEnd := strings.Index(tag[envStart:], "\"") + envStart
								envName := tag[envStart:envEnd]
								defaultStart := strings.Index(tag, "default:\"")
								if defaultStart != -1 {
									defaultStart += len("default:\"")
									defaultEnd := strings.Index(tag[defaultStart:], "\"") + defaultStart
									defaultValue := tag[defaultStart:defaultEnd]
									envVars = append(envVars, fmt.Sprintf("%s=%s", envName, defaultValue))
									envVarsGoFile = append(envVarsGoFile, fmt.Sprintf(
										"os.Setenv(\"%s\", \"%s\")", envName,
										strings.ReplaceAll(defaultValue, "\"", "\\\"")))

								} else {
									envVarsGoFile = append(envVarsGoFile, fmt.Sprintf(
										"os.Setenv(\"%s\", \"EMPTY\")", envName))
									envVars = append(envVars, fmt.Sprintf("%s=EMPTY", envName))
								}
							}
						}
					}
				}

				structDef = false
				structFields = nil
				continue
			}

			// Check for a field definition inside a struct
			if structDef && regexp.MustCompile(`^\s*\w+\s+\w+\s*`).MatchString(line) {
				structFields = append(structFields, strings.TrimSpace(line))
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	// Write the output file
	/*
		f, err := os.Create("setenvAuto.go")
		if err != nil {
			fmt.Printf("Error creating file: %s\n", err)
			return
		}
		defer f.Close()

		fmt.Fprintln(f, "package main\n")
		fmt.Fprintln(f, "import \"os\"\n")
		fmt.Fprintln(f, "func init() {")
		for _, envVar := range envVarsGoFile {
			fmt.Fprintln(f, envVar)
		}
		fmt.Fprintln(f, "}")
	*/
	f2, err := os.Create("asset-backend.env")
	if err != nil {
		fmt.Printf("Error creating file: %s\n", err)
		return
	}
	defer f2.Close()

	for _, envVar := range envVars {
		fmt.Fprintln(f2, envVar)
	}
}
