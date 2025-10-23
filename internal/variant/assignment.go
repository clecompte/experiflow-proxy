package variant

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/experiflow/proxy/internal/transform"
)

// Assigner handles variant assignment logic
type Assigner struct {
	salt string
}

// NewAssigner creates a new variant assigner
func NewAssigner(salt string) *Assigner {
	if salt == "" {
		salt = "default-salt-change-in-production"
	}
	return &Assigner{salt: salt}
}

// AssignVariant deterministically assigns a variant to a user
// Uses HMAC-based bucketing for consistent assignment
func (a *Assigner) AssignVariant(userID, experimentID string, variants []transform.Variant) *transform.Variant {
	if len(variants) == 0 {
		return nil
	}

	// If only one variant, return it
	if len(variants) == 1 {
		return &variants[0]
	}

	// Create deterministic bucket
	bucket := a.getBucket(userID, experimentID)

	// Assign based on traffic allocation
	cumulative := 0.0
	for i := range variants {
		cumulative += variants[i].TrafficAllocation
		if float64(bucket) < cumulative*100 {
			return &variants[i]
		}
	}

	// Fallback to first variant (should not reach here if allocations sum to 1.0)
	return &variants[0]
}

// SelectRandomVariant randomly selects a variant (for new users)
func (a *Assigner) SelectRandomVariant(variants []transform.Variant) *transform.Variant {
	if len(variants) == 0 {
		return nil
	}

	if len(variants) == 1 {
		return &variants[0]
	}

	// Random selection based on traffic allocation
	r := rand.Float64()
	cumulative := 0.0

	for i := range variants {
		cumulative += variants[i].TrafficAllocation
		if r < cumulative {
			return &variants[i]
		}
	}

	// Fallback
	return &variants[0]
}

// getBucket returns a deterministic bucket (0-99) for the user+experiment
func (a *Assigner) getBucket(userID, experimentID string) int {
	// Create HMAC hash
	h := hmac.New(sha256.New, []byte(a.salt))
	h.Write([]byte(fmt.Sprintf("%s:%s", userID, experimentID)))
	hash := hex.EncodeToString(h.Sum(nil))

	// Take first 8 characters and convert to int
	hashInt, err := strconv.ParseUint(hash[:8], 16, 64)
	if err != nil {
		// Fallback to random if parsing fails
		return rand.Intn(100)
	}

	// Return bucket 0-99
	return int(hashInt % 100)
}

// GetUserID generates a user ID from request context
// Priority: Cookie > IP + User-Agent hash > Random
func GetUserID(cookieValue, ipAddress, userAgent string) string {
	if cookieValue != "" {
		return cookieValue
	}

	// Hash IP + User-Agent for semi-stable ID
	if ipAddress != "" || userAgent != "" {
		h := sha256.New()
		h.Write([]byte(ipAddress + ":" + userAgent))
		return hex.EncodeToString(h.Sum(nil))[:16]
	}

	// Fallback to random (not ideal but better than nothing)
	return fmt.Sprintf("anon_%d", rand.Intn(1000000))
}
