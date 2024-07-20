package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	dataRegex           = regexp.MustCompile(`^data:(?P<disk>[^\:]+):(?P<path>[^\:]+)`)
	dataDiskSubexpIndex = dataRegex.SubexpIndex("disk")
	dataPathSubexpIndex = dataRegex.SubexpIndex("path")

	dupRegex            = regexp.MustCompile(`^dup:(?P<disk1>[^\:]+):(?P<path1>[^\:]+):(?P<disk2>[^\:]+):(?P<path2>[^\:]+):(?P<bytes>[^\:]+): dup$`)
	dupDisk1SubexpIndex = dupRegex.SubexpIndex("disk1")
	dupPath1SubexpIndex = dupRegex.SubexpIndex("path1")
	dupDisk2SubexpIndex = dupRegex.SubexpIndex("disk2")
	dupPath2SubexpIndex = dupRegex.SubexpIndex("path2")
	dupBytesSubexpIndex = dupRegex.SubexpIndex("bytes")

	dataSet        = make(map[string]string)
	dupKeepSet     = make(map[string]bool)
	dupRemoveSet   = make(map[string]bool)
	dupRemoveCount = 0
	dupRemoveBytes = uint64(0)
)

func main() {
	var pathRegexStr string
	var minBytes uint64

	flag.StringVar(&pathRegexStr, "regex", ".*", "File path regex")
	flag.Uint64Var(&minBytes, "minbytes", 0, "File minimum bytes")

	flag.Parse()

	dupLogFileStr := flag.Arg(0)

	var dupLogInput io.Reader

	if dupLogFileStr != "" {
		dupFile, err := os.Open(dupLogFileStr)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		defer func() {
			err := dupFile.Close()

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}()

		dupLogInput = dupFile
	} else {
		dupLogInput = os.Stdin
	}

	dupLogScanner := bufio.NewScanner(dupLogInput)

	dupLogScanner.Split(bufio.ScanLines)

	pathRegex := regexp.MustCompile(pathRegexStr)

	for dupLogScanner.Scan() {
		line := dupLogScanner.Text()

		if dupLineMatch := dupRegex.FindStringSubmatch(line); dupLineMatch != nil {
			bytes, err := strconv.ParseUint(dupLineMatch[dupBytesSubexpIndex], 10, 64)

			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if bytes < minBytes {
				continue
			}

			disk1 := dupLineMatch[dupDisk1SubexpIndex]
			path1 := dupLineMatch[dupPath1SubexpIndex]
			disk2 := dupLineMatch[dupDisk2SubexpIndex]
			path2 := dupLineMatch[dupPath2SubexpIndex]

			fullPath1 := fmt.Sprintf("%s%s", dataSet[disk1], path1)
			fullPath2 := fmt.Sprintf("%s%s", dataSet[disk2], path2)

			if !pathRegex.MatchString(fullPath1) && !pathRegex.MatchString(fullPath2) {
				continue
			}

			dupKeepSet[fullPath1] = true
			dupRemoveSet[fullPath2] = true
			dupRemoveCount += 1
			dupRemoveBytes += bytes

			fmt.Println(fullPath2)
		} else if dataLineMatch := dataRegex.FindStringSubmatch(line); dataLineMatch != nil {
			dataDisk := dataLineMatch[dataDiskSubexpIndex]
			dataPath := dataLineMatch[dataPathSubexpIndex]

			dataSet[dataDisk] = dataPath
		}
	}

	for path := range dupKeepSet {
		if dupRemoveSet[path] {
			fmt.Fprintf(os.Stderr, "# ERROR: A file to be kept is found to be deleted: %s\n", path)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "# File path regex: %s\n", pathRegexStr)
	fmt.Fprintf(os.Stderr, "# File minimum bytes: %d\n", minBytes)
	fmt.Fprintf(os.Stderr, "# Total files: %d\n", dupRemoveCount)
	fmt.Fprintf(os.Stderr, "# Total bytes: %d\n", dupRemoveBytes)

	if dupLogFileStr != "" {
		escapedPathRegexStr := strings.ReplaceAll(pathRegexStr, `'`, `'\''`)
		fmt.Fprintf(os.Stderr, "# Suggested delete command: snapraid-dupls -regex '%s' -minbytes %d %s | xargs -I{} rm -v '{}'\n", escapedPathRegexStr, minBytes, dupLogFileStr)
	}
}
