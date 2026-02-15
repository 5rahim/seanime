package scanner

import (
	"testing"
)

// BenchmarkNormalizeTitle benchmarks the title normalization
func BenchmarkNormalizeTitle(b *testing.B) {
	titles := []string{
		"Attack on Titan Season 2",
		"Kono Subarashii Sekai ni Shukufuku wo! 2",
		"Boku no Hero Academia 5th Season",
		"[SubsPlease] Mushoku Tensei S2 - 01 (1080p) [EC64C8B1].mkv",
		"Overlord III",
		"Steins;Gate 0",
		"Jujutsu Kaisen 2nd Season",
		"86 - Eighty Six Part 2",
		"The Melancholy of Haruhi Suzumiya (2009)",
		"KonoSuba.God's.Blessing.On.This.Wonderful.World.S02E01.1080p.BluRay.10-Bit.Dual-Audio.FLAC2.0.x265-YURASUKA",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, title := range titles {
			NormalizeTitle(title)
		}
	}
}

// BenchmarkNormalizeTitleParallel benchmarks parallel title normalization
func BenchmarkNormalizeTitleParallel(b *testing.B) {
	title := "Kono Subarashii Sekai ni Shukufuku wo! Season 2 Part 1"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NormalizeTitle(title)
		}
	})
}
