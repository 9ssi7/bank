package middlewares

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var localeKey = "lang"

func ParseLocale(ctx *fiber.Ctx) string {
	return ctx.Locals(localeKey).(string)
}

func NewI18n(locales []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		acceptedLanguages := locales
		l := c.Query("lang")
		list := strings.Split(l, ";")
		alternative := ""
		locales := findLocales(list, locales)
		for _, v := range acceptedLanguages {
			if locales[v] {
				l = v
				break
			}
		}
		if len(list) > 1 {
			alternative = list[1]
		}
		if alternative != "" && l != "" && locales[alternative] {
			l = alternative
		}
		c.Locals(localeKey, l)
		c.SetUserContext(context.WithValue(c.UserContext(), localeKey, l))
		return c.Next()
	}
}

func findLocales(list []string, defaultLocales []string) map[string]bool {
	locales := make(map[string]bool)
	acceptedLanguages := defaultLocales
	for _, li := range list {
		lineItems := strings.Split(li, ",")
		for _, word := range lineItems {
			for _, v := range acceptedLanguages {
				if strings.Contains(word, v) {
					locales[v] = true
				}
			}
			if len(word) == 2 && word[1] == '-' {
				locales[strings.ToLower(word)] = true
			}
			if len(word) == 5 && word[2] == '-' {
				double := strings.Split(word, "-")
				locales[double[0]] = true
			}
		}
	}
	return locales
}
