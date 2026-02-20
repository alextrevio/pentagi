package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HTTPHeaderScanner es una herramienta para verificar cabeceras de seguridad HTTP
type HTTPHeaderScanner struct {
	client *http.Client
}

// HTTPHeaderScannerArgs representa los argumentos para el escaneo de cabeceras
type HTTPHeaderScannerArgs struct {
	URL string `json:"url"`
}

// HTTPHeaderScannerResult representa el resultado del escaneo
type HTTPHeaderScannerResult struct {
	URL             string            `json:"url"`
	MissingHeaders  []string          `json:"missing_headers"`
	PresentHeaders  map[string]string `json:"present_headers"`
	SecurityScore   int               `json:"security_score"`
	Recommendations []string          `json:"recommendations"`
}

// SecurityHeaders define las cabeceras de seguridad recomendadas
var SecurityHeaders = []string{
	"Content-Security-Policy",
	"Strict-Transport-Security",
	"X-Frame-Options",
	"X-Content-Type-Options",
	"X-XSS-Protection",
	"Referrer-Policy",
	"Permissions-Policy",
}

// NewHTTPHeaderScanner crea una nueva instancia del escáner
func NewHTTPHeaderScanner() *HTTPHeaderScanner {
	return &HTTPHeaderScanner{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // No seguir redirecciones
			},
		},
	}
}

// Name devuelve el nombre de la herramienta
func (h *HTTPHeaderScanner) Name() string {
	return "http_header_scanner"
}

// Description devuelve la descripción de la herramienta
func (h *HTTPHeaderScanner) Description() string {
	return "Escanea una URL para verificar la presencia de cabeceras de seguridad HTTP recomendadas"
}

// IsAvailable indica si la herramienta está disponible
func (h *HTTPHeaderScanner) IsAvailable() bool {
	return true
}

// Execute ejecuta el escaneo de cabeceras
func (h *HTTPHeaderScanner) Execute(ctx context.Context, args json.RawMessage) (map[string]interface{}, error) {
	var scanArgs HTTPHeaderScannerArgs
	if err := json.Unmarshal(args, &scanArgs); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	if scanArgs.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}

	// Realizar la petición HTTP
	req, err := http.NewRequestWithContext(ctx, "GET", scanArgs.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Analizar las cabeceras de respuesta
	result := h.analyzeHeaders(resp.Header)
	result.URL = scanArgs.URL

	// Convertir el resultado a map para devolverlo
	resultMap := map[string]interface{}{
		"success":          true,
		"url":              result.URL,
		"missing_headers":  result.MissingHeaders,
		"present_headers":  result.PresentHeaders,
		"security_score":   result.SecurityScore,
		"recommendations":  result.Recommendations,
	}

	return resultMap, nil
}

// analyzeHeaders analiza las cabeceras HTTP y genera un reporte
func (h *HTTPHeaderScanner) analyzeHeaders(headers http.Header) *HTTPHeaderScannerResult {
	result := &HTTPHeaderScannerResult{
		MissingHeaders:  make([]string, 0),
		PresentHeaders:  make(map[string]string),
		Recommendations: make([]string, 0),
	}

	// Verificar cada cabecera de seguridad
	for _, header := range SecurityHeaders {
		value := headers.Get(header)
		if value == "" {
			result.MissingHeaders = append(result.MissingHeaders, header)
			result.Recommendations = append(result.Recommendations, h.getRecommendation(header))
		} else {
			result.PresentHeaders[header] = value
		}
	}

	// Calcular puntuación de seguridad (0-100)
	totalHeaders := len(SecurityHeaders)
	presentCount := len(result.PresentHeaders)
	result.SecurityScore = (presentCount * 100) / totalHeaders

	return result
}

// getRecommendation devuelve una recomendación para una cabecera faltante
func (h *HTTPHeaderScanner) getRecommendation(header string) string {
	recommendations := map[string]string{
		"Content-Security-Policy": "Añadir: Content-Security-Policy: default-src 'self'",
		"Strict-Transport-Security": "Añadir: Strict-Transport-Security: max-age=31536000; includeSubDomains",
		"X-Frame-Options": "Añadir: X-Frame-Options: DENY",
		"X-Content-Type-Options": "Añadir: X-Content-Type-Options: nosniff",
		"X-XSS-Protection": "Añadir: X-XSS-Protection: 1; mode=block",
		"Referrer-Policy": "Añadir: Referrer-Policy: no-referrer-when-downgrade",
		"Permissions-Policy": "Añadir: Permissions-Policy: geolocation=(), microphone=(), camera=()",
	}

	if rec, ok := recommendations[header]; ok {
		return rec
	}
	return fmt.Sprintf("Añadir cabecera: %s", header)
}
