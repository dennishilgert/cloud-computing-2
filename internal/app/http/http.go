package http

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/dennishilgert/cloud-computing-2/internal/app/cache"
	"github.com/dennishilgert/cloud-computing-2/internal/app/translate"
	"github.com/dennishilgert/cloud-computing-2/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var log = logger.NewLogger("app.http")

type Options struct {
	Port int
}

type Server interface {
	Run(ctx context.Context) error
	Ready(ctx context.Context) error
}

type httpServer struct {
	port       int
	readyCh    chan struct{}
	running    atomic.Bool
	translator translate.Translator
	cache      cache.Cache
}

func NewHttpServer(translator translate.Translator, cache cache.Cache, opts Options) Server {
	return &httpServer{
		port:       opts.Port,
		readyCh:    make(chan struct{}),
		translator: translator,
		cache:      cache,
	}
}

type TemplateRenderer struct {
	templates *template.Template
}

// Render renders a template with the given parameters.
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// Run starts the http server.
func (a *httpServer) Run(ctx context.Context) error {
	if !a.running.CompareAndSwap(false, true) {
		return errors.New("http server is already running")
	}

	log.Infof("starting http server on port %d", a.port)

	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Recover())

	// Initialize the template renderer
	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("resources/frontend/*.html")),
	}
	e.Renderer = renderer

	// return the rendered template to the client
	e.GET("/", func(c echo.Context) error {
		// Render the index template with any dynamic data (if necessary)
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{})
	})

	//
	e.POST("/languages", func(c echo.Context) error {
		values, err := c.FormParams()
		if err != nil {
			return c.String(http.StatusBadRequest, "Invalid form data")
		}

		excludeSelection := values.Get("targetLang")
		currentSelection := values.Get("sourceLang")
		// swap current and excluded language if the element that
		// triggered the request is the target language selection
		if values.Get("element") == "targetLang" {
			excludeSelection, currentSelection = currentSelection, excludeSelection
		}

		var htmlOut strings.Builder
		// add the current selected language as first and therefore
		// automatically selected option if it should no be excluded
		if currentSelection != excludeSelection {
			htmlOut.WriteString(fmt.Sprintf("<option>%v</option>", currentSelection))
		}

		// append each available language except the current and excluded language
		for _, lang := range a.translator.AvailableLanguages().DisplayNames() {
			if strings.EqualFold(excludeSelection, lang) || strings.EqualFold(currentSelection, lang) {
				continue
			}
			htmlOut.WriteString(fmt.Sprintf("<option>%v</option>\n", lang))
		}

		return c.HTML(http.StatusOK, htmlOut.String())
	})

	e.POST("/translate", func(c echo.Context) error {
		values, _ := c.FormParams()
		sourceLang := a.translator.AvailableLanguages().ByDisplayName(values.Get("sourceLang"))
		targetLang := a.translator.AvailableLanguages().ByDisplayName(values.Get("targetLang"))
		inputText := strings.TrimSpace(values.Get("sourceText"))

		if inputText == "" {
			return c.String(http.StatusOK, "")
		}

		hashedKey, has := a.cache.Has(ctx, inputText, targetLang.IsoCode)
		log.Infof("checking if translation is cached: %s", hashedKey)
		if has {
			log.Infof("retrieving translation from cache: %s", hashedKey)
			return c.String(http.StatusOK, a.cache.Get(ctx, inputText, targetLang.IsoCode))
		}

		log.Infof("retrieving translation from cloud translation api: %s", hashedKey)
		translated, err := a.translator.Translate(ctx, sourceLang.IsoCode, targetLang.IsoCode, inputText)
		if err != nil {
			log.Errorf("failed to translate text: %v", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
		log.Infof("storing translation in cache: %s", hashedKey)
		if err := a.cache.Add(ctx, inputText, targetLang.IsoCode, *translated); err != nil {
			log.Errorf("failed to cache translation: %s, reason: %v", hashedKey, err)
		}

		return c.String(http.StatusOK, *translated)
	})

	// close ready channel to mark server as listening
	close(a.readyCh)

	errCh := make(chan error, 1)
	go func() {
		defer close(errCh) // ensure channel is closed to avoid goroutine leak

		if err := e.Start(fmt.Sprintf(":%d", a.port)); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				errCh <- fmt.Errorf("error while starting http server: %w", err)
			}
			return
		}
		errCh <- nil
	}()

	// block until the context is done
	<-ctx.Done()

	a.translator.Close()

	err := e.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = <-errCh
	if err != nil {
		return err
	}

	return nil
}

// Ready waits until the http server is ready or the context is cancelled due to timeout.
func (a *httpServer) Ready(ctx context.Context) error {
	select {
	case <-a.readyCh:
		return nil
	case <-ctx.Done():
		return errors.New("timeout while waiting for the http server to be ready")
	}
}
