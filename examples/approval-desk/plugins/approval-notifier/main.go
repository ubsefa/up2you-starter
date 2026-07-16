package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var supportedEffects = map[string]bool{
	"notify_request_approved": true,
	"notify_request_rejected": true,
}

type effectRequest struct {
	EffectName string         `json:"effect_name"`
	Action     string         `json:"action"`
	TenantID   string         `json:"tenant_id"`
	EntityID   string         `json:"entity_id"`
	EntityType string         `json:"entity_type"`
	EventID    string         `json:"event_id"`
	Transition string         `json:"transition"`
	Payload    map[string]any `json:"payload"`
}

type effectResponse struct {
	Success      bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
	ShouldRetry  bool   `json:"should_retry,omitempty"`
}

type idempotencyStore struct {
	mu   sync.Mutex
	seen map[string]time.Time
}

func newIdempotencyStore() *idempotencyStore {
	return &idempotencyStore{seen: map[string]time.Time{}}
}

func (s *idempotencyStore) seenBefore(key string) bool {
	if key == "" {
		return false
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, ts := range s.seen {
		if now.Sub(ts) > 24*time.Hour {
			delete(s.seen, k)
		}
	}
	if _, ok := s.seen[key]; ok {
		return true
	}
	s.seen[key] = now
	return false
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8202"
	}
	jwtSecret := os.Getenv("PLUGIN_EXECUTION_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = os.Getenv("JWT_SECRET")
	}
	if len(jwtSecret) < 32 || strings.EqualFold(jwtSecret, "changeme") {
		log.Fatal("JWT_SECRET must be at least 32 characters and must not be a placeholder")
	}
	webhookURL := os.Getenv("WEBHOOK_URL")
	httpClient := &http.Client{Timeout: 8 * time.Second}
	store := newIdempotencyStore()
	effectsList := make([]string, 0, len(supportedEffects))
	for name := range supportedEffects {
		effectsList = append(effectsList, name)
	}
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":             "ok",
			"plugin":             "approval-notifier",
			"webhook_configured": webhookURL != "",
			"supported_effects":  effectsList,
		})
	})
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed", false)
			return
		}
		if err := verifyBearer(r.Header.Get("Authorization"), []byte(jwtSecret)); err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", false)
			return
		}
		var req effectRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request", false)
			return
		}
		if !supportedEffects[req.EffectName] {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("unknown effect: %s", req.EffectName), false)
			return
		}
		if store.seenBefore(req.EventID) {
			writeSuccess(w)
			return
		}
		payload := req.Payload
		if payload == nil {
			payload = map[string]any{}
		}
		title, _ := payload["request_title"].(string)
		requester, _ := payload["requester_user_id"].(string)
		if webhookURL == "" {
			log.Printf("[approval-notifier] tenant=%s effect=%s request=%s title=%q requester=%s (no webhook configured, logging only)",
				req.TenantID, req.EffectName, req.EntityID, title, requester)
			writeSuccess(w)
			return
		}
		body := map[string]any{
			"effect":     req.EffectName,
			"tenant_id":  req.TenantID,
			"request_id": req.EntityID,
			"transition": req.Transition,
			"event_id":   req.EventID,
			"payload":    payload,
			"sent_at":    time.Now().UTC().Format(time.RFC3339),
		}
		raw, err := json.Marshal(body)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "encode failed", false)
			return
		}
		httpReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, webhookURL, bytes.NewReader(raw))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "webhook request build failed", false)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("X-Up2You-Effect", req.EffectName)
		if req.EventID != "" {
			httpReq.Header.Set("X-Up2You-Event-Id", req.EventID)
		}
		resp, err := httpClient.Do(httpReq)
		if err != nil {
			log.Printf("[approval-notifier] webhook network error: %v", err)
			writeError(w, http.StatusBadGateway, "webhook network error", true)
			return
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
		switch {
		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			log.Printf("[approval-notifier] tenant=%s effect=%s request=%s webhook=%d delivered",
				req.TenantID, req.EffectName, req.EntityID, resp.StatusCode)
			writeSuccess(w)
		case resp.StatusCode == http.StatusRequestTimeout, resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= 500:
			log.Printf("[approval-notifier] webhook transient failure status=%d effect=%s", resp.StatusCode, req.EffectName)
			writeError(w, http.StatusBadGateway, fmt.Sprintf("webhook transient failure: %d", resp.StatusCode), true)
		default:
			log.Printf("[approval-notifier] webhook permanent failure status=%d effect=%s", resp.StatusCode, req.EffectName)
			writeError(w, http.StatusBadGateway, fmt.Sprintf("webhook permanent failure: %d", resp.StatusCode), false)
		}
	})
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	log.Printf("approval-notifier plugin listening on :%s (webhook_configured=%t)", port, webhookURL != "")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("approval-notifier error: %v", err)
	}
}

func writeSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(effectResponse{Success: true})
}

func writeError(w http.ResponseWriter, status int, message string, shouldRetry bool) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(effectResponse{Success: false, ErrorMessage: message, ShouldRetry: shouldRetry})
}

func verifyBearer(header string, secret []byte) error {
	if !strings.HasPrefix(header, "Bearer ") {
		return errors.New("missing bearer token")
	}
	parts := strings.Split(strings.TrimPrefix(header, "Bearer "), ".")
	if len(parts) != 3 {
		return errors.New("invalid token")
	}
	signingInput := parts[0] + "." + parts[1]
	expected := hmacSHA256(signingInput, secret)
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}
	if !hmac.Equal(actual, expected) {
		return errors.New("invalid signature")
	}
	var headerClaims map[string]any
	if err := decodeSegment(parts[0], &headerClaims); err != nil {
		return err
	}
	if alg, _ := headerClaims["alg"].(string); alg != "HS256" {
		return errors.New("unexpected signing method")
	}
	var claims map[string]any
	if err := decodeSegment(parts[1], &claims); err != nil {
		return err
	}
	if iss, _ := claims["iss"].(string); iss != "up2you" {
		return errors.New("invalid issuer")
	}
	if exp, ok := numericClaim(claims["exp"]); ok && time.Now().Unix() >= exp {
		return errors.New("token expired")
	}
	role, _ := claims["role"].(string)
	if role != "admin" && role != "system" {
		return errors.New("forbidden role")
	}
	return nil
}

func hmacSHA256(input string, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(input))
	return mac.Sum(nil)
}

func decodeSegment(segment string, out any) error {
	data, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func numericClaim(value any) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true
	case json.Number:
		n, err := v.Int64()
		return n, err == nil
	default:
		return 0, false
	}
}
