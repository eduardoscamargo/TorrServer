package torr

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/anacrolix/torrent"

	"server/settings"
	utils2 "server/utils"
)

func sortedTorrentFiles(t *Torrent) []*torrent.File {
	files := append([]*torrent.File(nil), t.Files()...)
	sort.Slice(files, func(i, j int) bool { return utils2.CompareStrings(files[i].Path(), files[j].Path()) })
	return files
}

func pinnedPieces(files []*torrent.File, pins map[int]bool, pieceLength int64) map[int]bool {
	ids := make(map[int]bool)
	for i, file := range files {
		if !pins[i+1] || file.Length() == 0 {
			continue
		}
		for id := int(file.Offset() / pieceLength); id <= int((file.Offset()+file.Length()-1)/pieceLength); id++ {
			ids[id] = true
		}
	}
	return ids
}

func (t *Torrent) restorePins() {
	if t.cache == nil || !settings.BTsets.UseDisk {
		return
	}
	t.muTorrent.Lock()
	alreadyLoaded := t.pinnedFiles != nil
	t.muTorrent.Unlock()
	if alreadyLoaded {
		return
	}
	db := GetTorrentDB(t.Hash())
	if db == nil {
		return
	}
	pins := make(map[int]bool)
	db.muTorrent.Lock()
	for id := range db.pinnedFiles {
		pins[id] = true
	}
	db.muTorrent.Unlock()
	files := sortedTorrentFiles(t)
	t.muTorrent.Lock()
	t.pinnedFiles = pins
	t.muTorrent.Unlock()
	t.cache.SetPinnedPieces(pinnedPieces(files, pins, t.Info().PieceLength))
}

// PinFile retains a selected file on disk and downloads its missing pieces.
// File IDs use the same path-sorted, one-based numbering as /play.
func (t *Torrent) PinFile(fileID int, pin bool) error {
	if !settings.BTsets.UseDisk || settings.BTsets.TorrentsSavePath == "" {
		return errors.New("pinning requires disk cache storage")
	}
	if !t.GotInfo() || t.cache == nil {
		return errors.New("torrent metadata unavailable")
	}
	files := sortedTorrentFiles(t)
	if fileID < 1 || fileID > len(files) {
		return errors.New("file not found")
	}
	if pin {
		dir := filepath.Join(settings.BTsets.TorrentsSavePath, t.Hash().HexString())
		if err := os.MkdirAll(dir, 0o777); err != nil {
			return err
		}
		probe, err := os.CreateTemp(dir, ".pin-write-check-")
		if err != nil {
			return err
		}
		_ = probe.Close()
		_ = os.Remove(probe.Name())
	}
	t.muTorrent.Lock()
	if t.pinnedFiles == nil {
		t.pinnedFiles = make(map[int]bool)
	}
	if pin {
		t.pinnedFiles[fileID] = true
	} else {
		delete(t.pinnedFiles, fileID)
	}
	pins := make(map[int]bool, len(t.pinnedFiles))
	for id := range t.pinnedFiles {
		pins[id] = true
	}
	t.muTorrent.Unlock()
	t.cache.SetPinnedPieces(pinnedPieces(files, pins, t.Info().PieceLength))
	AddTorrentDB(t)
	return nil
}

func restorePinnedTorrents(bt *BTServer) {
	for _, db := range settings.ListPinnedTorrents() {
		if db.TorrentSpec == nil {
			continue
		}
		tor, err := NewTorrent(db.TorrentSpec, bt)
		if err != nil {
			continue
		}
		tor.Title, tor.Poster, tor.Category, tor.Data = db.Title, db.Poster, db.Category, db.Data
		if !tor.GotInfo() {
			continue
		}
	}
}
