package snowboard

import (
	"html/template"
	"io"
	"path"
	"strings"

	"github.com/subosito/snowboard/blueprint"
)

func markdownize(s string) template.HTML {
	return template.HTML(string(markdown([]byte(s))))
}

func parameterize(s string) string {
	return strings.Replace(strings.ToLower(s), " ", "-", -1)
}

func multiParameterize(g blueprint.ResourceGroup, r blueprint.Resource, t blueprint.Transition) (s string) {
	xs := []string{}

	if g.Title != "" {
		xs = append(xs, parameterize(g.Title))
	}

	if r.Title != "" {
		xs = append(xs, parameterize(r.Title))
	} else {
		xs = append(xs, parameterize(r.Href.Path))
	}

	if t.Title != "" {
		xs = append(xs, parameterize(t.Title))
	} else {
		xs = append(xs, parameterize(requestMethod(t)))
	}

	return strings.Join(xs, "-")
}

func requestMethod(t blueprint.Transition) string {
	for _, m := range t.Transactions {
		return m.Request.Method
	}

	return ""
}

func transitionColorize(t blueprint.Transition) string {
	return colorize(requestMethod(t))
}

func apiUrl(b *API, s string, sr string) string {
	var h string

	for _, m := range b.Metadata {
		if m.Key == "HOST" {
			h = m.Value
		}
	}

	if s != "" {
		return path.Join(h, s)
	}

	return path.Join(h, sr)
}

func iColorize(i int) string {
	switch i {
	case 200, 201, 202, 204:
		return "blue"
	case 401, 403, 404, 422:
		return "orange"
	case 500:
		return "red"
	}

	return ""
}

func colorize(s string) string {
	switch s {
	case "GET":
		return "green"
	case "POST":
		return "blue"
	case "PUT":
		return "teal"
	case "PATCH":
		return "violet"
	case "DELETE":
		return "red"
	}

	return ""
}

func alias(s string) string {
	switch s {
	case "application/json":
		return "json"
	}

	return ""
}

func buildDataStructures(t blueprint.Transaction, s blueprint.Transition, r blueprint.Resource, a API) (ds []blueprint.DataStructure) {
	for _, ts := range t.Response.DataStructures {
		for _, rs := range r.DataStructures {
			if ts.Name == rs.ID && rs.Name != "array" {
				ds = append(ds, rs)
			}

			for _, as := range a.DataStructures {
				if rs.Name == as.ID {
					ds = append(ds, as)
				}

				for _, ss := range s.DataStructures {
					if ss.Name == as.ID {
						ds = append(ds, as)
					}
				}
			}
		}
	}

	return
}

// HTML renders blueprint.API struct as HTML document
func HTML(tpl string, w io.Writer, b *API) error {
	funcMap := template.FuncMap{
		"markdownize":         markdownize,
		"parameterize":        parameterize,
		"mParameterize":       multiParameterize,
		"colorize":            colorize,
		"iColorize":           iColorize,
		"transitionColorize":  transitionColorize,
		"apiUrl":              apiUrl,
		"buildDataStructures": buildDataStructures,
		"requestMethod":       requestMethod,
		"alias":               alias,
	}

	tmpl, err := template.New("api").Funcs(funcMap).Parse(tpl)
	if err != nil {
		return err
	}

	err = tmpl.Execute(w, b)
	if err != nil {
		return err
	}

	return nil
}
