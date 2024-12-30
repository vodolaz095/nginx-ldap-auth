package endpoints

import (
	"bytes"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/vodolaz095/nginx-ldap-auth/views"
)

func unescape(s string) template.HTML {
	return template.HTML(s)
}

func injectTemplates(r *gin.Engine) (err error) {
	templatesAsBytes := bytes.NewBufferString("")
	entries, err := views.Views.ReadDir(".")
	if err != nil {
		return err
	}
	var data []byte
	for i := range entries {
		log.Debug().Msgf("Reading %s...", entries[i].Name())
		data, err = views.Views.ReadFile(entries[i].Name())
		if err != nil {
			return err
		}
		_, err = templatesAsBytes.Write(data)
		if err != nil {
			return err
		}
	}
	f := template.FuncMap{
		"unescape": unescape,
		"inc": func(n int) int {
			return n + 1
		},
	}
	tpl, err := template.New("nginx-ldap-auth").Funcs(f).Parse(templatesAsBytes.String())
	if err != nil {
		return err
	}
	r.SetHTMLTemplate(tpl)
	return nil
}
