package offline

import "github.com/seanime-app/seanime/internal/api/anilist"

type (
	MockHub struct {
	}
)

func NewMockHub() *MockHub {
	return &MockHub{}
}

func (h *MockHub) RetrieveCurrentSnapshot() (ret *Snapshot, ok bool) {
	return nil, false
}

func (h *MockHub) GetCurrentSnapshot() (ret *Snapshot, ok bool) {
	return nil, false
}

func (h *MockHub) UpdateAnimeListStatus(mediaId int, progress int, status anilist.MediaListStatus) (err error) {
	return nil
}

func (h *MockHub) UpdateEntryListData(mediaId *int, status *anilist.MediaListStatus, score *int, progress *int, startDate *string, endDate *string, t string) (err error) {
	return nil
}

func (h *MockHub) UpdateMangaListStatus(mediaId int, progress int, status anilist.MediaListStatus) (err error) {
	return nil
}

func (h *MockHub) SyncListData() error {
	return nil
}

func (h *MockHub) CreateSnapshot(opts *NewSnapshotOptions) error {
	return nil
}

func (h *MockHub) GetLatestSnapshotEntry() (snapshotEntry *SnapshotEntry, err error) {
	return &SnapshotEntry{}, nil
}

func (h *MockHub) GetLatestSnapshot(bypassCache bool) (snapshot *Snapshot, err error) {
	return &Snapshot{}, nil
}
