# UNGROK
Convert GROK patterns to Regex

## Installation
```cmd
go install github.com/siddweiker/ungrok@latest
```

## Usage

```cmd
Usage of ungrok:
  -config string
        A directory containing grok pattern files
  -output string
        The output file to write too, default stdout
  -pattern string
        The GROK pattern to translate into regex
```

## Example
```cmd
ungrok -pattern "%{TIMESTAMP_ISO8601}"
ungrok -pattern "%{MONTH}" -output out.txt
ungrok -pattern "%{POSTGRESQL}" -config /path/to/patternsdir
```

## Building from source
```cmd
go generate
# Alternatively go run gen.go -url "custom_pattern_file"
go build
```
