package plugin

import (
	"encoding/json"
	"errors"
	"seanime/internal/database/models"
	"seanime/internal/extension"
	"seanime/internal/util/result"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// Storage is used to store data for an extension.
// A new instance is created for each extension.
type Storage struct {
	ctx             *AppContextImpl
	ext             *extension.Extension
	logger          *zerolog.Logger
	pluginDataCache *result.Map[string, *models.PluginData] // Cache to avoid repeated database calls
	keyDataCache    *result.Map[string, interface{}]        // Cache to avoid repeated database calls
}

var (
	ErrDatabaseNotInitialized = errors.New("database is not initialized")
)

// BindStorage binds the storage API to the Goja runtime.
// Permissions need to be checked by the caller.
// Permissions needed: storage
func (a *AppContextImpl) BindStorage(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension) {
	storageLogger := logger.With().Str("id", ext.ID).Logger()
	storage := &Storage{
		ctx:             a,
		ext:             ext,
		logger:          &storageLogger,
		pluginDataCache: result.NewResultMap[string, *models.PluginData](),
		keyDataCache:    result.NewResultMap[string, interface{}](),
	}
	storageObj := vm.NewObject()
	_ = storageObj.Set("get", storage.Get)
	_ = storageObj.Set("set", storage.Set)
	_ = storageObj.Set("remove", storage.Delete)
	_ = storageObj.Set("drop", storage.Drop)
	_ = storageObj.Set("clear", storage.Clear)
	_ = storageObj.Set("keys", storage.Keys)
	_ = storageObj.Set("has", storage.Has)
	_ = vm.Set("$storage", storageObj)
}

// getDB returns the database instance or an error if not initialized
func (s *Storage) getDB() (*gorm.DB, error) {
	db, ok := s.ctx.database.Get()
	if !ok {
		return nil, ErrDatabaseNotInitialized
	}
	return db.Gorm(), nil
}

// getPluginData retrieves the plugin data from the database
// If createIfNotExists is true, it will create an empty record if none exists
func (s *Storage) getPluginData(createIfNotExists bool) (*models.PluginData, error) {
	// Check cache first
	if cachedData, ok := s.pluginDataCache.Get(s.ext.ID); ok {
		return cachedData, nil
	}

	db, err := s.getDB()
	if err != nil {
		return nil, err
	}

	pluginData := &models.PluginData{}
	if err := db.Where("plugin_id = ?", s.ext.ID).Find(pluginData).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && createIfNotExists {
			// Create empty data structure
			baseData := make(map[string]interface{})
			baseDataMarshaled, err := json.Marshal(baseData)
			if err != nil {
				return nil, err
			}

			newPluginData := &models.PluginData{
				PluginID: s.ext.ID,
				Data:     baseDataMarshaled,
			}

			if err := db.Create(newPluginData).Error; err != nil {
				return nil, err
			}

			// Cache the new plugin data
			s.pluginDataCache.Set(s.ext.ID, newPluginData)
			return newPluginData, nil
		}
		return nil, err
	}

	// Cache the plugin data
	s.pluginDataCache.Set(s.ext.ID, pluginData)
	return pluginData, nil
}

// getDataMap unmarshals the plugin data into a map
func (s *Storage) getDataMap(pluginData *models.PluginData) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(pluginData.Data, &data); err != nil {
		return make(map[string]interface{}), err
	}
	return data, nil
}

// saveDataMap marshals and saves the data map to the database
func (s *Storage) saveDataMap(pluginData *models.PluginData, data map[string]interface{}) error {
	marshaled, err := json.Marshal(data)
	if err != nil {
		return err
	}

	pluginData.Data = marshaled

	db, err := s.getDB()
	if err != nil {
		return err
	}

	err = db.Save(pluginData).Error
	if err != nil {
		return err
	}

	// Update the cache
	s.pluginDataCache.Set(s.ext.ID, pluginData)

	// Don't clear the key data cache here as it would invalidate
	// recently set values. Individual operations (Delete, Clear, Drop)
	// will handle their own cache invalidation as needed.

	return nil
}

// getNestedValue retrieves a value from a nested map using dot notation
func getNestedValue(data map[string]interface{}, path string) interface{} {
	if !strings.Contains(path, ".") {
		return data[path]
	}

	parts := strings.Split(path, ".")
	current := data

	// Navigate through all parts except the last one
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			return nil
		}

		// Try to convert to map for next level
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			// Try to convert from unmarshaled JSON
			jsonMap, ok := next.(map[string]interface{})
			if !ok {
				return nil
			}
			nextMap = jsonMap
		}

		current = nextMap
	}

	// Return the value at the final part
	return current[parts[len(parts)-1]]
}

// setNestedValue sets a value in a nested map using dot notation
// It creates intermediate maps as needed
func setNestedValue(data map[string]interface{}, path string, value interface{}) {
	if !strings.Contains(path, ".") {
		data[path] = value
		return
	}

	parts := strings.Split(path, ".")
	current := data

	// Navigate and create intermediate maps as needed
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			// Create new map if key doesn't exist
			next = make(map[string]interface{})
			current[part] = next
		}

		// Try to convert to map for next level
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			// Try to convert from unmarshaled JSON
			jsonMap, ok := next.(map[string]interface{})
			if !ok {
				// Replace with a new map if not convertible
				nextMap = make(map[string]interface{})
				current[part] = nextMap
			} else {
				nextMap = jsonMap
				current[part] = nextMap
			}
		}

		current = nextMap
	}

	// Set the value at the final part
	current[parts[len(parts)-1]] = value
}

// deleteNestedValue deletes a value from a nested map using dot notation
// Returns true if the key was found and deleted
func deleteNestedValue(data map[string]interface{}, path string) bool {
	if !strings.Contains(path, ".") {
		_, exists := data[path]
		if exists {
			delete(data, path)
			return true
		}
		return false
	}

	parts := strings.Split(path, ".")
	current := data

	// Navigate through all parts except the last one
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			return false
		}

		// Try to convert to map for next level
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			// Try to convert from unmarshaled JSON
			jsonMap, ok := next.(map[string]interface{})
			if !ok {
				return false
			}
			nextMap = jsonMap
		}

		current = nextMap
	}

	// Delete the value at the final part
	lastPart := parts[len(parts)-1]
	_, exists := current[lastPart]
	if exists {
		delete(current, lastPart)
		return true
	}
	return false
}

// hasNestedKey checks if a nested key exists using dot notation
func hasNestedKey(data map[string]interface{}, path string) bool {
	if !strings.Contains(path, ".") {
		_, exists := data[path]
		return exists
	}

	parts := strings.Split(path, ".")
	current := data

	// Navigate through all parts except the last one
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		next, ok := current[part]
		if !ok {
			return false
		}

		// Try to convert to map for next level
		nextMap, ok := next.(map[string]interface{})
		if !ok {
			// Try to convert from unmarshaled JSON
			jsonMap, ok := next.(map[string]interface{})
			if !ok {
				return false
			}
			nextMap = jsonMap
		}

		current = nextMap
	}

	// Check if the final key exists
	_, exists := current[parts[len(parts)-1]]
	return exists
}

// getAllKeys recursively gets all keys from a nested map using dot notation
func getAllKeys(data map[string]interface{}, prefix string) []string {
	keys := make([]string, 0)

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		keys = append(keys, fullKey)

		// If value is a map, recursively get its keys
		if nestedMap, ok := value.(map[string]interface{}); ok {
			nestedKeys := getAllKeys(nestedMap, fullKey)
			keys = append(keys, nestedKeys...)
		}
	}

	return keys
}

func (s *Storage) Delete(key string) error {
	s.logger.Trace().Msgf("plugin: Deleting key %s", key)

	// Remove from key cache
	s.keyDataCache.Delete(key)

	pluginData, err := s.getPluginData(false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	data, err := s.getDataMap(pluginData)
	if err != nil {
		return err
	}

	if deleteNestedValue(data, key) {
		return s.saveDataMap(pluginData, data)
	}

	return nil
}

func (s *Storage) Drop() error {
	s.logger.Trace().Msg("plugin: Dropping storage")

	// Clear caches
	s.pluginDataCache.Clear()
	s.keyDataCache.Clear()

	db, err := s.getDB()
	if err != nil {
		return err
	}

	return db.Where("plugin_id = ?", s.ext.ID).Delete(&models.PluginData{}).Error
}

func (s *Storage) Clear() error {
	s.logger.Trace().Msg("plugin: Clearing storage")

	// Clear key cache
	s.keyDataCache.Clear()

	pluginData, err := s.getPluginData(true)
	if err != nil {
		return err
	}

	cleanData := make(map[string]interface{})
	return s.saveDataMap(pluginData, cleanData)
}

func (s *Storage) Keys() ([]string, error) {
	pluginData, err := s.getPluginData(false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []string{}, nil
		}
		return nil, err
	}

	data, err := s.getDataMap(pluginData)
	if err != nil {
		return nil, err
	}

	return getAllKeys(data, ""), nil
}

func (s *Storage) Has(key string) (bool, error) {
	// Check key cache first
	if s.keyDataCache.Has(key) {
		return true, nil
	}

	pluginData, err := s.getPluginData(false)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	data, err := s.getDataMap(pluginData)
	if err != nil {
		return false, err
	}

	exists := hasNestedKey(data, key)

	// If key exists, we can also cache its value for future Get calls
	if exists {
		value := getNestedValue(data, key)
		if value != nil {
			s.keyDataCache.Set(key, value)
		}
	}

	return exists, nil
}

func (s *Storage) Get(key string) (interface{}, error) {
	s.logger.Trace().Msgf("plugin: Getting key %s", key)

	// Check key cache first
	if cachedValue, ok := s.keyDataCache.Get(key); ok {
		return cachedValue, nil
	}

	pluginData, err := s.getPluginData(true)
	if err != nil {
		return nil, err
	}

	data, err := s.getDataMap(pluginData)
	if err != nil {
		return nil, err
	}

	value := getNestedValue(data, key)

	// Cache the value
	if value != nil {
		s.keyDataCache.Set(key, value)
	}

	return value, nil
}

func (s *Storage) Set(key string, value interface{}) error {
	s.logger.Trace().Msgf("plugin: Setting key %s", key)
	pluginData, err := s.getPluginData(true)
	if err != nil {
		return err
	}

	data, err := s.getDataMap(pluginData)
	if err != nil {
		data = make(map[string]interface{})
	}

	setNestedValue(data, key, value)

	// Update key cache
	s.keyDataCache.Set(key, value)

	return s.saveDataMap(pluginData, data)
}
