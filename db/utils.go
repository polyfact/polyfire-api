package db

import (
	"fmt"
	"regexp"
)

var uuidRegexp = "^[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$"
var slugRegexp = "^[a-z0-9-_]*$"

func PrepareQueryByStringIdentifier(id string, fieldMap map[string]string) (string, error) {
	matchUUID, _ := regexp.MatchString(uuidRegexp, id)
	matchSlug, _ := regexp.MatchString(slugRegexp, id)

	if !matchUUID && !matchSlug {
		return "", fmt.Errorf("Invalid identifier")
	}

	var query string

	if matchUUID {
		query = fmt.Sprintf("%s = ?", fieldMap["uuid"])
	} else {
		query = fmt.Sprintf("%s = ?", fieldMap["slug"])
	}

	return query, nil
}
