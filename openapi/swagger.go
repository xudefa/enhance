package openapi

import (
	"fmt"
	"net/http"
)

// ServeSwaggerUI 提供 Swagger UI 服务
func ServeSwaggerUI(doc *DocumentBuilder, basePath string, port int) error {
	mux := http.NewServeMux()

	// OpenAPI JSON 端点
	mux.Handle(basePath+"/openapi.json", doc)

	// Swagger UI HTML
	mux.HandleFunc(basePath+"/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprint(w, swaggerUIHTML(basePath+"/openapi.json"))
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Swagger UI available at http://localhost%s%s/\n", addr, basePath)
	fmt.Printf("OpenAPI JSON at http://localhost%s%s/openapi.json\n", addr, basePath)

	return http.ListenAndServe(addr, mux)
}

// swaggerUIHTML Swagger UI HTML 模板
func swaggerUIHTML(openAPIURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '%s',
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: 'BaseLayout'
            });
        };
    </script>
</body>
</html>`, openAPIURL)
}
