package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/experiflow/proxy/internal/config"
	"github.com/experiflow/proxy/internal/transform"
	"github.com/experiflow/proxy/internal/variant"
	"golang.org/x/net/html"
)

// ExperiFlowMiddleware handles A/B testing transformations
type ExperiFlowMiddleware struct {
	config       *config.Config
	client       *transform.Client
	assigner     *variant.Assigner
	experiments  map[string]bool // Active experiment IDs
}

// NewExperiFlowMiddleware creates a new middleware instance
func NewExperiFlowMiddleware(cfg *config.Config, experimentIDs []string) *ExperiFlowMiddleware {
	experiments := make(map[string]bool)
	for _, id := range experimentIDs {
		experiments[id] = true
	}

	return &ExperiFlowMiddleware{
		config:      cfg,
		client:      transform.NewClient(cfg.APIBaseURL, cfg.EdgeToken, cfg.Timeout),
		assigner:    variant.NewAssigner("production-salt"), // TODO: Move to config
		experiments: experiments,
	}
}

// ModifyResponse transforms the HTML response
func (m *ExperiFlowMiddleware) ModifyResponse(resp *http.Response, req *http.Request) error {
	startTime := time.Now()

	// Only transform HTML responses
	if !m.isHTML(resp) {
		return nil
	}

	// For each active experiment, apply transformations
	for experimentID := range m.experiments {
		if err := m.applyExperiment(resp, req, experimentID, startTime); err != nil {
			if m.config.EnableLogging {
				log.Printf("[ExperiFlow] Error applying experiment %s: %v", experimentID, err)
			}

			// Fail open: continue without transformation if configured
			if m.config.FailOpen {
				return nil
			}
			return err
		}
	}

	return nil
}

// applyExperiment applies a single experiment's transformations
func (m *ExperiFlowMiddleware) applyExperiment(resp *http.Response, req *http.Request, experimentID string, startTime time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.Timeout)
	defer cancel()

	// 1. Get or assign variant
	cookieName := fmt.Sprintf("ef_var_%s", experimentID)
	variantID, variantKey, isNew := m.getOrAssignVariant(ctx, req, experimentID, cookieName)
	if variantID == "" {
		return fmt.Errorf("failed to assign variant")
	}

	// 2. Set cookie if new assignment
	if isNew {
		cookie := &http.Cookie{
			Name:     cookieName,
			Value:    variantID,
			MaxAge:   30 * 24 * 60 * 60, // 30 days
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		// Add cookie to response headers
		if resp.Header.Get("Set-Cookie") == "" {
			resp.Header.Set("Set-Cookie", cookie.String())
		} else {
			resp.Header.Add("Set-Cookie", cookie.String())
		}
	}

	// 3. Fetch transform spec
	spec, err := m.client.GetTransformSpec(ctx, experimentID, variantID)
	if err != nil {
		return fmt.Errorf("fetch transform spec: %w", err)
	}

	// If no operations (control variant), skip transformation
	if len(spec.Operations) == 0 {
		if m.config.EnableLogging {
			log.Printf("[ExperiFlow] Control variant - no transformations applied")
		}
		m.addHeaders(resp, experimentID, variantKey, "control", startTime)
		return nil
	}

	// 4. Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	resp.Body.Close()

	// 5. Parse HTML
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("parse HTML: %w", err)
	}

	// 6. Apply transformations
	if err := transform.ApplyTransformations(doc, spec.Operations); err != nil {
		return fmt.Errorf("apply transformations: %w", err)
	}

	// 7. Render transformed HTML
	transformed, err := transform.RenderHTML(doc)
	if err != nil {
		return fmt.Errorf("render HTML: %w", err)
	}

	// 8. Update response with transformed HTML
	transformedBody := []byte(transformed)
	resp.Body = io.NopCloser(bytes.NewReader(transformedBody))
	resp.ContentLength = int64(len(transformedBody))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(transformedBody)))

	// 9. Add observability headers
	m.addHeaders(resp, experimentID, variantKey, "hit", startTime)

	if m.config.EnableLogging {
		log.Printf("[ExperiFlow] Applied %d transformations for variant %s (took %v)",
			len(spec.Operations), variantKey, time.Since(startTime))
	}

	return nil
}

// getOrAssignVariant gets existing variant from cookie or assigns a new one
func (m *ExperiFlowMiddleware) getOrAssignVariant(ctx context.Context, req *http.Request, experimentID, cookieName string) (string, string, bool) {
	// Check for existing assignment in cookie
	if cookie, err := req.Cookie(cookieName); err == nil && cookie.Value != "" {
		// TODO: Also store variant key in cookie to avoid lookup
		return cookie.Value, "", false
	}

	// New assignment needed - fetch variants
	variants, err := m.client.GetVariants(ctx, experimentID)
	if err != nil {
		if m.config.EnableLogging {
			log.Printf("[ExperiFlow] Failed to fetch variants: %v", err)
		}
		return "", "", false
	}

	if len(variants) == 0 {
		if m.config.EnableLogging {
			log.Printf("[ExperiFlow] No variants found for experiment %s", experimentID)
		}
		return "", "", false
	}

	// Generate user ID
	userID := variant.GetUserID("", req.RemoteAddr, req.UserAgent())

	// Assign variant
	assigned := m.assigner.AssignVariant(userID, experimentID, variants)
	if assigned == nil {
		return "", "", false
	}

	if m.config.EnableLogging {
		log.Printf("[ExperiFlow] Assigned user to variant: %s (control: %v)", assigned.Name, assigned.IsControl)
	}

	return assigned.ID, assigned.Name, true
}

// isHTML checks if the response is HTML
func (m *ExperiFlowMiddleware) isHTML(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "text/html")
}

// addHeaders adds observability headers to the response
func (m *ExperiFlowMiddleware) addHeaders(resp *http.Response, experimentID, variantKey, status string, startTime time.Time) {
	resp.Header.Set("X-EF-Experiment", experimentID)
	resp.Header.Set("X-EF-Variant", variantKey)
	resp.Header.Set("X-EF-Transform", status)
	resp.Header.Set("X-EF-Timing", fmt.Sprintf("total=%dms", time.Since(startTime).Milliseconds()))
}
