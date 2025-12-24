package library_explorer

import (
	"fmt"
	"seanime/internal/api/anilist"
	"seanime/internal/database/db"
	"seanime/internal/extension"
	"seanime/internal/platforms/anilist_platform"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func TestLibraryExplorer_LogFileTreeStructure(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t)

	logger := util.NewLogger()

	database, err := db.NewDatabase(test_utils.ConfigData.Path.DataDir, test_utils.ConfigData.Database.Name, logger)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	anilistClient := anilist.TestGetMockAnilistClient()
	extensionBankRef := util.NewRef(extension.NewUnifiedBank())
	anilistPlatform := anilist_platform.NewAnilistPlatform(util.NewRef(anilistClient), extensionBankRef, logger, database)

	explorer := NewLibraryExplorer(NewLibraryExplorerOptions{
		PlatformRef: util.NewRef(anilistPlatform),
		Logger:      logger,
		Database:    database,
	})

	settings, err := database.GetSettings()
	if err != nil {
		t.Fatalf("Failed to get settings: %v", err)
	}

	libraryPaths := settings.GetLibrary().GetLibraryPaths()
	explorer.SetLibraryPaths(libraryPaths)

	t.Logf("Using library paths: %v", libraryPaths)

	explorer.LoadDirectoryChildren("/Users/rahim/Documents/collection")
	explorer.LoadDirectoryChildren("/Users/rahim/Documents/collection/Sousou no Frieren")

	// Build file tree
	fileTreeJSON, err := explorer.GetFileTree()
	if err != nil {
		t.Fatalf("Failed to build file tree: %v", err)
	}

	logFileTreeStructure(t, fileTreeJSON.Root, 0)

	t.Logf("Total directories: %d", countNodesByKind(fileTreeJSON.Root, DirectoryNode))
	t.Logf("Total files: %d", countNodesByKind(fileTreeJSON.Root, FileNode))
	t.Logf("Files with LocalFile associations: %d", countFilesWithLocalFile(fileTreeJSON.Root))
	t.Logf("Unique media IDs: %v", collectAllMediaIds(fileTreeJSON.Root))
}

func logFileTreeStructure(t *testing.T, node *FileTreeNodeJSON, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	nodeType := "📁"
	if node.Kind == FileNode {
		nodeType = "📄"
	}

	mediaInfo := ""
	if len(node.MediaIds) > 0 {
		mediaInfo = fmt.Sprintf(" [MediaIds: %v]", node.MediaIds)
	}

	localFileInfo := ""
	if node.LocalFile != nil {
		localFileInfo = fmt.Sprintf(" [LocalFile: MediaId=%d]", node.LocalFile.MediaId)
	}

	t.Logf("%s%s %s%s%s", indent, nodeType, node.Name, mediaInfo, localFileInfo)

	for _, child := range node.Children {
		logFileTreeStructure(t, child, depth+1)
	}
}

func countNodesByKind(node *FileTreeNodeJSON, kind NodeKind) int {
	count := 0
	if node.Kind == kind {
		count++
	}

	for _, child := range node.Children {
		count += countNodesByKind(child, kind)
	}

	return count
}

func countFilesWithLocalFile(node *FileTreeNodeJSON) int {
	count := 0
	if node.Kind == FileNode && node.LocalFile != nil {
		count++
	}

	for _, child := range node.Children {
		count += countFilesWithLocalFile(child)
	}

	return count
}

func collectAllMediaIds(node *FileTreeNodeJSON) []int {
	mediaIdSet := make(map[int]struct{})

	var collectIds func(*FileTreeNodeJSON)
	collectIds = func(n *FileTreeNodeJSON) {
		for _, id := range n.MediaIds {
			mediaIdSet[id] = struct{}{}
		}
		for _, child := range n.Children {
			collectIds(child)
		}
	}

	collectIds(node)

	var mediaIds []int
	for id := range mediaIdSet {
		mediaIds = append(mediaIds, id)
	}

	return mediaIds
}
