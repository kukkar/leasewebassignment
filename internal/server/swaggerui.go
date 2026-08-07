package server

import (
	"net/http"

	swaggerfiles "github.com/swaggo/files"
)

// swaggerIndexHTML is a deliberately minimal Swagger UI page - just enough
// to load the bundle assets from swaggerfiles.HTTP and point them at our
// hand-written spec. We don't pull in swaggo/http-swagger for this: that
// package drags in swaggo/swag (and, transitively, go-openapi/* and
// golang.org/x/tools) purely to support generating a spec from source
// annotations, a feature this project doesn't use - the spec is
// hand-written at docs/openapi.yaml. swaggo/files alone (the actual
// swagger-ui-dist static assets) has a single lightweight dependency.
const swaggerIndexHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Leaseweb Assignment API</title>
  <link rel="stylesheet" href="./swagger-ui.css">
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="./swagger-ui-bundle.js"></script>
  <script src="./swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
        docExpansion: "list",
        validatorUrl: null,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        plugins: [SwaggerUIBundle.plugins.DownloadUrl],
        layout: "StandaloneLayout"
      });
    };
  </script>
</body>
</html>
`

// newSwaggerUIHandler serves the interactive API docs. It's mounted at
// /docs/ (see routes.go) and expects to have that prefix already stripped.
func newSwaggerUIHandler() http.Handler {
	assets := http.FileServer(swaggerfiles.HTTP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" || r.URL.Path == "/" || r.URL.Path == "/index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(swaggerIndexHTML))
			return
		}
		assets.ServeHTTP(w, r)
	})
}
