package version

import "fmt"

const MAJOR uint = 1
const MINOR uint = 0
const PATCH uint = 0

var COMMIT string
var IDENTIFIER string
var METADATA string

func Version() string {
	var suffix string
	if len(IDENTIFIER) > 0 {
		suffix = fmt.Sprintf("-%s", IDENTIFIER)
	}

	if len(COMMIT) > 0 || len(METADATA) > 0 {
		suffix = suffix + "+"
	}

	if len(COMMIT) > 0 {
		suffix = fmt.Sprintf("%s"+"commit.%s", suffix, COMMIT)

	}

	if len(METADATA) > 0 {
		if len(COMMIT) > 0 {
			suffix = suffix + "."
		}
		suffix = suffix + METADATA
	}

	return fmt.Sprintf("%d.%d.%d%s", MAJOR, MINOR, PATCH, suffix)
}
