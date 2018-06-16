package installservd

import (
	"fmt"
	"net/http"
	"path/filepath"

	"bytes"

	"github.com/labstack/echo"
)

func listProfiles(context echo.Context) (err error) {
	res := context.Response()
	res.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	if _, err = fmt.Fprintf(res, "<pre>\n"); err != nil {
		return
	}
	for _, d := range Profiles {
		name := d.Name
		color := "#e91e63"
		if _, err = fmt.Fprintf(res, "<a href=\"/profiles/%s\" style=\"color: %s;\">%s</a>\n", name, color, name); err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(res, "</pre>\n")
	return
}

func getProfile(c echo.Context) (err error) {
	var profile *Profile
	enteredName := c.Param("name")
	for _, p := range Profiles {
		if p.Name == enteredName {
			profile = &p
		}
	}
	if profile == nil {
		return c.NoContent(http.StatusNotFound)
	}
	res := c.Response()
	res.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	color := "#212121"
	if _, err = fmt.Fprintf(res, "<pre>\n"); err != nil {
		return
	}
	if _, err = fmt.Fprintf(res, "<a href=\"/profiles/%s/config.json\" style=\"color: %s;\">%s</a>\n", profile.Name, color, "config.json"); err != nil {
		return
	}
	for _, d := range profile.Templates {
		_, name := filepath.Split(d.Path)
		if _, err = fmt.Fprintf(res, "<a href=\"/profiles/%s/%s\" style=\"color: %s;\">%s</a>\n", profile.Name, name, color, name); err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(res, "</pre>\n")
	return
}

func getProfileConfig(c echo.Context) (err error) {
	var profile *Profile
	enteredName := c.Param("name")
	for _, p := range Profiles {
		if p.Name == enteredName {
			profile = &p
		}
	}
	if profile == nil {
		return c.NoContent(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, profile.Config)
}

func getTemplate(i *Installservd) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var profile *Profile
		enteredName := c.Param("name")
		templateName := c.Param("template")
		for _, p := range Profiles {
			if p.Name == enteredName {
				profile = &p
			}
		}
		if profile == nil {
			return c.NoContent(http.StatusNotFound)
		}
		for _, t := range profile.Templates {
			_, name := filepath.Split(t.Path)
			if name == templateName {
				buf := new(bytes.Buffer)
				if err = i.Echo.Renderer.Render(buf, i.getAssetPath(*t), profile, c); err != nil {
					return
				}
				return c.String(http.StatusOK, buf.String())
			}
		}
		return c.NoContent(http.StatusNotFound)
	}
}
