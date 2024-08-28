package main

// https://github.com/pquerna/otp
import (
	"html/template"
	"io"
	"log"
	traveller "visualsource/traveller/internal"
	"visualsource/traveller/internal/handler"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplateRegistry struct {
	templates *template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {

	if viewContent, isMap := data.(map[string]interface{}); isMap {
		viewContent["reverse"] = c.Echo().Reverse
	}

	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {

	h, err := handler.New()
	if err != nil {
		log.Fatal(err)
		return
	}

	defer h.Close()

	e := echo.New()
	e.Debug = true

	e.Renderer = &TemplateRegistry{
		templates: template.Must(template.ParseGlob("web/template/*.html")),
	}

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("secret-password"))))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	e.Static("/static", "./web/static")

	e.File("/", "web/app/index.html")

	traveller.RegisterAuthPages(e, h)
	traveller.RegisterSessionPages(e, h)

	e.Logger.Fatal(e.Start(":8080"))
}
