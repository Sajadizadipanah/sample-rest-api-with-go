package ginSwagger

import (
	htmlTemplate "html/template"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sync"

	"golang.org/x/net/webdav"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/swag"
)

type swaggerConfig struct {
	URL                      string
	DocExpansion             string
	Title                    string
	Oauth2RedirectURL        htmlTemplate.JS
	DefaultModelsExpandDepth int
	DeepLinking              bool
	PersistAuthorization     bool
	Oauth2DefaultClientID    string
}

// Config stores ginSwagger configuration variables.
type Config struct {
	// The url pointing to API definition (normally swagger.json or swagger.yaml). Default is `doc.json`.
	URL                      string
	DocExpansion             string
	InstanceName             string
	Title                    string
	DefaultModelsExpandDepth int
	DeepLinking              bool
	PersistAuthorization     bool
	Oauth2DefaultClientID    string
}

func (config Config) toSwaggerConfig() swaggerConfig {
	return swaggerConfig{
		URL:                      config.URL,
		DeepLinking:              config.DeepLinking,
		DocExpansion:             config.DocExpansion,
		DefaultModelsExpandDepth: config.DefaultModelsExpandDepth,
		Oauth2RedirectURL: "`${window.location.protocol}//${window.location.host}$" +
			"{window.location.pathname.split('/').slice(0, window.location.pathname.split('/').length - 1).join('/')}" +
			"/oauth2-redirect.html`",
		Title:                 config.Title,
		PersistAuthorization:  config.PersistAuthorization,
		Oauth2DefaultClientID: config.Oauth2DefaultClientID,
	}
}

// URL presents the url pointing to API definition (normally swagger.json or swagger.yaml).
func URL(url string) func(*Config) {
	return func(c *Config) {
		c.URL = url
	}
}

// DocExpansion list, full, none.
func DocExpansion(docExpansion string) func(*Config) {
	return func(c *Config) {
		c.DocExpansion = docExpansion
	}
}

// DeepLinking set the swagger deep linking configuration.
func DeepLinking(deepLinking bool) func(*Config) {
	return func(c *Config) {
		c.DeepLinking = deepLinking
	}
}

// DefaultModelsExpandDepth set the default expansion depth for models
// (set to -1 completely hide the models).
func DefaultModelsExpandDepth(depth int) func(*Config) {
	return func(c *Config) {
		c.DefaultModelsExpandDepth = depth
	}
}

// InstanceName set the instance name that was used to generate the swagger documents
// Defaults to swag.Name ("swagger").
func InstanceName(name string) func(*Config) {
	return func(c *Config) {
		c.InstanceName = name
	}
}

// PersistAuthorization Persist authorization information over browser close/refresh.
// Defaults to false.
func PersistAuthorization(persistAuthorization bool) func(*Config) {
	return func(c *Config) {
		c.PersistAuthorization = persistAuthorization
	}
}

// Oauth2DefaultClientID set the default client ID used for OAuth2
func Oauth2DefaultClientID(oauth2DefaultClientID string) func(*Config) {
	return func(c *Config) {
		c.Oauth2DefaultClientID = oauth2DefaultClientID
	}
}

// WrapHandler wraps `http.Handler` into `gin.HandlerFunc`.
func WrapHandler(handler *webdav.Handler, options ...func(*Config)) gin.HandlerFunc {
	var config = Config{
		URL:                      "doc.json",
		DocExpansion:             "list",
		InstanceName:             swag.Name,
		Title:                    "Swagger UI",
		DefaultModelsExpandDepth: 1,
		DeepLinking:              true,
		PersistAuthorization:     false,
		Oauth2DefaultClientID:    "",
	}

	for _, c := range options {
		c(&config)
	}

	return CustomWrapHandler(&config, handler)
}

// CustomWrapHandler wraps `http.Handler` into `gin.HandlerFunc`.
func CustomWrapHandler(config *Config, handler *webdav.Handler) gin.HandlerFunc {
	var once sync.Once

	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}

	if config.Title == "" {
		config.Title = "Swagger UI"
	}

	// create a template with name

	var matcher = regexp.MustCompile(`(.*)(index\.html|index\.css|swagger-initializer\.js|doc\.json|favicon-16x16\.png|favicon-32x32\.png|/oauth2-redirect\.html|swagger-ui\.css|swagger-ui\.css\.map|swagger-ui\.js|swagger-ui\.js\.map|swagger-ui-bundle\.js|swagger-ui-bundle\.js\.map|swagger-ui-standalone-preset\.js|swagger-ui-standalone-preset\.js\.map)[?|.]*`)

	return func(ctx *gin.Context) {
		if ctx.Request.Method != http.MethodGet {
			ctx.AbortWithStatus(http.StatusMethodNotAllowed)

			return
		}

		matches := matcher.FindStringSubmatch(ctx.Request.RequestURI)

		if len(matches) != 3 {
			ctx.String(http.StatusNotFound, http.StatusText(http.StatusNotFound))

			return
		}

		path := matches[2]
		once.Do(func() {
			handler.Prefix = matches[1]
		})

		switch filepath.Ext(path) {
		case ".html":
			ctx.Header("Content-Type", "text/html; charset=utf-8")
		case ".css":
			ctx.Header("Content-Type", "text/css; charset=utf-8")
		case ".js":
			ctx.Header("Content-Type", "application/javascript")
		case ".png":
			ctx.Header("Content-Type", "image/png")
		case ".json":
			ctx.Header("Content-Type", "application/json; charset=utf-8")
		}

	}
}

// DisablingWrapHandler turn handler off
// if specified environment variable passed.
func DisablingWrapHandler(handler *webdav.Handler, envName string) gin.HandlerFunc {
	if os.Getenv(envName) != "" {
		return func(c *gin.Context) {
			// Simulate behavior when route unspecified and
			// return 404 HTTP code
			c.String(http.StatusNotFound, "")
		}
	}

	return WrapHandler(handler)
}

// DisablingCustomWrapHandler turn handler off
// if specified environment variable passed.
func DisablingCustomWrapHandler(config *Config, handler *webdav.Handler, envName string) gin.HandlerFunc {
	if os.Getenv(envName) != "" {
		return func(c *gin.Context) {
			// Simulate behavior when route unspecified and
			// return 404 HTTP code
			c.String(http.StatusNotFound, "")
		}
	}

	return CustomWrapHandler(config, handler)
}
