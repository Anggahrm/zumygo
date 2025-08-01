package downloader

import (
	"fmt"
	"strings"
	"zumygo/libs"
	"zumygo/systems"
)

func init() {
	libs.NewCommands(&libs.ICommand{
		Name:        "(tiktok|ttdl|tiktokdl|tiktoknowm|tt)",
		As:          []string{"tiktok"},
		Tags:        "downloader",
		IsPrefix:    true,
		IsQuery:     true,
		Description: "Download TikTok videos without watermark",
		Execute: func(conn *libs.IClient, m *libs.IMessage) bool {
			// Check if URL is provided
			if len(m.Args) == 0 {
				m.Reply("✳️ Masukkan tautan Tiktok\n\n 📌 Contoh : .tiktok https://vm.tiktok.com/ZMYG92bUh/")
				return false
			}

			url := m.Args[0]

			// Validate TikTok URL
			if !strings.Contains(strings.ToLower(url), "tiktok") {
				m.Reply("❎ Verifikasi bahwa tautan tersebut berasal dari TikTok")
				return false
			}

			// Send processing reaction
			m.React("⏱️")

			// Get downloader system from global systems
			downloaderSystem := systems.GetGlobalDownloaderSystem()
			if downloaderSystem == nil {
				m.Reply("❎ Downloader system not available")
				return false
			}

			// Download TikTok video
			result, err := downloaderSystem.DownloadMedia("tiktok", url)
			if err != nil {
				m.Reply("❎ Kesalahan mengunduh video")
				return false
			}

			if !result.Success {
				m.Reply("❎ " + result.Error)
				return false
			}

			// Download video data
			videoData, err := conn.GetBytes(result.URL)
			if err != nil {
				m.Reply("❎ Gagal mengunduh data video")
				return false
			}

			// Create caption
			caption := fmt.Sprintf(`┌─⊷ TIKTOK
▢ *Deskripsi:* %s
└───────────`, result.Title)

			// Send video file
			_, err = conn.SendVideo(m.Info.Chat, videoData, caption, nil)
			if err != nil {
				m.Reply("❎ Gagal mengirim video")
				return false
			}

			// Send success reaction
			m.React("✅")
			return true
		},
	})
} 