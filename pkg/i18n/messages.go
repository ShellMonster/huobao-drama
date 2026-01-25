package i18n

import "strings"

var messages = map[string]map[string]string{
	LangZhCN: {},
	LangEnUS: {},
}

func Translate(lang string, key string, args map[string]string) string {
	if key == "" {
		return ""
	}
	target := key
	if catalog, ok := messages[Normalize(lang)]; ok {
		if value, ok := catalog[key]; ok {
			target = value
		}
	}
	if len(args) == 0 {
		return target
	}
	result := target
	for name, value := range args {
		if name == "" {
			continue
		}
		result = strings.ReplaceAll(result, "{"+name+"}", value)
	}
	return result
}

func AddMessages(lang string, entries map[string]string) {
	if len(entries) == 0 {
		return
	}
	normalized := Normalize(lang)
	if _, ok := messages[normalized]; !ok {
		messages[normalized] = map[string]string{}
	}
	for key, value := range entries {
		messages[normalized][key] = value
	}
}
