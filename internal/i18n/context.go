package i18n

var current = EN

func SetCurrent(l Locale) {
	current = l
}

func Current() Locale {
	return current
}

func TC(key string) string {
	return T(current, key)
}

func CurrentPageTitle(english string) string {
	return PageTitle(current, english)
}
