package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var supportedEffects = map[string]bool{
	"notify_plan_high_risk":    true,
	"notify_checkin_high_risk": true,
	"log_plan_completed":       true,
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
		port = "8204"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 || strings.EqualFold(jwtSecret, "changeme") {
		log.Fatal("JWT_SECRET must be at least 32 characters and must not be a placeholder")
	}
	store := newIdempotencyStore()

	effectsList := make([]string, 0, len(supportedEffects))
	for name := range supportedEffects {
		effectsList = append(effectsList, name)
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":            "ok",
			"plugin":            "care-alerts",
			"supported_effects": effectsList,
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
		log.Printf("[care-alerts] tenant=%s effect=%s entity=%s/%s transition=%s payload=%v",
			req.TenantID, req.EffectName, req.EntityType, req.EntityID, req.Transition, payload)
		writeSuccess(w)
	})

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("care-alerts plugin listening on :%s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("care-alerts error: %v", err)
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
