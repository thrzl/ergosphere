package utils

import "strings"

import (
	"bufio"
)

func ParseHostFile(hosts string) ([]string, error) {
	var records []string

	scanner := bufio.NewScanner(strings.NewReader(hosts))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		records = append(records, fields[1]+".")
	}

	if err := scanner.Err(); err != nil {
		return records, err
	}

	return records, nil
}