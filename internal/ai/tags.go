package ai

import (
	"regexp"
	"strings"
)

// Tag types emitted by the DM AI inside its narrative response.
// The server strips these before sending text to the client.
const (
	TagRoll       = "ROLL"
	TagMapUpdate  = "MAP_UPDATE"
	TagLore       = "LORE"
	TagSuggestRoll = "SUGGEST_ROLL"
	TagXP         = "XP"
	TagLevelUp    = "LEVEL_UP"
)

// ParsedTag holds a single extracted AI tag with its raw payload.
type ParsedTag struct {
	Type    string
	Payload string
}

var tagRe = regexp.MustCompile(`\[([A-Z_]+):([^\]]*)\]`)

// ParseTags extracts all AI control tags from raw DM output and returns
// the clean narrative text (tags removed) plus the list of parsed tags.
func ParseTags(raw string) (narrative string, tags []ParsedTag) {
	matches := tagRe.FindAllStringSubmatchIndex(raw, -1)
	if len(matches) == 0 {
		return raw, nil
	}

	var cleanParts []string
	prev := 0
	for _, loc := range matches {
		// text before this tag
		cleanParts = append(cleanParts, raw[prev:loc[0]])
		prev = loc[1]

		tagType := raw[loc[2]:loc[3]]
		payload := strings.TrimSpace(raw[loc[4]:loc[5]])
		tags = append(tags, ParsedTag{Type: tagType, Payload: payload})
	}
	cleanParts = append(cleanParts, raw[prev:])

	narrative = strings.TrimSpace(strings.Join(cleanParts, ""))
	return narrative, tags
}
