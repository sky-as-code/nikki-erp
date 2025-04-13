package utility

var mime map[string]string = map[string]string{
	"aac":  "audio/x-aac",
	"css":  "text/css",
	"gif":  "image/gif",
	"htm":  "text/html",
	"html": "text/html",
	"ico":  "image/x-icon",
	"jpeg": "image/jpeg",
	"jpg":  "image/jpeg",
	"js":   "application/javascript",
	"json": "application/json",
	"m3u8": "application/vnd.apple.mpegurl",
	"m4a":  "audio/x-m4a",
	"m4v":  "video/x-m4v",
	"mp4":  "video/mp4",
	"png":  "image/png",
	"ts":   "video/mp2t",
	"txt":  "text/plain",
	"vtt":  "text/vtt",
	"wav":  "audio/x-wav",
	"webp": "image/webp",
}

func GetContentType(ext string) string {

	mime, ok := mime[ext]
	if !ok {
		return "application/octet-stream"
	}

	return mime
}
