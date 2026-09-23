// Package localization connects Feed's translations and regional formatters.
package localization

import (
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/goodsign/monday"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed locales/*.json
var catalogs embed.FS

type Locale struct {
	Tag        language.Tag
	Name       string
	DateLocale monday.Locale
}

// The first locale is the fallback. Treat this list as read-only after startup.
var supportedLocales = []Locale{
	{Tag: language.AmericanEnglish, Name: "English", DateLocale: monday.LocaleEnUS},
	{Tag: language.MustParse("ru-RU"), Name: "Русский", DateLocale: monday.LocaleRuRU},
}

type Catalog struct {
	bundle  *i18n.Bundle
	matcher language.Matcher
}

func New() (*Catalog, error) {
	bundle := i18n.NewBundle(supportedLocales[0].Tag)
	tags := make([]language.Tag, 0, len(supportedLocales))
	for _, locale := range supportedLocales {
		path := "locales/" + locale.Tag.String() + ".json"
		if _, err := bundle.LoadMessageFileFS(catalogs, path); err != nil {
			return nil, fmt.Errorf("load translations %s: %w", path, err)
		}
		tags = append(tags, locale.Tag)
	}
	return &Catalog{bundle: bundle, matcher: language.NewMatcher(tags)}, nil
}

// Lookup accepts a supported locale, including equivalent casing such as en-us.
// Unlike language negotiation, an explicit choice must name a supported locale.
func Lookup(value string) (Locale, bool) {
	tag, err := language.Parse(value)
	if err != nil {
		return Locale{}, false
	}
	for _, locale := range supportedLocales {
		if tag == locale.Tag {
			return locale, true
		}
	}
	return Locale{}, false
}

type Localizer struct {
	locale   Locale
	messages *i18n.Localizer
	numbers  *message.Printer
}

func (c *Catalog) Localizer(cookieValue, acceptLanguage string) *Localizer {
	locale, ok := Lookup(cookieValue)
	if !ok {
		locale = supportedLocales[0]
		tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
		if err == nil && len(tags) > 0 {
			_, index, confidence := c.matcher.Match(tags...)
			if confidence >= language.High {
				// Use our exact supported tag for both translations and formats.
				locale = supportedLocales[index]
			}
		}
	}
	return &Localizer{
		locale:   locale,
		messages: i18n.NewLocalizer(c.bundle, locale.Tag.String()),
		numbers:  message.NewPrinter(locale.Tag),
	}
}

func (l *Localizer) Locale() string {
	return l.locale.Tag.String()
}

// Locales supplies the language selector. Callers must not modify the slice.
func (l *Localizer) Locales() []Locale {
	return supportedLocales
}

// T translates a message with an optional map of named template parameters.
func (l *Localizer) T(id string, data ...map[string]any) string {
	config := &i18n.LocalizeConfig{MessageID: id}
	if len(data) > 0 {
		config.TemplateData = data[0]
	}
	text, err := l.messages.Localize(config)
	if err != nil {
		log.Printf("localize %q (%s): %v", id, l.Locale(), err)
	}
	// go-i18n can return the fallback translation together with an error.
	return text
}

func (l *Localizer) DateTime(value time.Time, zone *time.Location) string {
	value = value.In(zone)
	locale := l.locale.DateLocale
	return l.T("datetime.full", map[string]any{
		"Date": monday.Format(value, monday.FullFormatsByLocale[locale], locale),
		"Time": monday.Format(value, monday.TimeFormatsByLocale[locale], locale),
		"Zone": zone.String(),
	})
}

func (l *Localizer) Integer(value int64) string {
	return l.numbers.Sprintf("%d", value)
}

// Decimal formats a number with the requested number of fractional digits.
// Callers provide a nonnegative precision; it is not a user input setting.
func (l *Localizer) Decimal(value float64, precision int) string {
	return l.numbers.Sprintf("%.*f", precision, value)
}
