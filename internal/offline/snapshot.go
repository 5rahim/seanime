package offline

type (
	NewSnapshotOptions struct {
		MediaIds []int
	}
)

// Snapshot populates offline data
func (h *Hub) Snapshot(opts *NewSnapshotOptions) error {

	// Use NewMediaEntry (anime)
	// Modify NewMediaEntry, so we don't concern ourselves with the DownloadInfo

	panic("not implemented")
}
