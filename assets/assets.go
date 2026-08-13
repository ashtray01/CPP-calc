// Package assets 提供编译进可执行文件的照片和音乐资源。
package assets

import "embed"

// FS 包含主席照片 leader.png 和主席模式音乐 music.mp3。
//
//go:embed leader.png
//go:embed music.mp3
var FS embed.FS
