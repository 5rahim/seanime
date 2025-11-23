package library_explorer

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/platforms/platform"
	"seanime/internal/util"
	"sync"

	"github.com/rs/zerolog"
)

type LibraryExplorer struct {
	mu              sync.RWMutex
	animeCollection *anilist.AnimeCollection
	platformRef     *util.Ref[platform.Platform]
	libraryPaths    []string
	logger          *zerolog.Logger
	database        *db.Database

	fileTree  *FileTree
	filePaths map[string][]string // latest scanned file paths, keyed by library path
}

type NewLibraryExplorerOptions struct {
	PlatformRef *util.Ref[platform.Platform]
	Logger      *zerolog.Logger
	Database    *db.Database
}

func NewLibraryExplorer(opts NewLibraryExplorerOptions) *LibraryExplorer {
	return &LibraryExplorer{
		platformRef: opts.PlatformRef,
		logger:      opts.Logger,
		database:    opts.Database,
	}
}

func (l *LibraryExplorer) SetAnimeCollection(collection *anilist.AnimeCollection) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.animeCollection = collection
}

func (l *LibraryExplorer) SetLibraryPaths(paths []string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.libraryPaths = paths
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Client functions
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// GetFileTree returns the file tree of the library (root level only for lazy loading)
func (l *LibraryExplorer) GetFileTree() (*FileTreeJSON, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	tree, err := l.getFileTree()
	if err != nil {
		return nil, err
	}

	// Hydrate the tree with local file data
	localFileMap, err := l.hydrateLocalFileData(tree)
	if err != nil {
		return nil, err
	}

	return &FileTreeJSON{
		Root:       tree.Root.toJSON(l),
		LocalFiles: localFileMap,
	}, nil
}

// LoadDirectoryChildren is no longer needed since we build the complete tree upfront
// This method is kept for API compatibility but does nothing
func (l *LibraryExplorer) LoadDirectoryChildren(dirPath string) error {
	// Validate that the path is within our library paths for security
	isValidPath := false
	for _, libraryPath := range l.libraryPaths {
		if dirPath == libraryPath || util.IsSubdirectory(libraryPath, dirPath) {
			isValidPath = true
			break
		}
	}

	if !isValidPath {
		return fmt.Errorf("path %s is not within library directories", dirPath)
	}

	// Since we now build the complete tree upfront, this is a no-op
	// The tree is already complete when built
	return nil
}

// findNodeByPath recursively searches for a node with the given path
func (l *LibraryExplorer) findNodeByPath(node *FileTreeNode, targetPath string) *FileTreeNode {
	if node.Path == targetPath {
		return node
	}

	for _, child := range node.Children {
		if found := l.findNodeByPath(child, targetPath); found != nil {
			return found
		}
	}

	return nil
}

func (l *LibraryExplorer) getFileTree() (*FileTree, error) {
	if l.fileTree == nil {
		var err error
		l.fileTree, err = l.buildFileTree()
		if err != nil {
			return nil, err
		}
	}

	return l.fileTree, nil
}

func (l *LibraryExplorer) Refresh() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.fileTree = nil
	l.filePaths = nil

	_, err := l.getFileTree()
	if err != nil {
		return err
	}

	return nil
}
