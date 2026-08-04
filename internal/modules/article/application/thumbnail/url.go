package thumbnail

import (
	"strings"
)

const (
	PrefixS3  = "s3:"
	PrefixExt = "ext:"
)

// ToStorage converts a raw URL/key to storage format with appropriate prefix.
// - If it looks like an external URL (starts with http/https), prefix with "ext:"
// - Otherwise, assume it's an S3 key and prefix with "s3:"
// - Returns nil if input is nil or empty
func ToStorage(raw *string) *string {
	if raw == nil {
		return nil
	}

	v := strings.TrimSpace(*raw)
	if v == "" {
		return nil
	}

	// Already has prefix
	if strings.HasPrefix(v, PrefixS3) || strings.HasPrefix(v, PrefixExt) {
		return &v
	}

	// External URL
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		prefixed := PrefixExt + v
		return &prefixed
	}

	// S3 key
	prefixed := PrefixS3 + v
	return &prefixed
}

// FromStorage converts a storage-formatted URL back to a public URL.
// - "s3:key" -> resolved via resolver callback
// - "ext:url" -> returned as-is (the URL itself)
// - Returns nil if input is nil or empty
func FromStorage(stored *string, s3Resolver func(key string) string) *string {
	if stored == nil {
		return nil
	}

	v := strings.TrimSpace(*stored)
	if v == "" {
		return nil
	}

	// S3 key - resolve via callback
	if strings.HasPrefix(v, PrefixS3) {
		key := strings.TrimPrefix(v, PrefixS3)
		if s3Resolver != nil && key != "" {
			resolved := s3Resolver(key)
			return &resolved
		}
		return &v // Return as-is if no resolver
	}

	// External URL - return as-is
	if strings.HasPrefix(v, PrefixExt) {
		extURL := strings.TrimPrefix(v, PrefixExt)
		return &extURL
	}

	// Legacy format without prefix - assume S3 key
	if s3Resolver != nil {
		resolved := s3Resolver(v)
		return &resolved
	}

	return &v
}

// IsExternal checks if a stored URL is an external URL
func IsExternal(stored *string) bool {
	if stored == nil {
		return false
	}
	return strings.HasPrefix(*stored, PrefixExt)
}

// IsS3 checks if a stored URL is an S3 key
func IsS3(stored *string) bool {
	if stored == nil {
		return false
	}
	return strings.HasPrefix(*stored, PrefixS3)
}

// ExtractKey extracts the actual key/URL from a stored value
func ExtractKey(stored *string) string {
	if stored == nil {
		return ""
	}

	v := *stored
	if strings.HasPrefix(v, PrefixS3) {
		return strings.TrimPrefix(v, PrefixS3)
	}
	if strings.HasPrefix(v, PrefixExt) {
		return strings.TrimPrefix(v, PrefixExt)
	}
	return v
}
