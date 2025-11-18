package i18n

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var i *I18N

type I18N struct {
	localizer *i18n.Localizer
	bundle    *i18n.Bundle
}

func InitI18N(messageFilePaths []string) error {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, messageFile := range messageFilePaths {
		_, err := bundle.LoadMessageFile(messageFile)
		if err != nil {
			return err
		}
	}

	localizer := i18n.NewLocalizer(bundle, language.English.String(), language.Spanish.String())

	i = &I18N{localizer: localizer, bundle: bundle}

	return nil
}

// SetLanguage changes the language used by the localizer
func SetLanguage(lang language.Tag) {
	localizer := i18n.NewLocalizer(i.bundle, lang.String())

	i = &I18N{localizer: localizer, bundle: i.bundle}
}

func SetWithCode(code string) error {
	lang, err := language.Parse(code)
	if err != nil {
		return err
	}
	SetLanguage(lang)
	return nil
}

// GetString retrieves a localized string by key
// If the key is not found, it returns the key itself as fallback
func GetString(key string) string {
	msg, err := i.localizer.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	if err != nil {
		return "I18N Error"
	}
	return msg
}

// GetStringWithData retrieves a localized string by key with template data
// templateData should be a map[string]interface{} containing values for template variables
func GetStringWithData(key string, templateData map[string]interface{}) string {
	msg, err := i.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: templateData,
	})
	if err != nil {
		return "I18N Error"
	}
	return msg
}

// GetPluralString retrieves a localized string with plural support
// count determines which plural form to use
func GetPluralString(key string, count int) string {
	msg, err := i.localizer.Localize(&i18n.LocalizeConfig{
		MessageID:   key,
		PluralCount: count,
	})
	if err != nil {
		return "I18N Error"
	}
	return msg
}
