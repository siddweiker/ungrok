package main

//go:generate go run gen.go

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
)

var (
	findGrok = regexp.MustCompile(`(%{(\w+)(?::([\w\[\]-]+))?(?::[^}]+)?})`)
	//go:embed grok-patterns
	embedGroks []byte
)

func main() {
	pattern := flag.String("pattern", "", "The GROK pattern to translate into regex")
	config := flag.String("config", "", "A directory containing grok pattern files")
	output := flag.String("output", "", "The output file to write too, default stdout")
	flag.Parse()

	if *pattern == "" {
		fmt.Println("A 'pattern' argument is required")
		flag.Usage()
		os.Exit(1)
	}

	var f io.Reader = bytes.NewBuffer(embedGroks)
	if *config != "" {
		files, err := readDirContents(*config)
		if err != nil {
			log.Fatal(err)
		}
		f = files
	}

	groks, err := parsePatterns(f)
	if err != nil {
		log.Fatal(err)
	}

	result, err := ungrok(*pattern, groks)
	if err != nil {
		log.Fatal(err)
	}

	// Must use regexp2 as golang regexp is RE2 based which does not support
	// featues like backtracking, and group names within '<>'
	_, err = regexp2.Compile(result, 0)
	if err != nil {
		log.Printf("Generated regex does not compile:\n%s", result)
		log.Fatalf("Error: %s", err)
	}

	var out io.Writer = os.Stdout
	if *output != "" {
		o, err := os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
		defer o.Close()
		out = o
	}

	fmt.Fprintf(out, "%s\n", result)
}

func readDirContents(dir string) (io.Reader, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var b bytes.Buffer
	var errs error
	for _, entry := range files {
		if entry.IsDir() {
			continue
		}

		filename := filepath.Join(dir, entry.Name())
		d, err := os.ReadFile(filename)
		if err != nil {
			errs = errors.Join(errs, fmt.Errorf("error reading file '%s': %w", filename, err))
			continue
		}

		_, err = b.Write(d)
		if err != nil {
			errs = errors.Join(fmt.Errorf("error copying file '%s': %w", filename, err))
		}
	}

	if errs != nil {
		return nil, errs
	}

	return &b, nil
}

func parsePatterns(in io.Reader) (map[string]string, error) {
	groks := make(map[string]string)

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if scanner.Text() == "" || strings.HasPrefix(scanner.Text(), "#") {
			continue
		}
		p, v, ok := strings.Cut(scanner.Text(), " ")
		if ok {
			groks[p] = v
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return groks, nil
}

func ungrok(pattern string, groks map[string]string) (string, error) {
	original := pattern
	match := true

	for match {
		match = false
		for _, group := range findGrok.FindAllStringSubmatch(pattern, -1) {
			match = true
			replace, grk, name := group[1], group[2], group[3]
			val, ok := groks[grk]

			if name != "" {
				name = strings.ReplaceAll(name, "-", "_")
				name = strings.ReplaceAll(name, "][", "_")
				name = strings.Trim(name, "[]")
				val = fmt.Sprintf("(?<%s>%s)", name, val)
			}

			if ok {
				pattern = strings.ReplaceAll(pattern, replace, val)
			} else {
				return "", fmt.Errorf("could not find grok pattern: %s", grk)
			}
		}
	}

	if original == pattern {
		return "", fmt.Errorf("no pattern found for: %s", original)
	}

	return pattern, nil
}
