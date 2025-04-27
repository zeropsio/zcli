package uuid

import (
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
	zeropsUuid "github.com/zeropsio/zerops-go/types/uuid"
)

func GetShort() string {
	x := uuid.New()
	return encode(x[:])
}

func encode(uuid []byte) string {
	b64 := base64.RawURLEncoding.EncodeToString(uuid)
	// TODO(tikinang): Fix for 1.22, improve.
	b64 = strings.ReplaceAll(b64, "-", "A")
	b64 = strings.ReplaceAll(b64, "_", "B")
	// Should already be 22 chars, just making sure if creators of base64 package change their mind.
	return b64[:22]
}

// IsValidServiceId checks if a string is a valid Zerops service ID format
func IsValidServiceId(id string) bool {
	// First check if it's a valid UUID format
	_, err := uuid.Parse(id)
	if err != nil {
		return false
	}
	
	// Zerops also validates that the ID represents a service ID
	serviceId := zeropsUuid.ServiceStackId(id)
	return serviceId.Native() != "" 
}
