package query

import (
	"strings"

	ports "sipon-be/internal/modules/kesantrian/application/ports"
)

// resolveAvatarURL mengubah nilai tersimpan di users.avatar_key menjadi URL
// publik. Nilai berprefix "ext:" (avatar eksternal, mis. picture Google)
// langsung dikembalikan apa adanya; "s3:" (atau legacy tanpa prefix) di-resolve
// lewat fileUploader.
func resolveAvatarURL(fileUploader ports.FileUploader, avatarKey *string) *string {
	if avatarKey == nil || *avatarKey == "" {
		return nil
	}
	if strings.HasPrefix(*avatarKey, "ext:") {
		extURL := strings.TrimPrefix(*avatarKey, "ext:")
		if extURL == "" {
			return nil
		}
		return &extURL
	}
	u := fileUploader.PublicURL(strings.TrimPrefix(*avatarKey, "s3:"))
	return &u
}
