package extension_repo

import (
	"archive/zip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"seanime/internal/extension"
	"seanime/internal/plugin"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGojaPluginSystemOS tests the $os bindings in the Goja plugin system
func TestGojaPluginSystemOS(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test files and directories
	testFilePath := filepath.Join(tempDir, "test.txt")
	testDirPath := filepath.Join(tempDir, "testdir")
	testContent := []byte("Hello, world!")

	err := os.WriteFile(testFilePath, testContent, 0644)
	require.NoError(t, err)

	err = os.Mkdir(testDirPath, 0755)
	require.NoError(t, err)

	// Test $os.platform and $os.arch
	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $os bindings");
		
		// Test platform and arch
		console.log("Platform:", $os.platform);
		console.log("Arch:", $os.arch);
		
		// Test tempDir
		const tempDirPath = $os.tempDir();
		console.log("Temp dir:", tempDirPath);
		
		// Test readFile
		const content = $os.readFile("${TEST_FILE_PATH}");
		console.log("File content:", $toString(content));
		$store.set("fileContent", $toString(content));
		
		// Test writeFile
		$os.writeFile("${TEST_FILE_PATH}.new", $toBytes("New content"), 0644);
		const newContent = $os.readFile("${TEST_FILE_PATH}.new");
		console.log("New file content:", $toString(newContent));
		$store.set("newFileContent", $toString(newContent));
		
		// Test readDir
		const entries = $os.readDir("${TEST_DIR}");
		console.log("Directory entries:");
		for (const entry of entries) {
			console.log("    Entry:", entry.name());
		}
		$store.set("dirEntries", entries.length);
		
		// Test mkdir
		$os.mkdir("${TEST_DIR}/newdir", 0755);
		const newEntries = $os.readDir("${TEST_DIR}");
		console.log("New directory entries:");
		for (const entry of newEntries) {
			console.log("    Entry:", entry.name());
		}
		$store.set("newDirEntries", newEntries.length);
		
		// Test stat
		const stats = $os.stat("${TEST_FILE_PATH}");
		console.log("File stats:", stats);
		$store.set("fileSize", stats.size());
		
		// Test rename
		$os.rename("${TEST_FILE_PATH}.new", "${TEST_FILE_PATH}.renamed");
		const renamedExists = $os.stat("${TEST_FILE_PATH}.renamed") !== null;
		console.log("Renamed file exists:", renamedExists);
		$store.set("renamedExists", renamedExists);
		
		// Test remove
		$os.remove("${TEST_FILE_PATH}.renamed");
		let removeSuccess = true;
		try {
			$os.stat("${TEST_FILE_PATH}.renamed");
			removeSuccess = false;
		} catch (e) {
			// File should not exist
			removeSuccess = true;
		}
		console.log("Remove success:", removeSuccess);
		$store.set("removeSuccess", removeSuccess);
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEST_FILE_PATH}", testFilePath)
	payload = strings.ReplaceAll(payload, "${TEST_DIR}", tempDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	fileContent, ok := plugin.store.GetOk("fileContent")
	require.True(t, ok, "fileContent should be set in store")
	assert.Equal(t, "Hello, world!", fileContent)

	newFileContent, ok := plugin.store.GetOk("newFileContent")
	require.True(t, ok, "newFileContent should be set in store")
	assert.Equal(t, "New content", newFileContent)

	dirEntries, ok := plugin.store.GetOk("dirEntries")
	require.True(t, ok, "dirEntries should be set in store")
	assert.Equal(t, int64(3), dirEntries) // test.txt, test.txt.new and testdir

	newDirEntries, ok := plugin.store.GetOk("newDirEntries")
	require.True(t, ok, "newDirEntries should be set in store")
	assert.Equal(t, int64(4), newDirEntries) // test.txt, test.txt.new, testdir, and newdir

	fileSize, ok := plugin.store.GetOk("fileSize")
	require.True(t, ok, "fileSize should be set in store")
	assert.Equal(t, int64(13), fileSize) // "Hello, world!" is 13 bytes

	renamedExists, ok := plugin.store.GetOk("renamedExists")
	require.True(t, ok, "renamedExists should be set in store")
	assert.True(t, renamedExists.(bool))

	removeSuccess, ok := plugin.store.GetOk("removeSuccess")
	require.True(t, ok, "removeSuccess should be set in store")
	assert.True(t, removeSuccess.(bool))

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemOSUnauthorized tests that unauthorized paths are rejected
func TestGojaPluginSystemOSUnauthorized(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	unauthorizedDir := filepath.Join(os.TempDir(), "unauthorized")

	// Ensure the unauthorized directory exists
	_ = os.MkdirAll(unauthorizedDir, 0755)
	defer os.RemoveAll(unauthorizedDir)

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing unauthorized $os operations");
		
		// Try to read from unauthorized path
		try {
			const content = $os.readFile("${UNAUTHORIZED_PATH}/test.txt");
			$store.set("unauthorizedRead", false);
		} catch (e) {
			console.log("Unauthorized read error:", e.message);
			$store.set("unauthorizedRead", true);
			$store.set("unauthorizedReadError", e.message);
		}
		
		// Try to write to unauthorized path
		try {
			$os.writeFile("${UNAUTHORIZED_PATH}/test.txt", $toBytes("Unauthorized"), 0644);
			$store.set("unauthorizedWrite", false);
		} catch (e) {
			console.log("Unauthorized write error:", e.message);
			$store.set("unauthorizedWrite", true);
			$store.set("unauthorizedWriteError", e.message);
		}
		
		// Try to read directory from unauthorized path
		try {
			const entries = $os.readDir("${UNAUTHORIZED_PATH}");
			$store.set("unauthorizedReadDir", false);
		} catch (e) {
			console.log("Unauthorized readDir error:", e.message);
			$store.set("unauthorizedReadDir", true);
			$store.set("unauthorizedReadDirError", e.message);
		}
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${UNAUTHORIZED_PATH}", unauthorizedDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/*"},
			WritePaths: []string{tempDir + "/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check that unauthorized operations were rejected
	unauthorizedRead, ok := plugin.store.GetOk("unauthorizedRead")
	require.True(t, ok, "unauthorizedRead should be set in store")
	assert.True(t, unauthorizedRead.(bool))

	unauthorizedReadError, ok := plugin.store.GetOk("unauthorizedReadError")
	require.True(t, ok, "unauthorizedReadError should be set in store")
	assert.Contains(t, unauthorizedReadError.(string), "not authorized for read")

	unauthorizedWrite, ok := plugin.store.GetOk("unauthorizedWrite")
	require.True(t, ok, "unauthorizedWrite should be set in store")
	assert.True(t, unauthorizedWrite.(bool))

	unauthorizedWriteError, ok := plugin.store.GetOk("unauthorizedWriteError")
	require.True(t, ok, "unauthorizedWriteError should be set in store")
	assert.Contains(t, unauthorizedWriteError.(string), "not authorized for write")

	unauthorizedReadDir, ok := plugin.store.GetOk("unauthorizedReadDir")
	require.True(t, ok, "unauthorizedReadDir should be set in store")
	assert.True(t, unauthorizedReadDir.(bool))

	unauthorizedReadDirError, ok := plugin.store.GetOk("unauthorizedReadDirError")
	require.True(t, ok, "unauthorizedReadDirError should be set in store")
	assert.Contains(t, unauthorizedReadDirError.(string), "not authorized for read")

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemOSOpenFile tests the $os.openFile, $os.create functions
func TestGojaPluginSystemOSOpenFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $os.openFile and $os.create");
		
		// Test create
		const file = $os.create("${TEMP_DIR}/created.txt");
		file.writeString("Created file content");
		file.close();
		
		const createdContent = $os.readFile("${TEMP_DIR}/created.txt");
		console.log("Created file content:", $toString(createdContent));
		$store.set("createdContent", $toString(createdContent));
		
		// Test openFile for reading and writing
		const fileRW = $os.openFile("${TEMP_DIR}/rw.txt", $os.O_RDWR | $os.O_CREATE, 0644);
		fileRW.writeString("Read-write file content");
		fileRW.close();
		
		const fileRead = $os.openFile("${TEMP_DIR}/rw.txt", $os.O_RDONLY, 0644);
		const buffer = new Uint8Array(100);
		const bytesRead = fileRead.read(buffer);
		const content = $toString(buffer.subarray(0, bytesRead));
		console.log("Read-write file content:", content);
		$store.set("rwContent", content);
		fileRead.close();
		
		// Test openFile with append
		const fileAppend = $os.openFile("${TEMP_DIR}/rw.txt", $os.O_WRONLY | $os.O_APPEND, 0644);
		fileAppend.writeString(" - Appended content");
		fileAppend.close();
		
		const appendedContent = $os.readFile("${TEMP_DIR}/rw.txt");
		console.log("Appended file content:", $toString(appendedContent));
		$store.set("appendedContent", $toString(appendedContent));
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/*"},
			WritePaths: []string{tempDir + "/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	createdContent, ok := plugin.store.GetOk("createdContent")
	require.True(t, ok, "createdContent should be set in store")
	assert.Equal(t, "Created file content", createdContent)

	rwContent, ok := plugin.store.GetOk("rwContent")
	require.True(t, ok, "rwContent should be set in store")
	assert.Equal(t, "Read-write file content", rwContent)

	appendedContent, ok := plugin.store.GetOk("appendedContent")
	require.True(t, ok, "appendedContent should be set in store")
	assert.Equal(t, "Read-write file content - Appended content", appendedContent)

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemOSMkdirAll tests the $os.mkdirAll function
func TestGojaPluginSystemOSMkdirAll(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $os.mkdirAll");
		
		// Test mkdirAll with nested directories
		$os.mkdirAll("${TEMP_DIR}/nested/dirs/structure", 0755);
		
		// Check if the directories were created
		const nestedExists = $os.stat("${TEMP_DIR}/nested") !== null;
		const dirsExists = $os.stat("${TEMP_DIR}/nested/dirs") !== null;
		const structureExists = $os.stat("${TEMP_DIR}/nested/dirs/structure") !== null;
		
		console.log("Nested directories exist:", nestedExists, dirsExists, structureExists);
		$store.set("nestedExists", nestedExists);
		$store.set("dirsExists", dirsExists);
		$store.set("structureExists", structureExists);
		
		// Create a file in the nested directory
		$os.writeFile("${TEMP_DIR}/nested/dirs/structure/test.txt", $toBytes("Nested file"), 0644);
		const nestedContent = $os.readFile("${TEMP_DIR}/nested/dirs/structure/test.txt");
		console.log("Nested file content:", $toString(nestedContent));
		$store.set("nestedContent", $toString(nestedContent));
		
		// Test removeAll
		$os.removeAll("${TEMP_DIR}/nested");
		let removeAllSuccess = true;
		try {
			$os.stat("${TEMP_DIR}/nested");
			removeAllSuccess = false;
		} catch (e) {
			// Directory should not exist
			removeAllSuccess = true;
		}
		console.log("RemoveAll success:", removeAllSuccess);
		$store.set("removeAllSuccess", removeAllSuccess);
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	nestedExists, ok := plugin.store.GetOk("nestedExists")
	require.True(t, ok, "nestedExists should be set in store")
	assert.True(t, nestedExists.(bool))

	dirsExists, ok := plugin.store.GetOk("dirsExists")
	require.True(t, ok, "dirsExists should be set in store")
	assert.True(t, dirsExists.(bool))

	structureExists, ok := plugin.store.GetOk("structureExists")
	require.True(t, ok, "structureExists should be set in store")
	assert.True(t, structureExists.(bool))

	nestedContent, ok := plugin.store.GetOk("nestedContent")
	require.True(t, ok, "nestedContent should be set in store")
	assert.Equal(t, "Nested file", nestedContent)

	removeAllSuccess, ok := plugin.store.GetOk("removeAllSuccess")
	require.True(t, ok, "removeAllSuccess should be set in store")
	assert.True(t, removeAllSuccess.(bool))

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemOSPermissions tests that the plugin system enforces permissions correctly
func TestGojaPluginSystemOSPermissions(t *testing.T) {
	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $os permissions");
		
		// Try to use $os without system permission
		try {
			const tempDirPath = $os.tempDir();
			$store.set("noPermissionAccess", true);
		} catch (e) {
			console.log("No permission error:", e.message);
			$store.set("noPermissionAccess", false);
			$store.set("noPermissionError", e.message);
		}
	});
}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	// Deliberately NOT including the system permission
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check that operations were rejected due to missing permissions
	noPermissionAccess, ok := plugin.store.GetOk("noPermissionAccess")
	require.True(t, ok, "noPermissionAccess should be set in store")
	assert.False(t, noPermissionAccess.(bool))

	noPermissionError, ok := plugin.store.GetOk("noPermissionError")
	require.True(t, ok, "noPermissionError should be set in store")
	assert.Contains(t, noPermissionError.(string), "$os is not defined")

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemFilepath tests the $filepath bindings in the Goja plugin system
func TestGojaPluginSystemFilepath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test files and directories
	testFilePath := filepath.Join(tempDir, "test.txt")
	nestedDir := filepath.Join(tempDir, "nested", "dir")
	err := os.MkdirAll(nestedDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(testFilePath, []byte("Hello, world!"), 0644)
	require.NoError(t, err)

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $filepath bindings");
		
		// Test base
		const baseName = $filepath.base("${TEST_FILE_PATH}");
		console.log("Base name:", baseName);
		$store.set("baseName", baseName);
		
		// Test dir
		const dirName = $filepath.dir("${TEST_FILE_PATH}");
		console.log("Dir name:", dirName);
		$store.set("dirName", dirName);
		
		// Test ext
		const extName = $filepath.ext("${TEST_FILE_PATH}");
		console.log("Ext name:", extName);
		$store.set("extName", extName);
		
		// Test join
		const joinedPath = $filepath.join("${TEMP_DIR}", "subdir", "file.txt");
		console.log("Joined path:", joinedPath);
		$store.set("joinedPath", joinedPath);
		
		// Test split
		const [dir, file] = $filepath.split("${TEST_FILE_PATH}");
		console.log("Split path:", dir, file);
		$store.set("splitDir", dir);
		$store.set("splitFile", file);
		
		// Test glob
		const globResults = $filepath.glob("${TEMP_DIR}", "*.txt");
		console.log("Glob results:", globResults);
		$store.set("globResults", globResults.length);
		
		// Test match
		const isMatch = $filepath.match("*.txt", "test.txt");
		console.log("Match result:", isMatch);
		$store.set("isMatch", isMatch);
		
		// Test isAbs
		const isAbsPath = $filepath.isAbs("${TEST_FILE_PATH}");
		console.log("Is absolute path:", isAbsPath);
		$store.set("isAbsPath", isAbsPath);
		
		// Test toSlash and fromSlash
		const slashPath = $filepath.toSlash("${TEST_FILE_PATH}");
		console.log("To slash:", slashPath);
		$store.set("slashPath", slashPath);
		
		const fromSlashPath = $filepath.fromSlash(slashPath);
		console.log("From slash:", fromSlashPath);
		$store.set("fromSlashPath", fromSlashPath);
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEST_FILE_PATH}", testFilePath)
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	baseName, ok := plugin.store.GetOk("baseName")
	require.True(t, ok, "baseName should be set in store")
	assert.Equal(t, "test.txt", baseName)

	dirName, ok := plugin.store.GetOk("dirName")
	require.True(t, ok, "dirName should be set in store")
	assert.Equal(t, tempDir, dirName)

	extName, ok := plugin.store.GetOk("extName")
	require.True(t, ok, "extName should be set in store")
	assert.Equal(t, ".txt", extName)

	joinedPath, ok := plugin.store.GetOk("joinedPath")
	require.True(t, ok, "joinedPath should be set in store")
	assert.Equal(t, filepath.Join(tempDir, "subdir", "file.txt"), joinedPath)

	splitDir, ok := plugin.store.GetOk("splitDir")
	require.True(t, ok, "splitDir should be set in store")
	assert.Equal(t, tempDir+string(filepath.Separator), splitDir)

	splitFile, ok := plugin.store.GetOk("splitFile")
	require.True(t, ok, "splitFile should be set in store")
	assert.Equal(t, "test.txt", splitFile)

	globResults, ok := plugin.store.GetOk("globResults")
	require.True(t, ok, "globResults should be set in store")
	assert.Equal(t, int64(1), globResults) // test.txt

	isMatch, ok := plugin.store.GetOk("isMatch")
	require.True(t, ok, "isMatch should be set in store")
	assert.True(t, isMatch.(bool))

	isAbsPath, ok := plugin.store.GetOk("isAbsPath")
	require.True(t, ok, "isAbsPath should be set in store")
	assert.True(t, isAbsPath.(bool))

	slashPath, ok := plugin.store.GetOk("slashPath")
	require.True(t, ok, "slashPath should be set in store")
	assert.Contains(t, slashPath.(string), "/")

	fromSlashPath, ok := plugin.store.GetOk("fromSlashPath")
	require.True(t, ok, "fromSlashPath should be set in store")
	assert.Equal(t, testFilePath, fromSlashPath)

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemIO tests the $io bindings in the Goja plugin system
func TestGojaPluginSystemIO(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test file
	testFilePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFilePath, []byte("Hello, world!"), 0644)
	require.NoError(t, err)

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $io bindings");
		
		// Test readAll
		const file = $os.openFile("${TEST_FILE_PATH}", $os.O_RDONLY, 0);
		const content = $io.readAll(file);
		file.close();
		console.log("Read content:", $toString(content));
		$store.set("readAllContent", $toString(content));
		
		// Test copy
		const srcFile = $os.openFile("${TEST_FILE_PATH}", $os.O_RDONLY, 0);
		const destFile = $os.create("${TEMP_DIR}/copy.txt");
		const bytesCopied = $io.copy(destFile, srcFile);
		srcFile.close();
		destFile.close();
		console.log("Bytes copied:", bytesCopied);
		$store.set("bytesCopied", bytesCopied);
		
		// Test writeString
		const stringFile = $os.create("${TEMP_DIR}/string.txt");
		const bytesWritten = $io.writeString(stringFile, "Written with writeString");
		stringFile.close();
		console.log("Bytes written:", bytesWritten);
		$store.set("bytesWritten", bytesWritten);
		
		// Read the file back to verify
		const stringContent = $os.readFile("${TEMP_DIR}/string.txt");
		console.log("String content:", $toString(stringContent));
		$store.set("stringContent", $toString(stringContent));
		
		// Test copyN
		const srcFileN = $os.openFile("${TEST_FILE_PATH}", $os.O_RDONLY, 0);
		const destFileN = $os.create("${TEMP_DIR}/copyN.txt");
		const bytesCopiedN = $io.copyN(destFileN, srcFileN, 5); // Copy only 5 bytes
		srcFileN.close();
		destFileN.close();
		console.log("Bytes copied with copyN:", bytesCopiedN);
		$store.set("bytesCopiedN", bytesCopiedN);
		
		// Read the file back to verify
		const copyNContent = $os.readFile("${TEMP_DIR}/copyN.txt");
		console.log("CopyN content:", $toString(copyNContent));
		$store.set("copyNContent", $toString(copyNContent));
		
		// Test limitReader
		const bigFile = $os.openFile("${TEST_FILE_PATH}", $os.O_RDONLY, 0);
		const limitedReader = $io.limitReader(bigFile, 5); // Limit to 5 bytes
		const limitBuffer = new Uint8Array(100);
		const limitBytesRead = limitedReader.read(limitBuffer);
		bigFile.close();
		console.log("Limited bytes read:", limitBytesRead);
		$store.set("limitBytesRead", limitBytesRead);
		$store.set("limitContent", $toString(limitBuffer.subarray(0, limitBytesRead)));
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEST_FILE_PATH}", testFilePath)
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	readAllContent, ok := plugin.store.GetOk("readAllContent")
	require.True(t, ok, "readAllContent should be set in store")
	assert.Equal(t, "Hello, world!", readAllContent)

	bytesCopied, ok := plugin.store.GetOk("bytesCopied")
	require.True(t, ok, "bytesCopied should be set in store")
	assert.Equal(t, int64(13), bytesCopied) // "Hello, world!" is 13 bytes

	bytesWritten, ok := plugin.store.GetOk("bytesWritten")
	require.True(t, ok, "bytesWritten should be set in store")
	assert.Equal(t, int64(24), bytesWritten) // "Written with writeString" is 24 bytes

	stringContent, ok := plugin.store.GetOk("stringContent")
	require.True(t, ok, "stringContent should be set in store")
	assert.Equal(t, "Written with writeString", stringContent)

	bytesCopiedN, ok := plugin.store.GetOk("bytesCopiedN")
	require.True(t, ok, "bytesCopiedN should be set in store")
	assert.Equal(t, int64(5), bytesCopiedN)

	copyNContent, ok := plugin.store.GetOk("copyNContent")
	require.True(t, ok, "copyNContent should be set in store")
	assert.Equal(t, "Hello", copyNContent)

	limitBytesRead, ok := plugin.store.GetOk("limitBytesRead")
	require.True(t, ok, "limitBytesRead should be set in store")
	assert.Equal(t, int64(5), limitBytesRead)

	limitContent, ok := plugin.store.GetOk("limitContent")
	require.True(t, ok, "limitContent should be set in store")
	assert.Equal(t, "Hello", limitContent)

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemBufio tests the $bufio bindings in the Goja plugin system
func TestGojaPluginSystemBufio(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test file with multiple lines
	testFilePath := filepath.Join(tempDir, "multiline.txt")
	err := os.WriteFile(testFilePath, []byte("Line 1\nLine 2\nLine 3\nLine 4\n"), 0644)
	require.NoError(t, err)

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $bufio bindings");
		
		// Test NewReader and ReadString
		const file = $os.openFile("${TEST_FILE_PATH}", $os.O_RDONLY, 0);
		const reader = $bufio.newReader(file);
		
		// Read lines one by one with try/catch to handle EOF
		const lines = [];
		for (let i = 0; i < 10; i++) { // Try to read more lines than exist
			try {
				const line = reader.readString($toBytes('\n'));
				console.log("Read line:", line);
				lines.push(line.trim());
			} catch (e) {
				console.log("Caught expected EOF:", e.message);
				$store.set("eofCaught", true);
			}
		}
		file.close();
		
		console.log("Read lines:", lines);
		$store.set("lines", lines);
		
		// Test NewWriter
		const writeFile = $os.create("${TEMP_DIR}/bufio_write.txt");
		const writer = $bufio.newWriter(writeFile);
		
		// Write multiple strings
		writer.writeString("Buffered ");
		writer.writeString("write ");
		writer.writeString("test");
		
		// Flush to ensure data is written
		writer.flush();
		writeFile.close();
		
		// Read back the file to verify
		const writtenContent = $os.readFile("${TEMP_DIR}/bufio_write.txt");
		console.log("Written content:", $toString(writtenContent));
		$store.set("writtenContent", $toString(writtenContent));
		
		// Test Scanner
		const scanFile = $os.openFile("${TEST_FILE_PATH}", $os.O_RDONLY, 0);
		const scanner = $bufio.newScanner(scanFile);
		
		// Scan lines
		const scannedLines = [];
		while (scanner.scan()) {
			scannedLines.push(scanner.text());
		}
		scanFile.close();
		
		console.log("Scanned lines:", scannedLines);
		$store.set("scannedLines", scannedLines);
		
		// Test ReadBytes
		try {
			const bytesFile = $os.openFile("${TEST_FILE_PATH}", $os.O_RDONLY, 0);
			const bytesReader = $bufio.newReader(bytesFile);
			
			const bytesLines = [];
			try {
				for (let i = 0; i < 10; i++) {
					const lineBytes = bytesReader.readBytes('\n'.charCodeAt(0));
					bytesLines.push($toString(lineBytes).trim());
				}
			} catch (e) {
				console.log("Caught expected EOF in readBytes:", e.message);
				$store.set("eofCaughtBytes", true);
			}
			bytesFile.close();
			
			console.log("Read bytes lines:", bytesLines);
			$store.set("bytesLines", bytesLines);
		} catch (e) {
			console.log("Error in ReadBytes test:", e.message);
		}
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEST_FILE_PATH}", testFilePath)
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	lines, ok := plugin.store.GetOk("lines")
	require.True(t, ok, "lines should be set in store")
	assert.Equal(t, []interface{}{"Line 1", "Line 2", "Line 3", "Line 4"}, lines)

	eofCaught, ok := plugin.store.GetOk("eofCaught")
	require.True(t, ok, "eofCaught should be set in store")
	assert.True(t, eofCaught.(bool))

	writtenContent, ok := plugin.store.GetOk("writtenContent")
	require.True(t, ok, "writtenContent should be set in store")
	assert.Equal(t, "Buffered write test", writtenContent)

	scannedLines, ok := plugin.store.GetOk("scannedLines")
	require.True(t, ok, "scannedLines should be set in store")
	assert.Equal(t, []interface{}{"Line 1", "Line 2", "Line 3", "Line 4"}, scannedLines)

	bytesLines, ok := plugin.store.GetOk("bytesLines")
	if ok {
		assert.Equal(t, []interface{}{"Line 1", "Line 2", "Line 3", "Line 4"}, bytesLines)
	}

	eofCaughtBytes, ok := plugin.store.GetOk("eofCaughtBytes")
	if ok {
		assert.True(t, eofCaughtBytes.(bool))
	}

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemBytes tests the $bytes bindings in the Goja plugin system
func TestGojaPluginSystemBytes(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $bytes bindings");
		
		// Test NewBuffer
		const buffer = $bytes.newBuffer($toBytes("Hello"));
		buffer.writeString(", world!");
		
		const bufferContent = $toString(buffer.bytes());
		console.log("Buffer content:", bufferContent);
		$store.set("bufferContent", bufferContent);
		
		// Test NewBufferString
		const strBuffer = $bytes.newBufferString("String buffer");
		strBuffer.writeString(" test");
		
		const strBufferContent = strBuffer.string();
		console.log("String buffer content:", strBufferContent);
		$store.set("strBufferContent", strBufferContent);
		
		// Test NewReader
		const reader = $bytes.newReader($toBytes("Bytes reader test"));
		const readerBuffer = new Uint8Array(100);
		const bytesRead = reader.read(readerBuffer);
		
		const readerContent = $toString(readerBuffer.subarray(0, bytesRead));
		console.log("Reader content:", readerContent);
		$store.set("readerContent", readerContent);
		
		// Test buffer methods
		const testBuffer = $bytes.newBuffer($toBytes(""));
		testBuffer.writeString("Test");
		testBuffer.writeByte(32); // Space
		testBuffer.writeString("methods");
		
		const testBufferContent = testBuffer.string();
		console.log("Test buffer content:", testBufferContent);
		$store.set("testBufferContent", testBufferContent);
		
		// Test read methods
		const readBuffer = $bytes.newBuffer($toBytes("Read test"));
		const readByte = readBuffer.readByte();
		console.log("Read byte:", String.fromCharCode(readByte));
		$store.set("readByte", readByte);
		
		const nextBytes = new Uint8Array(4);
		readBuffer.read(nextBytes);
		console.log("Next bytes:", $toString(nextBytes));
		$store.set("nextBytes", $toString(nextBytes));
	});
}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	bufferContent, ok := plugin.store.GetOk("bufferContent")
	require.True(t, ok, "bufferContent should be set in store")
	assert.Equal(t, "Hello, world!", bufferContent)

	strBufferContent, ok := plugin.store.GetOk("strBufferContent")
	require.True(t, ok, "strBufferContent should be set in store")
	assert.Equal(t, "String buffer test", strBufferContent)

	readerContent, ok := plugin.store.GetOk("readerContent")
	require.True(t, ok, "readerContent should be set in store")
	assert.Equal(t, "Bytes reader test", readerContent)

	testBufferContent, ok := plugin.store.GetOk("testBufferContent")
	require.True(t, ok, "testBufferContent should be set in store")
	assert.Equal(t, "Test methods", testBufferContent)

	readByte, ok := plugin.store.GetOk("readByte")
	require.True(t, ok, "readByte should be set in store")
	assert.Equal(t, int64('R'), readByte)

	nextBytes, ok := plugin.store.GetOk("nextBytes")
	require.True(t, ok, "nextBytes should be set in store")
	assert.Equal(t, "ead ", nextBytes)

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemDownloader tests the ctx.downloader bindings in the Goja plugin system
func TestGojaPluginSystemDownloader(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test HTTP server that serves a large file in chunks to simulate download progress
	const totalSize = 1024 * 1024 // 1MB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set content length for proper progress calculation
		w.Header().Set("Content-Length", fmt.Sprintf("%d", totalSize))

		// Flush headers to client
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Send data in chunks with delays to simulate download progress
		chunkSize := 32 * 1024 // 32KB chunks
		chunk := make([]byte, chunkSize)
		for i := 0; i < len(chunk); i++ {
			chunk[i] = byte(i % 256)
		}

		for sent := 0; sent < totalSize; sent += chunkSize {
			// Sleep to simulate network delay
			time.Sleep(100 * time.Millisecond)

			// Calculate remaining bytes
			remaining := totalSize - sent
			if remaining < chunkSize {
				chunkSize = remaining
			}

			// Write chunk
			w.Write(chunk[:chunkSize])

			// Flush to ensure client receives data immediately
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}))
	defer server.Close()

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing ctx.downloader bindings with large file");
		
		// Test download
		const downloadPath = "${TEMP_DIR}/large_download.bin";
		try {
			const downloadID = ctx.downloader.download("${SERVER_URL}", downloadPath, {
				timeout: 60 // 60 second timeout
			});
			console.log("Download started with ID:", downloadID);
			$store.set("downloadID", downloadID);
			
			// Track progress updates
			const progressUpdates = [];
			
			// Wait for download to complete
			let downloadComplete = ctx.state(false);
			const cancelWatch = ctx.downloader.watch(downloadID, (progress) => {
				// Store progress update
				progressUpdates.push({
					percentage: progress.percentage,
					totalBytes: progress.totalBytes,
					speed: progress.speed,
					status: progress.status
				});
				
				console.log("Download progress:", 
					progress.percentage.toFixed(2), "%, ", 
					"Speed:", (progress.speed / 1024).toFixed(2), "KB/s, ",
					"Downloaded:", (progress.totalBytes / 1024).toFixed(2), "KB"
				, progress);
				
				if (progress.status === "completed") {
					downloadComplete.set(true);
					$store.set("downloadComplete", true);
					$store.set("downloadProgress", progress);
					$store.set("progressUpdates", progressUpdates);
				} else if (progress.status === "error") {
					console.log("Download error:", progress.error);
					$store.set("downloadError", progress.error);
				}
			});
			
			// Wait for download to complete
			ctx.effect(() => {
				if (!downloadComplete.get()) {
				return
				}
				// Cancel watch
				cancelWatch();
				
				// Check downloaded file
				try {
					if (downloadComplete) {
						const stats = $os.stat(downloadPath);
						console.log("Downloaded file size:", stats.size(), "bytes");
						$store.set("downloadedSize", stats.size());
					}
					
					// List downloads
					const downloads = ctx.downloader.listDownloads();
					console.log("Active downloads:", downloads.length);
					$store.set("downloadsCount", downloads.length);
					
					// Get progress
					const progress = ctx.downloader.getProgress(downloadID);
					if (progress) {
						console.log("Final download progress:", progress);
						$store.set("finalProgress", progress);
					} else {
						console.log("Progress not found for ID:", downloadID);
						$store.set("progressNotFound", true);
					}
				} catch (e) {
					console.log("Error in download check:", e.message);
					$store.set("checkError", e.message);
				}
			}, [downloadComplete]);
		} catch (e) {
			console.log("Error starting download:", e.message);
			$store.set("startError", e.message);
		}
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)
	payload = strings.ReplaceAll(payload, "${SERVER_URL}", server.URL)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	p, logger, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute and download to complete
	time.Sleep(12 * time.Second)

	// Check the store values
	downloadID, ok := p.store.GetOk("downloadID")
	require.True(t, ok, "downloadID should be set in store")
	assert.NotEmpty(t, downloadID)

	// Check if download completed or if there was an error
	downloadComplete, ok := p.store.GetOk("downloadComplete")
	if ok && downloadComplete.(bool) {
		// If download completed, check file size
		downloadedSize, ok := p.store.GetOk("downloadedSize")
		require.True(t, ok, "downloadedSize should be set in store")
		assert.Equal(t, int64(totalSize), downloadedSize)

		// Check progress updates
		progressUpdates, ok := p.store.GetOk("progressUpdates")
		require.True(t, ok, "progressUpdates should be set in store")
		updates, ok := progressUpdates.([]interface{})
		require.True(t, ok, "progressUpdates should be a slice")

		// Should have multiple progress updates
		assert.Greater(t, len(updates), 1, "Should have multiple progress updates")

		// Print progress updates for debugging
		logger.Info().Msgf("Received %d progress updates", len(updates))
		for i, update := range updates {
			if i < 5 || i >= len(updates)-5 {
				logger.Info().Interface("update", update).Msgf("Progress update %d", i)
			} else if i == 5 {
				logger.Info().Msg("... more updates ...")
			}
		}

		finalProgress, ok := p.store.GetOk("finalProgress")
		if ok {
			progressMap, ok := finalProgress.(*plugin.DownloadProgress)
			require.Truef(t, ok, "finalProgress should be a map, got %T", finalProgress)
			assert.Equal(t, "completed", progressMap.Status)
			assert.InDelta(t, 100.0, progressMap.Percentage, 0.1)
		}
	} else {
		// If download failed, check error
		downloadError, _ := p.store.GetOk("downloadError")
		t.Logf("Download error: %v", downloadError)
		// Don't fail the test if there was an error, just log it
	}

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemFilepathWalk tests the walk and walkDir functions in the filepath module
func TestGojaPluginSystemFilepathWalk(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a directory structure for testing walk and walkDir
	dirs := []string{
		filepath.Join(tempDir, "dir1"),
		filepath.Join(tempDir, "dir1", "subdir1"),
		filepath.Join(tempDir, "dir1", "subdir2"),
		filepath.Join(tempDir, "dir2"),
		filepath.Join(tempDir, "dir2", "subdir1"),
	}

	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(tempDir, "dir1", "file3.txt"),
		filepath.Join(tempDir, "dir1", "subdir1", "file4.txt"),
		filepath.Join(tempDir, "dir1", "subdir2", "file5.txt"),
		filepath.Join(tempDir, "dir2", "file6.txt"),
		filepath.Join(tempDir, "dir2", "subdir1", "file7.txt"),
	}

	// Create directories
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	// Create files with content
	for i, file := range files {
		content := fmt.Sprintf("Content of file %d", i+1)
		err := os.WriteFile(file, []byte(content), 0644)
		require.NoError(t, err)
	}

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $filepath walk and walkDir");
		
		// Test walk
		const walkPaths = [];
		const walkErrors = [];
		
		$filepath.walk("${TEMP_DIR}", (path, info, err) => {
			if (err) {
				console.log("Walk error:", path, err);
				walkErrors.push({ path, error: err.message });
				return; // Continue walking
			}
			
			console.log("Walk path:", path, "isDir:", info.isDir());
			walkPaths.push({
				path: path,
				isDir: info.isDir(),
				name: info.name()
			});
			return; // Continue walking
		});
		
		console.log("Walk found", walkPaths.length, "paths");
		$store.set("walkPaths", walkPaths);
		$store.set("walkErrors", walkErrors);
		
		// Test walkDir
		const walkDirPaths = [];
		const walkDirErrors = [];
		
		$filepath.walkDir("${TEMP_DIR}", (path, d, err) => {
			if (err) {
				console.log("WalkDir error:", path, err);
				walkDirErrors.push({ path, error: err.message });
				return; // Continue walking
			}
			
			console.log("WalkDir path:", path, "isDir:", d.isDir());
			walkDirPaths.push({
				path: path,
				isDir: d.isDir(),
				name: d.name()
			});
			return; // Continue walking
		});
		
		console.log("WalkDir found", walkDirPaths.length, "paths");
		$store.set("walkDirPaths", walkDirPaths);
		$store.set("walkDirErrors", walkDirErrors);
		
		// Count files and directories found
		const walkFileCount = walkPaths.filter(p => !p.isDir).length;
		const walkDirCount = walkPaths.filter(p => p.isDir).length;
		const walkDirFileCount = walkDirPaths.filter(p => !p.isDir).length;
		const walkDirDirCount = walkDirPaths.filter(p => p.isDir).length;
		
		console.log("Walk found", walkFileCount, "files and", walkDirCount, "directories");
		console.log("WalkDir found", walkDirFileCount, "files and", walkDirDirCount, "directories");
		
		$store.set("walkFileCount", walkFileCount);
		$store.set("walkDirCount", walkDirCount);
		$store.set("walkDirFileCount", walkDirFileCount);
		$store.set("walkDirDirCount", walkDirDirCount);
		
		// Test skipping a directory
		const skipDirWalkPaths = [];
		
		$filepath.walk("${TEMP_DIR}", (path, info, err) => {
			if (err) {
				return; // Continue walking
			}
			
			// Skip dir1 and its subdirectories
			if (info.isDir() && info.name() === "dir1") {
				console.log("Skipping directory:", path);
				return $filepath.skipDir; // Skip this directory
			}
			
			skipDirWalkPaths.push(path);
			return; // Continue walking
		});
		
		console.log("Skip dir walk found", skipDirWalkPaths.length, "paths");
		$store.set("skipDirWalkPaths", skipDirWalkPaths);
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	walkPaths, ok := plugin.store.GetOk("walkPaths")
	require.True(t, ok, "walkPaths should be set in store")
	walkPathsSlice, ok := walkPaths.([]interface{})
	require.True(t, ok, "walkPaths should be a slice")

	// Total number of paths should be dirs + files + root dir
	assert.Equal(t, len(dirs)+len(files)+1, len(walkPathsSlice), "walkPaths should contain all directories and files")

	walkDirPaths, ok := plugin.store.GetOk("walkDirPaths")
	require.True(t, ok, "walkDirPaths should be set in store")
	walkDirPathsSlice, ok := walkDirPaths.([]interface{})
	require.True(t, ok, "walkDirPaths should be a slice")

	// Total number of paths should be dirs + files + root dir
	assert.Equal(t, len(dirs)+len(files)+1, len(walkDirPathsSlice), "walkDirPaths should contain all directories and files")

	// Check file and directory counts
	walkFileCount, ok := plugin.store.GetOk("walkFileCount")
	require.True(t, ok, "walkFileCount should be set in store")
	assert.Equal(t, int64(len(files)), walkFileCount, "walkFileCount should match the number of files")

	walkDirCount, ok := plugin.store.GetOk("walkDirCount")
	require.True(t, ok, "walkDirCount should be set in store")
	assert.Equal(t, int64(len(dirs)+1), walkDirCount, "walkDirCount should match the number of directories plus root")

	walkDirFileCount, ok := plugin.store.GetOk("walkDirFileCount")
	require.True(t, ok, "walkDirFileCount should be set in store")
	assert.Equal(t, int64(len(files)), walkDirFileCount, "walkDirFileCount should match the number of files")

	walkDirDirCount, ok := plugin.store.GetOk("walkDirDirCount")
	require.True(t, ok, "walkDirDirCount should be set in store")
	assert.Equal(t, int64(len(dirs)+1), walkDirDirCount, "walkDirDirCount should match the number of directories plus root")

	// Check skipping directories
	skipDirWalkPaths, ok := plugin.store.GetOk("skipDirWalkPaths")
	require.True(t, ok, "skipDirWalkPaths should be set in store")
	skipDirWalkPathsSlice, ok := skipDirWalkPaths.([]interface{})
	require.True(t, ok, "skipDirWalkPaths should be a slice")

	// Count how many paths should be left after skipping dir1 and its subdirectories
	// We should have tempDir, dir2, dir2/subdir1, and their files (file1.txt, file2.txt, file6.txt, file7.txt)
	expectedPathsAfterSkip := 1 + 2 + 4 // root + dir2 dirs + files in root and dir2
	assert.Equal(t, expectedPathsAfterSkip, len(skipDirWalkPathsSlice), "skipDirWalkPaths should not contain dir1 and its subdirectories")

	// Check for errors
	walkErrors, ok := plugin.store.GetOk("walkErrors")
	require.True(t, ok, "walkErrors should be set in store")
	walkErrorsSlice, ok := walkErrors.([]interface{})
	require.True(t, ok, "walkErrors should be a slice")
	assert.Empty(t, walkErrorsSlice, "There should be no walk errors")

	walkDirErrors, ok := plugin.store.GetOk("walkDirErrors")
	require.True(t, ok, "walkDirErrors should be set in store")
	walkDirErrorsSlice, ok := walkDirErrors.([]interface{})
	require.True(t, ok, "walkDirErrors should be a slice")
	assert.Empty(t, walkDirErrorsSlice, "There should be no walkDir errors")

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemCommands tests the command execution functionality in the system module
func TestGojaPluginSystemCommands(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file to use with commands
	testFilePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFilePath, []byte("Hello, world!"), 0644)
	require.NoError(t, err)

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing command execution");
		
		// Test executing a simple command
		try {
			// Create a command to list files
			const cmd = $os.cmd("ls", "-la", "${TEMP_DIR}");
			
			// Set up stdout capture
			const stdoutPipe = cmd.stdoutPipe();
			
			// Start the command
			cmd.start();
			
			// Read the output
			const output = $io.readAll(stdoutPipe);
			console.log("Command output:", $toString(output));
			$store.set("commandOutput", $toString(output));
			
			// Wait for the command to complete
			cmd.wait();
			
			// Check exit code
			const exitCode = cmd.processState.exitCode();
			console.log("Command exit code:", exitCode);
			$store.set("commandExitCode", exitCode);
		} catch (e) {
			console.log("Command execution error:", e.message);
			$store.set("commandError", e.message);
		}

		// Test executing an async command
		//try {
		//	// Create a command to list files
		//	const asyncCmd = $osExtra.asyncCmd("ls", "-la", "${TEMP_DIR}");
		//	
		//	
		//	asyncCmd.run((data, err, exitCode, signal) => {
		//	// console.log(data, err, exitCode, signal)
		//		if (data) {
		//			console.log("Async command data:", $toString(data));
		//		}
		//		if (err) {
		//			console.log("Async command error:", $toString(err));
		//		}
		//		if (exitCode) {
		//			console.log("Async command exit code:", exitCode);
		//		}
		//		if (signal) {
		//			console.log("Async command signal:", signal);
		//		}
		//	});
		//} catch (e) {
		//	console.log("Command execution error:", e.message);
		//	$store.set("asyncCommandError", e.message);
		//}

		// // Try unsafe goroutine
		// try {
		// 	// Create a command to list files
		// 	const cmd = $os.cmd("ls", "-la", "${TEMP_DIR}");
			
		// 	$store.watch("unsafeGoroutineOutput", (output) => {
		// 		console.log("Unsafe goroutine output:", output);
		// 	});

		// 	// Read the output using scanner
		// 	$unsafeGoroutine(function() {
		// 		// Set up stdout capture
		// 		const stdoutPipe = cmd.stdoutPipe();

		// 		console.log("Starting unsafe goroutine");
		// 		const output = $io.readAll(stdoutPipe);
		// 		$store.set("unsafeGoroutineOutput", $toString(output));
		// 		console.log("Unsafe goroutine output set", $toString(output));

		// 		cmd.wait();
		// 	});
			
		// 	// Start the command
		// 	cmd.start();
			
		// 	// Check exit code
		// 	const exitCode = cmd.processState.exitCode();
		// 	console.log("Command exit code:", exitCode);

		// } catch (e) {
		// 	console.log("Command execution error:", e.message);
		// 	$store.set("unsafeGoroutineError", e.message);
		// }
		
		// Test executing a command with combined output
		try {
			// Create a command to find a string in a file
			const cmd = $os.cmd("grep", "Hello", "${TEST_FILE_PATH}");
			
			// Run the command and capture output
			const output = cmd.combinedOutput();
			console.log("Grep output:", $toString(output));
			$store.set("grepOutput", $toString(output));
			
			// Check if the command found the string
			const foundString = $toString(output).includes("Hello");
			console.log("Found string:", foundString);
			$store.set("foundString", foundString);
		} catch (e) {
			console.log("Grep execution error:", e.message);
			$store.set("grepError", e.message);
		}
		
		// Test executing a command with input
		try {
			// Create a command to sort lines
			const cmd = $os.cmd("sort");
			
			// Set up stdin and stdout pipes
			const stdinPipe = cmd.stdinPipe();
			const stdoutPipe = cmd.stdoutPipe();
			
			// Start the command
			cmd.start();
			
			// Write to stdin
			$io.writeString(stdinPipe, "c\nb\na\n");
			stdinPipe.close();
			
			// Read sorted output
			const sortedOutput = $io.readAll(stdoutPipe);
			console.log("Sorted output:", $toString(sortedOutput));
			$store.set("sortedOutput", $toString(sortedOutput));
			
			// Wait for the command to complete
			cmd.wait();
		} catch (e) {
			console.log("Sort execution error:", e.message);
			$store.set("sortError", e.message);
		}
		
		// Test unauthorized command
		try {
			// Try to execute an unauthorized command
			const cmd = $os.cmd("open", "https://google.com");
			cmd.run();
			$store.set("unauthorizedCommandRan", true);
		} catch (e) {
			console.log("Unauthorized command error:", e.message);
			$store.set("unauthorizedCommandError", e.message);
			$store.set("unauthorizedCommandRan", false);
		}
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)
	payload = strings.ReplaceAll(payload, "${TEST_FILE_PATH}", testFilePath)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
			CommandScopes: []extension.CommandScope{
				{
					Command: "ls",
					Args: []extension.CommandArg{
						{Value: "-la"},
						{Validator: "$PATH"},
					},
				},
				{
					Command: "grep",
					Args: []extension.CommandArg{
						{Value: "Hello"},
						{Validator: "$PATH"},
					},
				},
				{
					Command: "sort",
					Args:    []extension.CommandArg{},
				},
			},
		},
	}

	fmt.Println(opts.Permissions.GetDescription())

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values for the ls command
	commandOutput, ok := plugin.store.GetOk("commandOutput")
	require.True(t, ok, "commandOutput should be set in store")
	assert.Contains(t, commandOutput.(string), "test.txt", "Command output should contain the test file")

	commandExitCode, ok := plugin.store.GetOk("commandExitCode")
	require.True(t, ok, "commandExitCode should be set in store")
	assert.Equal(t, int64(0), commandExitCode, "Command exit code should be 0")

	// Check the store values for the grep command
	grepOutput, ok := plugin.store.GetOk("grepOutput")
	if ok {
		assert.Contains(t, grepOutput.(string), "Hello", "Grep output should contain 'Hello'")
	}

	foundString, ok := plugin.store.GetOk("foundString")
	if ok {
		assert.True(t, foundString.(bool), "Should have found the string in the file")
	}

	// Check the store values for the sort command
	sortedOutput, ok := plugin.store.GetOk("sortedOutput")
	if ok {
		// Expected output: "a\nb\nc\n" (sorted)
		assert.Contains(t, sortedOutput.(string), "a", "Sorted output should contain 'a'")
		assert.Contains(t, sortedOutput.(string), "b", "Sorted output should contain 'b'")
		assert.Contains(t, sortedOutput.(string), "c", "Sorted output should contain 'c'")

		// Check if the lines are in the correct order
		lines := strings.Split(strings.TrimSpace(sortedOutput.(string)), "\n")
		if len(lines) >= 3 {
			assert.Equal(t, "a", lines[0], "First line should be 'a'")
			assert.Equal(t, "b", lines[1], "Second line should be 'b'")
			assert.Equal(t, "c", lines[2], "Third line should be 'c'")
		}
	}

	// Check that unauthorized command was rejected
	unauthorizedCommandRan, ok := plugin.store.GetOk("unauthorizedCommandRan")
	require.True(t, ok, "unauthorizedCommandRan should be set in store")
	assert.False(t, unauthorizedCommandRan.(bool), "Unauthorized command should not have run")

	unauthorizedCommandError, ok := plugin.store.GetOk("unauthorizedCommandError")
	require.True(t, ok, "unauthorizedCommandError should be set in store")
	assert.Contains(t, unauthorizedCommandError.(string), "not authorized", "Error should indicate command was not authorized")

	manager.PrintPluginPoolMetrics(opts.ID)
}

func TestGojaPluginSystemAsyncCommand(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file to use with commands
	testFilePath := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFilePath, []byte("Hello, world!"), 0644)
	require.NoError(t, err)

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing async command execution");

		// Test executing an async command
		try {
			// Create a command to list files
			let asyncCmd = $osExtra.asyncCmd("ls", "-la", "${TEMP_DIR}");
			
			let output = "";
			asyncCmd.run((data, err, exitCode, signal) => {
				// console.log(data, err, exitCode, signal)
				if (data) {
					// console.log("Async command data:", $toString(data));
					output += $toString(data) + "\n";
					$store.set("asyncCommandData", $toString(output));
				}
				if (err) {
					console.log("Async command error:", $toString(err));
					$store.set("asyncCommandError", $toString(err));
				}
				if (exitCode !== undefined) {
					console.log("output 1", output)
					console.log("Async command exit code:", exitCode);
					$store.set("asyncCommandExitCode", exitCode);
					console.log("Async command signal:", signal);
					$store.set("asyncCommandSignal", signal);
				}
			});

			console.log("Running second command")

			let asyncCmd2 = $osExtra.asyncCmd("ls", "-la", "${TEMP_DIR}");
			
			let output2 = "";
			asyncCmd2.run((data, err, exitCode, signal) => {
				// console.log(data, err, exitCode, signal)
				if (data) {
					// console.log("Async command data:", $toString(data));
					output2 += $toString(data) + "\n";
					$store.set("asyncCommandData", $toString(output2));
				}
				if (err) {
					console.log("Async command error:", $toString(err));
					$store.set("asyncCommandError", $toString(err));
				}
				if (exitCode !== undefined) {
					console.log("output 2", output2)
					console.log("Async command exit code:", exitCode);
					$store.set("asyncCommandExitCode", exitCode);
					console.log("Async command signal:", signal);
					$store.set("asyncCommandSignal", signal);
				}
			});

		} catch (e) {
			console.log("Command execution error:", e.message);
			$store.set("asyncCommandError", e.message);
		}
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${TEMP_DIR}", tempDir)
	payload = strings.ReplaceAll(payload, "${TEST_FILE_PATH}", testFilePath)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
			CommandScopes: []extension.CommandScope{
				{
					Command: "ls",
					Args: []extension.CommandArg{
						{Value: "-la"},
						{Validator: "$PATH"},
					},
				},
			},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(2 * time.Second)

	// Check the store values for the ls command
	asyncCommandData, ok := plugin.store.GetOk("asyncCommandData")
	require.True(t, ok, "asyncCommandData should be set in store")
	assert.Contains(t, asyncCommandData.(string), "test.txt", "Command output should contain the test file")

	asyncCommandExitCode, ok := plugin.store.GetOk("asyncCommandExitCode")
	require.True(t, ok, "asyncCommandExitCode should be set in store")
	assert.Equal(t, int64(0), asyncCommandExitCode, "Command exit code should be 0")

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemUnzip tests the unzip functionality in the osExtra module
func TestGojaPluginSystemUnzip(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test zip file
	zipPath := filepath.Join(tempDir, "test.zip")
	extractPath := filepath.Join(tempDir, "extracted")

	// Create a zip file with test content
	err := createTestZipFile(zipPath)
	require.NoError(t, err, "Failed to create test zip file")

	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $osExtra unzip functionality");
		
		try {
			// Create extraction directory
			$os.mkdirAll("${EXTRACT_PATH}", 0755);
			
			// Unzip the file
			$osExtra.unzip("${ZIP_PATH}", "${EXTRACT_PATH}");
			console.log("Unzip successful");
			$store.set("unzipSuccess", true);
			
			// Check if files were extracted
			const entries = $os.readDir("${EXTRACT_PATH}");
			const fileNames = entries.map(entry => entry.name());
			console.log("Extracted files:", fileNames);
			$store.set("extractedFiles", fileNames);
			
			// Read content of extracted file
			const content = $os.readFile("${EXTRACT_PATH}/test.txt");
			console.log("Extracted content:", $toString(content));
			$store.set("extractedContent", $toString(content));
			
			// Try to unzip to an unauthorized location
			try {
				$osExtra.unzip("${ZIP_PATH}", "/tmp/unauthorized");
				$store.set("unauthorizedUnzipSuccess", true);
			} catch (e) {
				console.log("Unauthorized unzip error:", e.message);
				$store.set("unauthorizedUnzipError", e.message);
				$store.set("unauthorizedUnzipSuccess", false);
			}
			
		} catch (e) {
			console.log("Unzip error:", e.message);
			$store.set("unzipError", e.message);
		}
	});
}
	`

	// Replace placeholders with actual paths
	payload = strings.ReplaceAll(payload, "${ZIP_PATH}", zipPath)
	payload = strings.ReplaceAll(payload, "${EXTRACT_PATH}", extractPath)

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
		Allow: extension.PluginAllowlist{
			ReadPaths:  []string{tempDir + "/**/*"},
			WritePaths: []string{tempDir + "/**/*"},
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values
	unzipSuccess, ok := plugin.store.GetOk("unzipSuccess")
	require.True(t, ok, "unzipSuccess should be set in store")
	assert.True(t, unzipSuccess.(bool), "Unzip should have succeeded")

	extractedFiles, ok := plugin.store.GetOk("extractedFiles")
	require.True(t, ok, "extractedFiles should be set in store")
	filesSlice, ok := extractedFiles.([]interface{})
	require.True(t, ok, "extractedFiles should be a slice")
	assert.Contains(t, filesSlice, "test.txt", "Extracted files should include test.txt")

	extractedContent, ok := plugin.store.GetOk("extractedContent")
	require.True(t, ok, "extractedContent should be set in store")
	assert.Equal(t, "Test content for zip file", extractedContent, "Extracted content should match original")

	// Check unauthorized unzip attempt
	unauthorizedUnzipSuccess, ok := plugin.store.GetOk("unauthorizedUnzipSuccess")
	require.True(t, ok, "unauthorizedUnzipSuccess should be set in store")
	assert.False(t, unauthorizedUnzipSuccess.(bool), "Unauthorized unzip should have failed")

	unauthorizedUnzipError, ok := plugin.store.GetOk("unauthorizedUnzipError")
	require.True(t, ok, "unauthorizedUnzipError should be set in store")
	assert.Contains(t, unauthorizedUnzipError.(string), "not authorized", "Error should indicate path not authorized")

	manager.PrintPluginPoolMetrics(opts.ID)
}

// TestGojaPluginSystemMime tests the mime module functionality
func TestGojaPluginSystemMime(t *testing.T) {
	payload := `
function init() {
	$ui.register((ctx) => {
		console.log("Testing $mime functionality");
		
		try {
			// Test parsing content type
			const contentType = "text/html; charset=utf-8";
			const parsed = $mime.parse(contentType);
			console.log("Parsed content type:", parsed);
			$store.set("parsedContentType", parsed);
			
			// Test parsing content type with multiple parameters
			const contentTypeWithParams = "application/json; charset=utf-8; boundary=something";
			const parsedWithParams = $mime.parse(contentTypeWithParams);
			console.log("Parsed content type with params:", parsedWithParams);
			$store.set("parsedContentTypeWithParams", parsedWithParams);
			
			// Test formatting content type
			const formatted = $mime.format("text/plain", { charset: "utf-8", boundary: "boundary" });
			console.log("Formatted content type:", formatted);
			$store.set("formattedContentType", formatted);
			
			// Test parsing invalid content type
			try {
				const invalidContentType = "invalid content type";
				const parsedInvalid = $mime.parse(invalidContentType);
				console.log("Parsed invalid content type:", parsedInvalid);
				$store.set("parsedInvalidContentType", parsedInvalid);
			} catch (e) {
				console.log("Invalid content type error:", e.message);
				$store.set("invalidContentTypeError", e.message);
			}
		} catch (e) {
			console.log("Mime test error:", e.message);
			$store.set("mimeTestError", e.message);
		}
	});
}
	`

	opts := DefaultTestPluginOptions()
	opts.Payload = payload
	opts.Permissions = extension.PluginPermissions{
		Scopes: []extension.PluginPermissionScope{
			extension.PluginPermissionSystem,
		},
	}

	plugin, _, manager, _, _, err := InitTestPlugin(t, opts)
	require.NoError(t, err)

	// Wait for the plugin to execute
	time.Sleep(1 * time.Second)

	// Check the store values for parsed content type
	parsedContentType, ok := plugin.store.GetOk("parsedContentType")
	require.True(t, ok, "parsedContentType should be set in store")
	parsedMap, ok := parsedContentType.(map[string]interface{})
	require.True(t, ok, "parsedContentType should be a map")

	// Check media type
	mediaType, ok := parsedMap["mediaType"]
	require.True(t, ok, "mediaType should be in parsed result")
	assert.Equal(t, "text/html", mediaType, "Media type should be text/html")

	// Check parameters
	parameters, ok := parsedMap["parameters"]
	require.True(t, ok, "parameters should be in parsed result")
	paramsMap, ok := parameters.(map[string]string)
	require.Truef(t, ok, "parameters should be a map but got %T", parameters)
	assert.Equal(t, "utf-8", paramsMap["charset"], "charset parameter should be utf-8")

	// Check parsed content type with multiple parameters
	parsedWithParams, ok := plugin.store.GetOk("parsedContentTypeWithParams")
	require.True(t, ok, "parsedContentTypeWithParams should be set in store")
	parsedWithParamsMap, ok := parsedWithParams.(map[string]interface{})
	require.Truef(t, ok, "parsedContentTypeWithParams should be a map but got %T", parsedWithParams)

	// Check media type
	mediaTypeWithParams, ok := parsedWithParamsMap["mediaType"]
	require.True(t, ok, "mediaType should be in parsed result")
	assert.Equal(t, "application/json", mediaTypeWithParams, "Media type should be application/json")

	// Check parameters
	parametersWithParams, ok := parsedWithParamsMap["parameters"]
	require.True(t, ok, "parameters should be in parsed result")
	require.Truef(t, ok, "parameters should be a map but got %T", parametersWithParams)

	// Check formatted content type
	formattedContentType, ok := plugin.store.GetOk("formattedContentType")
	require.True(t, ok, "formattedContentType should be set in store")
	assert.Contains(t, formattedContentType.(string), "text/plain", "Formatted content type should contain text/plain")
	assert.Contains(t, formattedContentType.(string), "charset=utf-8", "Formatted content type should contain charset=utf-8")
	assert.Contains(t, formattedContentType.(string), "boundary=boundary", "Formatted content type should contain boundary=boundary")

	// Check invalid content type error
	invalidContentTypeError, ok := plugin.store.GetOk("invalidContentTypeError")
	if ok {
		assert.NotEmpty(t, invalidContentTypeError, "Invalid content type should have produced an error")
	}

	manager.PrintPluginPoolMetrics(opts.ID)
}

// Helper function to create a test zip file
func createTestZipFile(zipPath string) error {
	// Create a buffer to write our zip to
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	// Create a new zip archive
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add a file to the archive
	fileWriter, err := zipWriter.Create("test.txt")
	if err != nil {
		return err
	}

	// Write content to the file
	_, err = fileWriter.Write([]byte("Test content for zip file"))
	if err != nil {
		return err
	}

	// Add a directory to the archive
	_, err = zipWriter.Create("testdir/")
	if err != nil {
		return err
	}

	// Add a file in the directory
	dirFileWriter, err := zipWriter.Create("testdir/nested.txt")
	if err != nil {
		return err
	}

	// Write content to the nested file
	_, err = dirFileWriter.Write([]byte("Nested file content"))
	if err != nil {
		return err
	}

	return nil
}
