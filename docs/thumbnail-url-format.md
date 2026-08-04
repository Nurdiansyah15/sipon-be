# Thumbnail URL Format

## Overview

Article thumbnails are stored in the database with a prefix that indicates the storage type. This makes it easy to parse and resolve URLs appropriately.

## Format

```
s3:<key>     - For manually uploaded images stored in MinIO/S3
ext:<url>    - For external/scraped images (direct URLs from RSS feeds)
```

## Examples

### Manual Upload (s3:)
```
Database: s3:articles/2026/01/thumbnail-abc123.jpg
Resolved: https://minio.example.com/sipon/articles/2026/01/thumbnail-abc123.jpg
```

### Scraped from RSS (ext:)
```
Database: ext:https://example.com/images/news-photo.jpg
Resolved: https://example.com/images/news-photo.jpg
```

## Implementation

### Storage (Writing to Database)

Use `thumbnail.ToStorage()` to convert raw URLs to storage format:

```go
import "sipon-be/internal/modules/article/application/thumbnail"

// Manual upload - key from MinIO
stored := thumbnail.ToStorage(&key)  // Returns: "s3:articles/2026/01/image.jpg"

// Scraped from RSS - external URL
stored := thumbnail.ToStorage(&externalURL)  // Returns: "ext:https://example.com/image.jpg"
```

**Auto-detection logic:**
- If URL starts with `http://` or `https://` → prefix with `ext:`
- Otherwise → prefix with `s3:`
- If already has prefix → return as-is

### Resolution (Reading from Database)

Use `thumbnail.FromStorage()` to convert storage format to public URL:

```go
import "sipon-be/internal/modules/article/application/thumbnail"

// With S3 resolver (for manual uploads)
publicURL := thumbnail.FromStorage(storedKey, func(key string) string {
    return fileUploader.PublicURL(key)
})

// For external URLs, the resolver is ignored and the URL is returned as-is
```

### Helper Functions

```go
// Check storage type
thumbnail.IsS3(storedKey)     // true if "s3:..."
thumbnail.IsExternal(storedKey) // true if "ext:..."

// Extract the actual key/URL without prefix
key := thumbnail.ExtractKey(storedKey)  // "s3:articles/img.jpg" → "articles/img.jpg"
url := thumbnail.ExtractKey(storedKey)  // "ext:https://..." → "https://..."
```

## Usage in Codebase

### Create/Update Article (Manual Upload)

```go
// In create_article.go
article, err := articleentity.NewArticle(articleentity.ArticleParams{
    // ...
    ThumbnailURL: thumbnail.ToStorage(req.ThumbnailURL),  // Prefix with s3:
    // ...
})

// Confirm upload to MinIO (only for s3: keys)
confirmThumbnailKey(ctx, uc.fileUploader, article.ThumbnailURL)
```

### Scrape Article (External Image)

```go
// In pipeline.go
var thumbnailURL *string
if it.Thumbnail != nil && *it.Thumbnail != "" {
    thumbnailURL = thumbnail.ToStorage(it.Thumbnail)  // Prefix with ext:
}
```

### Get/List Articles (Resolve URL)

```go
// In get_article.go and list_articles.go
ThumbnailURL: resolveThumbnailURL(uploader, a.ThumbnailURL)

// resolveThumbnailURL uses FromStorage with resolver
func resolveThumbnailURL(uploader ports.FileUploader, storedKey *string) *string {
    var s3Resolver func(key string) string
    if uploader != nil {
        s3Resolver = uploader.PublicURL
    }
    return thumbnail.FromStorage(storedKey, s3Resolver)
}
```

## Migration Notes

Existing articles without prefix will be handled gracefully:
- `resolveThumbnailURL` will treat them as S3 keys (legacy behavior)
- New articles will always have proper prefix

To migrate existing data:
```sql
-- Add s3: prefix to existing thumbnails (if they are S3 keys)
UPDATE articles 
SET thumbnail_url = 's3:' || thumbnail_url 
WHERE thumbnail_url IS NOT NULL 
  AND thumbnail_url NOT LIKE 's3:%' 
  AND thumbnail_url NOT LIKE 'ext:%';
```

## Benefits

1. **Clear separation** - Easy to distinguish between S3 and external images
2. **Simple parsing** - Just check prefix, no complex logic needed
3. **Backward compatible** - Existing code continues to work
4. **Future-proof** - Easy to add new storage types (e.g., `gcs:` for Google Cloud)
5. **Efficient storage** - Store only the key, not full URL for S3 images

## File Location

```
internal/modules/article/application/thumbnail/
└── url.go  # ToStorage, FromStorage, IsS3, IsExternal, ExtractKey
```
