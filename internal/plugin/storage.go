package plugin

import (
	"encoding/json"
	"errors"
	"seanime/internal/database/models"
	"seanime/internal/extension"
	goja_util "seanime/internal/util/goja"
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
	runtime         *goja.Runtime
	pluginDataCache *result.Map[string, *models.PluginData] // Cache to avoid repeated database calls
	keyDataCache    *result.Map[string, interface{}]        // Cache to avoid repeated database calls
	keySubscribers  *result.Map[string, []chan interface{}] // Subscribers for key changes
	scheduler       *goja_util.Scheduler
}

var (
	ErrDatabaseNotInitialized = errors.New("database is not initialized")
)

// BindStorage binds the storage API to the Goja runtime.
// Permissions need to be checked by the caller.
// Permissions needed: storage
func (a *AppContextImpl) BindStorage(vm *goja.Runtime, logger *zerolog.Logger, ext *extension.Extension, scheduler *goja_util.Scheduler) *Storage {
	storageLogger := logger.With().Str("id", ext.ID).Logger()
	storage := &Storage{
		ctx:             a,
		ext:             ext,
		logger:          &storageLogger,
		runtime:         vm,
		pluginDataCache: result.NewResultMap[string, *models.PluginData](),
		keyDataCache:    result.NewResultMap[string, interface{}](),
		keySubscribers:  result.NewResultMap[string, []chan interface{}](),
		scheduler:       scheduler,
	}
	storageObj := vm.NewObject()
	_ = storageObj.Set("get", storage.Get)
	_ = storageObj.Set("set", storage.Set)
	_ = storageObj.Set("remove", storage.Delete)
	_ = storageObj.Set("drop", storage.Drop)
	_ = storageObj.Set("clear", storage.Clear)
	_ = storageObj.Set("keys", storage.Keys)
	_ = storageObj.Set("has", storage.Has)
	_ = storageObj.Set("watch", storage.Watch)
	_ = vm.Set("$storage", storageObj)

	return storage
}

// Stop closes all subscriber channels.
func (s *Storage) Stop() {
	s.keySubscribers.Range(func(key string, subscribers []chan interface{}) bool {
		for _, ch := range subscribers {
			close(ch)
		}
		return true
	})
	s.keySubscribers.Clear()
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

	var pluginData models.PluginData
	if err := db.Where("plugin_id = ?", s.ext.ID).First(&pluginData).Error; err != nil {
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
	s.pluginDataCache.Set(s.ext.ID, &pluginData)
	return &pluginData, nil
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
	s.keyDataCache.Clear()

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

// notifyKeyAndParents sends notifications to subscribers of the given key and its parent keys
// If the value is nil, it indicates the key was deleted
func (s *Storage) notifyKeyAndParents(key string, value interface{}, data map[string]interface{}) {
	// Notify direct subscribers of this key
	if subscribers, ok := s.keySubscribers.Get(key); ok {
		for _, ch := range subscribers {
			// Non-blocking send to avoid deadlocks
			select {
			case ch <- value:
			default:
				// Channel is full or closed, skip
			}
		}
	}

	// Also notify parent key subscribers if this is a nested key
	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		for i := 1; i < len(parts); i++ {
			parentKey := strings.Join(parts[:i], ".")
			if subscribers, ok := s.keySubscribers.Get(parentKey); ok {
				// Get the current parent value
				parentValue := getNestedValue(data, parentKey)
				for _, ch := range subscribers {
					// Non-blocking send to avoid deadlocks
					select {
					case ch <- parentValue:
					default:
						// Channel is full or closed, skip
					}
				}
			}
		}
	}
}

func (s *Storage) Watch(key string, callback goja.Callable) goja.Value {
	s.logger.Trace().Msgf("plugin: Watching key %s", key)

	// Create a channel to receive updates
	updateCh := make(chan interface{}, 100)

	// Add this channel to the subscribers for this key
	subscribers := []chan interface{}{}
	if existingSubscribers, ok := s.keySubscribers.Get(key); ok {
		subscribers = existingSubscribers
	}
	subscribers = append(subscribers, updateCh)
	s.keySubscribers.Set(key, subscribers)

	// Start a goroutine to listen for updates
	go func() {
		for value := range updateCh {
			// Call the callback with the new value
			s.scheduler.ScheduleAsync(func() error {
				_, err := callback(goja.Undefined(), s.runtime.ToValue(value))
				if err != nil {
					s.logger.Error().Err(err).Msgf("plugin: Error calling watch callback for key %s", key)
				}
				return nil
			})
		}
	}()

	// Check if the key currently exists and immediately send its value
	// This allows watchers to get the current value right away
	currentValue, _ := s.Get(key)
	if currentValue != nil {
		// Use non-blocking send
		select {
		case updateCh <- currentValue:
		default:
			// Channel is full, skip
		}
	}

	// Return a function that can be used to cancel the watch
	cancelFn := func() {
		close(updateCh)
		// Remove this specific channel from subscribers
		if existingSubscribers, ok := s.keySubscribers.Get(key); ok {
			newSubscribers := make([]chan interface{}, 0, len(existingSubscribers)-1)
			for _, ch := range existingSubscribers {
				if ch != updateCh {
					newSubscribers = append(newSubscribers, ch)
				}
			}

			if len(newSubscribers) > 0 {
				s.keySubscribers.Set(key, newSubscribers)
			} else {
				s.keySubscribers.Delete(key)
			}
		}
	}

	return s.runtime.ToValue(cancelFn)
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

	// Notify subscribers that the key was deleted
	s.notifyKeyAndParents(key, nil, data)

	if deleteNestedValue(data, key) {
		return s.saveDataMap(pluginData, data)
	}

	return nil
}

func (s *Storage) Drop() error {
	s.logger.Trace().Msg("plugin: Dropping storage")

	// // Close all subscriber channels
	// s.keySubscribers.Range(func(key string, subscribers []chan interface{}) bool {
	// 	for _, ch := range subscribers {
	// 		close(ch)
	// 	}
	// 	return true
	// })
	// s.keySubscribers.Clear()

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

	// Get all keys before clearing
	data, err := s.getDataMap(pluginData)
	if err != nil {
		return err
	}

	// Get all keys to notify subscribers
	keys := getAllKeys(data, "")

	// Create empty data map
	cleanData := make(map[string]interface{})

	// Save the empty data first
	if err := s.saveDataMap(pluginData, cleanData); err != nil {
		return err
	}

	// Notify all subscribers that their keys were cleared
	for _, key := range keys {
		s.notifyKeyAndParents(key, nil, cleanData)
	}

	return nil
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

	// Notify subscribers
	s.notifyKeyAndParents(key, value, data)

	return s.saveDataMap(pluginData, data)
}
