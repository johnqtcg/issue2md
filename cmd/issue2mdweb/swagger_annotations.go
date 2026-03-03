package main

// @Summary Convert GitHub URL to Markdown
// @Description Fetch one GitHub issue, pull request, or discussion and render it as markdown.
// @Tags convert
// @Accept application/x-www-form-urlencoded
// @Produce plain
// @Param url formData string true "GitHub issue/pull/discussion URL"
// @Success 200 {string} string "markdown body"
// @Failure 400 {string} string "invalid request"
// @Failure 401 {string} string "unauthorized"
// @Failure 403 {string} string "forbidden"
// @Failure 404 {string} string "resource not found"
// @Failure 429 {string} string "rate limited"
// @Failure 502 {string} string "upstream failure"
// @Failure 500 {string} string "render failed"
// @Router /convert [post]
func swaggerConvertDoc() {} //nolint:unused // used by swag parser as a route annotation anchor.

// @Summary Get OpenAPI specification
// @Description Returns generated OpenAPI JSON. Run `make swagger` before calling this endpoint.
// @Tags docs
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {string} string "spec unavailable"
// @Failure 500 {string} string "read failed"
// @Router /openapi.json [get]
func swaggerOpenAPISpecDoc() {} //nolint:unused // used by swag parser as a route annotation anchor.
