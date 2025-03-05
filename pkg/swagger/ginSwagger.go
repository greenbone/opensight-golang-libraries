package ginSwagger

import (
	"html/template"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files/v2"
	"github.com/swaggo/swag"
	"sigs.k8s.io/yaml"
)

// Config stores ginSwagger configuration variables.
type Config struct {
	URLs                 []string
	DocExpansion         string
	DomID                string
	InstanceName         string
	DeepLinking          bool
	PersistAuthorization bool
	SyntaxHighlight      bool
	OAuth                *OAuthConfig
}

type OAuthConfig struct {
	ClientId string
	Realm    string
	AppName  string
}

func URL(url string) func(*Config) {
	return func(c *Config) {
		c.URLs = append(c.URLs, url)
	}
}

func DeepLinking(deepLinking bool) func(*Config) {
	return func(c *Config) {
		c.DeepLinking = deepLinking
	}
}

func SyntaxHighlight(syntaxHighlight bool) func(*Config) {
	return func(c *Config) {
		c.SyntaxHighlight = syntaxHighlight
	}
}

func DocExpansion(docExpansion string) func(*Config) {
	return func(c *Config) {
		c.DocExpansion = docExpansion
	}
}

func DomID(domID string) func(*Config) {
	return func(c *Config) {
		c.DomID = domID
	}
}

func InstanceName(instanceName string) func(*Config) {
	return func(c *Config) {
		c.InstanceName = instanceName
	}
}

func PersistAuthorization(persistAuthorization bool) func(*Config) {
	return func(c *Config) {
		c.PersistAuthorization = persistAuthorization
	}
}

func OAuth(config *OAuthConfig) func(*Config) {
	return func(c *Config) {
		c.OAuth = config
	}
}

func newConfig(configFns ...func(*Config)) *Config {
	config := Config{
		URLs:                 []string{"doc.json", "doc.yaml"},
		DocExpansion:         "list",
		DomID:                "swagger-ui",
		InstanceName:         "swagger",
		DeepLinking:          true,
		PersistAuthorization: false,
		SyntaxHighlight:      true,
	}
	for _, fn := range configFns {
		fn(&config)
	}
	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}
	return &config
}

var WrapHandler = GinWrapHandler()

func GinWrapHandler(options ...func(*Config)) gin.HandlerFunc {
	config := newConfig(options...)
	index, _ := template.New("swagger_index.html").Parse(indexTemplate)
	re := regexp.MustCompile(`^(.*/)([^?].*)?[?|.]*$`)

	return func(c *gin.Context) {
		// Set security headers to protect against ClickJacking and XSS
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; frame-ancestors 'none'")
		c.Header("X-Content-Type-Options", "nosniff")

		if c.Request.Method != http.MethodGet {
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		matches := re.FindStringSubmatch(c.Request.RequestURI)
		path := matches[2]

		switch filepath.Ext(path) {
		case ".html":
			c.Header("Content-Type", "text/html; charset=utf-8")
		case ".css":
			c.Header("Content-Type", "text/css; charset=utf-8")
		case ".js":
			c.Header("Content-Type", "application/javascript")
		case ".json":
			c.Header("Content-Type", "application/json; charset=utf-8")
		case ".yaml":
			c.Header("Content-Type", "text/plain; charset=utf-8")
		case ".png":
			c.Header("Content-Type", "image/png")
		}

		response := c.Writer
		if _, ok := response.(http.Flusher); ok {
			defer response.(http.Flusher).Flush()
		}

		switch path {
		case "index.html":
			_ = index.Execute(c.Writer, config)
		case "doc.json":
			doc, err := swag.ReadDoc(config.InstanceName)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err) //nolint:errcheck
				return
			}
			c.String(http.StatusOK, doc)
		case "doc.yaml":
			jsonString, err := swag.ReadDoc(config.InstanceName)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err) //nolint:errcheck
				return
			}
			doc, err := yaml.JSONToYAML([]byte(jsonString))
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err) //nolint:errcheck
				return
			}
			c.String(http.StatusOK, string(doc))
		default:
			c.Request.URL.Path = matches[2]
			http.FileServer(http.FS(swaggerFiles.FS)).ServeHTTP(c.Writer, c.Request)
		}
	}
}

const indexTemplate = `<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
  <style>
    body { margin:0; background: #fafafa; }
  </style>
</head>
<body>
<div id="{{.DomID}}"></div>
<script src="./swagger-ui-bundle.js"></script>
<script src="./swagger-ui-standalone-preset.js"></script>
<script>
window.onload = function() {
  const ui = SwaggerUIBundle({
    urls: [
    {{range $index, $url := .URLs}}
      { name: "{{$url}}", url: "{{$url}}" },
    {{end}}
    ],
    syntaxHighlight: {{.SyntaxHighlight}},
    deepLinking: {{.DeepLinking}},
    docExpansion: "{{.DocExpansion}}",
    persistAuthorization: {{.PersistAuthorization}},
    dom_id: "#{{.DomID}}",
    validatorUrl: null,
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
    plugins: [SwaggerUIBundle.plugins.DownloadUrl],
    layout: "StandaloneLayout"
  });
  {{if .OAuth}}
  ui.initOAuth({ clientId: "{{.OAuth.ClientId}}", realm: "{{.OAuth.Realm}}", appName: "{{.OAuth.AppName}}" });
  {{end}}
  window.ui = ui;
}
</script>
</body>
</html>`
