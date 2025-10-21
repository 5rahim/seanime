package library_explorer

import (
	"path/filepath"
	"seanime/internal/database/db_bridge"
	"seanime/internal/library/anime"
	"seanime/internal/library/filesystem"
	"seanime/internal/util"
	"sort"
	"strings"
	"sync"

	"github.com/samber/mo"
)

type NodeKind string

const (
	DirectoryNode NodeKind = "directory"
	FileNode      NodeKind = "file"
)

type (
	FileTreeNode struct {
		Name string
		Path string
		Kind NodeKind
		// used for comparison
		// windows -> lowercase with forward slashes, e.g. c:/foo/bar
		// unix -> same case with forward slashes, e.g. /foo/Bar
		NormalizedPath string
		LocalFile      mo.Option[*anime.LocalFile]
		MediaIds       []int
		LocalFiles     []*anime.LocalFile // For directory nodes
		Children       []*FileTreeNode
		cachedSize     int64 // 0 by default, has to be requested by the client
		parent         *FileTreeNode
	}

	FileTreeNodeJSON struct {
		Name           string              `json:"name"`
		Path           string              `json:"path"`
		NormalizedPath string              `json:"normalizedPath"`
		Kind           NodeKind            `json:"kind"`
		Children       []*FileTreeNodeJSON `json:"children"`
		Size           int64               `json:"size,omitempty"`
		// Only present if the node is a file
		LocalFile *anime.LocalFile `json:"localFile,omitempty"`
		// Media Ids of the files in this directory
		MediaIds              []int `json:"mediaIds,omitempty"`
		LocalFileCount        int   `json:"localFileCount,omitempty"`
		MatchedLocalFileCount int   `json:"matchedLocalFileCount,omitempty"`
	}

	FileTree struct {
		Root *FileTreeNode
	}

	FileTreeJSON struct {
		Root       *FileTreeNodeJSON           `json:"root"`
		LocalFiles map[string]*anime.LocalFile `json:"localFiles"`
	}
)

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (l *LibraryExplorer) buildFileTree() (*FileTree, error) {
	ret := &FileTree{}

	l.logger.Debug().Msg("library explorer: Building complete file tree from media file paths")

	// The root node
	ret.Root = &FileTreeNode{
		Name:           "root",
		Path:           ".",
		Kind:           DirectoryNode,
		NormalizedPath: util.NormalizePath("."),
		Children:       make([]*FileTreeNode, 0),
		MediaIds:       make([]int, 0),
	}

	// Get all media file paths from all library directories
	allMediaFiles := make([]string, 0)
	var wg sync.WaitGroup
	var mu sync.Mutex

	wg.Add(len(l.libraryPaths))

	for _, libraryPath := range l.libraryPaths {
		go func(libPath string) {
			defer wg.Done()

			filePaths, err := filesystem.GetMediaFilePathsFromDirS(libPath)
			if err != nil {
				l.logger.Err(err).Str("path", libPath).Msg("library explorer: Could not get file paths from library path")
				return
			}

			mu.Lock()
			allMediaFiles = append(allMediaFiles, filePaths...)
			mu.Unlock()
		}(libraryPath)
	}

	wg.Wait()

	l.logger.Debug().Int("count", len(allMediaFiles)).Msg("library explorer: Found media files")

	// Build the tree from all file paths
	l.buildTreeFromFilePaths(ret.Root, allMediaFiles)

	// Sort all children recursively
	l.sortTreeChildren(ret.Root)

	return ret, nil
}

// buildTreeFromFilePaths builds the complete tree structure from a list of file paths
func (l *LibraryExplorer) buildTreeFromFilePaths(rootNode *FileTreeNode, filePaths []string) {
	// Create a map to track created nodes by their normalized path
	nodeMap := make(map[string]*FileTreeNode)
	nodeMap[rootNode.NormalizedPath] = rootNode

	// First, create library directory nodes as direct children of root
	libraryNodeMap := make(map[string]*FileTreeNode)
	for _, libraryPath := range l.libraryPaths {
		libraryNode := &FileTreeNode{
			Name:           filepath.Base(libraryPath),
			Path:           libraryPath,
			Kind:           DirectoryNode,
			NormalizedPath: util.NormalizePath(libraryPath),
			Children:       make([]*FileTreeNode, 0),
			MediaIds:       make([]int, 0),
			LocalFile:      mo.None[*anime.LocalFile](),
			parent:         rootNode,
		}
		rootNode.Children = append(rootNode.Children, libraryNode)
		nodeMap[libraryNode.NormalizedPath] = libraryNode
		libraryNodeMap[libraryPath] = libraryNode
	}

	// Group files by their parent directory for efficient processing
	dirToFiles := make(map[string][]string)
	allDirs := make(map[string]bool)

	for _, filePath := range filePaths {
		// Get the library path this file belongs to
		var libraryPath string
		for _, libPath := range l.libraryPaths {
			if strings.HasPrefix(util.NormalizePath(filePath), util.NormalizePath(libPath)) {
				libraryPath = libPath
				break
			}
		}

		if libraryPath == "" {
			continue // Skip files not in any library path
		}

		// Get all parent directories relative to the library path
		dirs := l.getParentDirectoriesRelativeToLibrary(filePath, libraryPath)
		for _, dir := range dirs {
			allDirs[dir] = true
		}

		// Add file to its parent directory
		parentDir := filepath.Dir(filePath)
		dirToFiles[parentDir] = append(dirToFiles[parentDir], filePath)
	}

	// Create directory nodes (excluding library directories which are already created)
	for dirPath := range allDirs {
		if !l.isLibraryPath(dirPath) {
			l.ensureDirectoryNodeRelativeToLibrary(dirPath, nodeMap, libraryNodeMap)
		}
	}

	// Create file nodes
	for parentDir, files := range dirToFiles {
		parentNode := nodeMap[util.NormalizePath(parentDir)]
		if parentNode == nil {
			continue
		}

		for _, filePath := range files {
			fileNode := &FileTreeNode{
				Name:           filepath.Base(filePath),
				Path:           filePath,
				Kind:           FileNode,
				NormalizedPath: util.NormalizePath(filePath),
				Children:       make([]*FileTreeNode, 0),
				MediaIds:       make([]int, 0),
				LocalFile:      mo.None[*anime.LocalFile](), // Will be hydrated separately
				parent:         parentNode,
			}

			parentNode.Children = append(parentNode.Children, fileNode)
			nodeMap[fileNode.NormalizedPath] = fileNode
		}
	}
}

// getParentDirectoriesRelativeToLibrary returns all parent directories for a given file path relative to its library path
func (l *LibraryExplorer) getParentDirectoriesRelativeToLibrary(filePath string, libraryPath string) []string {
	var dirs []string
	currentPath := filepath.Dir(filePath)

	// Collect all directory paths from the file's parent up to (but not including) the library path
	for currentPath != libraryPath && currentPath != filepath.Dir(currentPath) {
		dirs = append(dirs, currentPath)
		currentPath = filepath.Dir(currentPath)
	}

	return dirs
}

// isLibraryPath checks if the given path is one of the configured library paths
func (l *LibraryExplorer) isLibraryPath(dirPath string) bool {
	for _, libPath := range l.libraryPaths {
		if util.NormalizePath(dirPath) == util.NormalizePath(libPath) {
			return true
		}
	}
	return false
}

// ensureDirectoryNodeRelativeToLibrary creates directory nodes and their parent hierarchy relative to library paths
func (l *LibraryExplorer) ensureDirectoryNodeRelativeToLibrary(dirPath string, nodeMap map[string]*FileTreeNode, libraryNodeMap map[string]*FileTreeNode) *FileTreeNode {
	normalizedPath := util.NormalizePath(dirPath)

	// Return existing node if it exists
	if node, exists := nodeMap[normalizedPath]; exists {
		return node
	}

	// Find which library this directory belongs to
	var libraryPath string
	var libraryNode *FileTreeNode
	for libPath, libNode := range libraryNodeMap {
		if strings.HasPrefix(normalizedPath, util.NormalizePath(libPath)) {
			libraryPath = libPath
			libraryNode = libNode
			break
		}
	}

	if libraryNode == nil {
		return nil // Directory doesn't belong to any library
	}

	// Find parent directory
	parentDir := filepath.Dir(dirPath)
	var parentNode *FileTreeNode

	if util.NormalizePath(parentDir) == util.NormalizePath(libraryPath) {
		parentNode = libraryNode
	} else {
		parentNode = l.ensureDirectoryNodeRelativeToLibrary(parentDir, nodeMap, libraryNodeMap)
	}

	if parentNode == nil {
		return nil
	}

	// Create the directory node
	dirNode := &FileTreeNode{
		Name:           filepath.Base(dirPath),
		Path:           dirPath,
		Kind:           DirectoryNode,
		NormalizedPath: normalizedPath,
		Children:       make([]*FileTreeNode, 0),
		MediaIds:       make([]int, 0),
		LocalFile:      mo.None[*anime.LocalFile](),
		parent:         parentNode,
	}

	// Add to parent
	parentNode.Children = append(parentNode.Children, dirNode)
	nodeMap[normalizedPath] = dirNode

	return dirNode
}

// sortTreeChildren recursively sorts all children in the tree
func (l *LibraryExplorer) sortTreeChildren(node *FileTreeNode) {
	// Sort current node's children
	sort.Slice(node.Children, func(i, j int) bool {
		// Directories first, then files
		if node.Children[i].Kind != node.Children[j].Kind {
			return node.Children[i].Kind == DirectoryNode
		}
		return strings.ToLower(node.Children[i].Name) < strings.ToLower(node.Children[j].Name)
	})

	// Recursively sort children
	for _, child := range node.Children {
		l.sortTreeChildren(child)
	}
}

// hydrateLocalFileData hydrates the file tree with LocalFile data and MediaIds
func (l *LibraryExplorer) hydrateLocalFileData(tree *FileTree) (map[string]*anime.LocalFile, error) {
	l.logger.Debug().Msg("library explorer: Hydrating file tree with local file data")

	// Get all local files
	localFiles, _, err := db_bridge.GetLocalFiles(l.database)
	if err != nil {
		l.logger.Warn().Err(err).Msg("library explorer: Failed to get local files, skipping hydration")
		return nil, nil // Don't fail the entire operation
	}

	// Create a map for quick LocalFile lookup by normalized path
	localFileMap := make(map[string]*anime.LocalFile)
	for _, lf := range localFiles {
		normalizedPath := util.NormalizePath(lf.Path)
		localFileMap[normalizedPath] = lf
	}

	// Recursively hydrate the tree
	l.hydrateNode(tree.Root, localFileMap)

	return localFileMap, nil
}

// hydrateNode recursively hydrates a node and its children with local file data
func (l *LibraryExplorer) hydrateNode(node *FileTreeNode, localFileMap map[string]*anime.LocalFile) {
	// Clear existing media IDs
	node.MediaIds = make([]int, 0)
	mediaIdSet := make(map[int]struct{})
	localFileSet := make(map[string]*anime.LocalFile)

	if node.Kind == FileNode {
		// For file nodes, try to find associated LocalFile
		if localFile, exists := localFileMap[node.NormalizedPath]; exists {
			node.LocalFile = mo.Some(localFile)
			if localFile.MediaId > 0 {
				node.MediaIds = []int{localFile.MediaId}
				mediaIdSet[localFile.MediaId] = struct{}{}
			}

		} else {
			node.LocalFile = mo.None[*anime.LocalFile]()
		}
	} else {
		// For directory nodes, collect media IDs from children if they are loaded
		for _, child := range node.Children {
			l.hydrateNode(child, localFileMap)
			// Collect media IDs from children
			for _, mediaId := range child.MediaIds {
				mediaIdSet[mediaId] = struct{}{}
			}
			// Collect local files from children
			for _, localFile := range child.LocalFiles {
				localFileSet[localFile.GetNormalizedPath()] = localFile
			}
		}

		// Additionally, collect media IDs from local files that are under this directory
		// even if children haven't been loaded yet
		l.hydrateDirectoryMediaIds(node, localFileMap, mediaIdSet, localFileSet)

		// Convert set to slice and sort
		node.MediaIds = make([]int, 0, len(mediaIdSet))
		for mediaId := range mediaIdSet {
			node.MediaIds = append(node.MediaIds, mediaId)
		}
		sort.Ints(node.MediaIds)

		// Collect local files
		node.LocalFiles = make([]*anime.LocalFile, 0, len(localFileSet))
		for _, localFile := range localFileSet {
			node.LocalFiles = append(node.LocalFiles, localFile)
		}
	}
}

// hydrateDirectoryMediaIds collects MediaIds from local files under a directory path
func (l *LibraryExplorer) hydrateDirectoryMediaIds(dirNode *FileTreeNode, localFileMap map[string]*anime.LocalFile, mediaIdSet map[int]struct{}, localFileSet map[string]*anime.LocalFile) {
	normalizedDirPath := dirNode.NormalizedPath

	// Ensure directory path ends with a separator for proper matching
	if !strings.HasSuffix(normalizedDirPath, "/") {
		normalizedDirPath += "/"
	}

	// Iterate through all local files to find ones under this directory
	for localFilePath, localFile := range localFileMap {
		// Check if this local file is under the current directory
		if strings.HasPrefix(localFilePath, normalizedDirPath) {
			if !localFile.Ignored && localFile.MediaId > 0 {
				mediaIdSet[localFile.MediaId] = struct{}{}
			}
			if !localFile.Ignored {
				localFileSet[localFilePath] = localFile
			}
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (n *FileTreeNode) toJSON(explorer *LibraryExplorer) *FileTreeNodeJSON {
	ret := &FileTreeNodeJSON{
		Name:           n.Name,
		Path:           n.Path,
		NormalizedPath: n.NormalizedPath,
		Kind:           n.Kind,
		Children:       make([]*FileTreeNodeJSON, len(n.Children)),
		Size:           n.cachedSize,
		MediaIds:       n.MediaIds,
	}

	if n.Kind == DirectoryNode {
		ret.LocalFileCount = len(n.LocalFiles)
		for _, localFile := range n.LocalFiles {
			if localFile.MediaId > 0 {
				ret.MatchedLocalFileCount++
			}
		}
	}

	for i, child := range n.Children {
		ret.Children[i] = child.toJSON(explorer)
	}

	if lf, ok := n.LocalFile.Get(); ok {
		ret.LocalFile = lf
		ret.LocalFileCount = 1
		ret.MatchedLocalFileCount = 1
	}

	return ret
}
