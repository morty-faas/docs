package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/morty-faas/cli/cmd"
	"github.com/spf13/cobra/doc"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	basePath   = "../../pages/cli"
	fmTemplate = `---
title: "%s"
---
`
)

func main() {
	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		if base == "morty" {
			base = "CLI"
		}
		return fmt.Sprintf(fmTemplate, strings.Replace(base, "_", " ", -1))
	}

	err := doc.GenMarkdownTreeCustom(cmd.RootCmd, basePath, filePrepender, func(s string) string {
		base := strings.TrimSuffix(s, path.Ext(s))
		if base == "morty" {
			return "/cli"
		}
		return "/cli/" + strings.Replace(base, "_", "-", -1)
	})
	if err != nil {
		log.Fatal(err)
	}

	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".md") {
			err := removeAutoGeneratedLine(path)
			if err != nil {
				fmt.Printf("Error removing line from %s: %s\n", path, err)
			} else {
				fmt.Printf("Removed line from %s\n", path)
			}
		}

		return os.Rename(path, strings.Replace(path, "_", "-", -1))
	})

	if err != nil {
		fmt.Printf("Error walking directory: %s\n", err)
	}

	if err := os.Rename(path.Join(basePath, "morty.md"), path.Join(basePath, "..", "cli.md")); err != nil {
		log.Fatal(err)
	}
}

func removeAutoGeneratedLine(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "###### Auto generated by spf13/cobra on") {
			if strings.Contains(line, "SEE ALSO") {
				lines = append(lines, cases.Title(language.English, cases.Compact).String(line))
			} else {
				lines = append(lines, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(filePath, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}
