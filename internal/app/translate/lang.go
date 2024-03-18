package translate

import (
	"slices"
	"strings"

	"cloud.google.com/go/translate/apiv3/translatepb"
)

type Language struct {
	DisplayName string
	IsoCode     string
}

type AvailableLanguages interface {
	ByDisplayName(displayName string) Language
	DisplayNames() []string
}

type availableLanguages struct {
	languages map[string]Language
}

// ParseAvailableLanguages parses the supported languages from Google Cloud.
func ParseAvailableLanguages(supportedLanguages []*translatepb.SupportedLanguage) AvailableLanguages {
	languages := map[string]Language{}
	for _, language := range supportedLanguages {
		languages[strings.ToLower(language.DisplayName)] = Language{
			DisplayName: language.GetDisplayName(),
			IsoCode:     language.GetLanguageCode(),
		}
	}
	return &availableLanguages{
		languages: languages,
	}
}

// ByDisplayName returns an available language by its display name.
func (a *availableLanguages) ByDisplayName(displayName string) Language {
	return a.languages[strings.ToLower(displayName)]
}

// DisplayNames returns a list with the display names of all available languages.
func (a *availableLanguages) DisplayNames() []string {
	names := make([]string, 0, len(a.languages))
	for _, lang := range a.languages {
		names = append(names, lang.DisplayName)
	}
	slices.Sort(names)
	return names
}
