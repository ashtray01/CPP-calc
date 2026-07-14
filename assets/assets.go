// Package assets provides embedded binary resources (image, music).
//
// Files are embedded into the binary at compile time via //go:embed.
package assets

import "embed"

// FS contains all embedded project assets.
//
// Available files:
//   - leader.png       — фоновое изображение с портретом Мао Цзэдуна
//   - music.mp3        — «Красное солнце в небе» (东方红)
//
//go:embed leader.png
//go:embed music.mp3
var FS embed.FS
