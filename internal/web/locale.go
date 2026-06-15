package web

import "github.com/school-management/pos/internal/i18n"

func SetRenderLang(lang i18n.Locale) {
	i18n.SetCurrent(lang)
}
