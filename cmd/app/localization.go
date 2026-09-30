package main

import (
	"context"
	"net/http"
	"time"

	"github.com/StaticD1/feed/internal/localization"
)

type localizerContextKey struct{}

func (app *application) withLocale(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var cookieValue string
		if cookie, err := r.Cookie("locale"); err == nil {
			cookieValue = cookie.Value
		}
		localizer := app.catalog.Localizer(cookieValue, r.Header.Get("Accept-Language"))
		w.Header().Set("Content-Language", localizer.Locale())
		w.Header().Add("Vary", "Accept-Language")
		w.Header().Add("Vary", "Cookie")

		ctx := context.WithValue(r.Context(), localizerContextKey{}, localizer)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// All application handlers are registered behind withLocale.
func requestLocalizer(r *http.Request) *localization.Localizer {
	return r.Context().Value(localizerContextKey{}).(*localization.Localizer)
}

type pageData struct {
	I18n     *localization.Localizer
	ReturnTo string
	TimeZone *time.Location
}

func newPageData(r *http.Request) pageData {
	return pageData{
		I18n:     requestLocalizer(r),
		ReturnTo: r.URL.Path,
		TimeZone: time.UTC,
	}
}

func (app *application) changeLocale(w http.ResponseWriter, r *http.Request) {
	localizer := requestLocalizer(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, localizer.T("errors.invalid_form"), http.StatusBadRequest)
		return
	}
	locale, ok := localization.Lookup(r.PostForm.Get("locale"))
	if !ok {
		http.Error(w, localizer.T("errors.unsupported_locale"), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "locale",
		Value:    locale.Tag.String(),
		Path:     "/",
		MaxAge:   365 * 24 * 60 * 60,
		Expires:  time.Now().Add(365 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	returnTo := r.PostForm.Get("return_to")
	switch returnTo {
	case "/", "/login", "/register":
	default:
		returnTo = "/"
	}
	http.Redirect(w, r, returnTo, http.StatusSeeOther)
}
