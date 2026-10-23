package torrstor

import (
	"testing"

	"github.com/anacrolix/torrent/metainfo"
)

func TestPinnedPieceIsNotEvictionCandidate(t *testing.T) {
	initTestSettings(t)
	storage := NewStorage(1)
	var hash metainfo.Hash
	impl, err := storage.OpenTorrent(testInfo(), hash)
	if err != nil {
		t.Fatal(err)
	}
	cache := impl.(*Cache)
	t.Cleanup(func() { _ = cache.Close() })
	cache.pieces[0].Size = cache.pieceLength
	cache.pieces[1].Size = cache.pieceLength
	cache.SetPinnedPieces(map[int]bool{0: true})
	candidates := cache.getRemPieces()
	for _, piece := range candidates {
		if piece.Id == 0 {
			t.Fatal("pinned piece selected for eviction")
		}
	}
	if len(candidates) != 1 || candidates[0].Id != 1 {
		t.Fatalf("unexpected eviction candidates: %v", candidates)
	}
}
