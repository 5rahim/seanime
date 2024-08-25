package extension

//import (
//	hibikemediaplayer "github.com/5rahim/hibike/pkg/extension/mediaplayer"
//)
//
//type MediaPlayerExtension interface {
//	BaseExtension
//	GetMediaPlayer() hibikemediaplayer.MediaPlayer
//}
//
//type MediaPlayerExtensionImpl struct {
//	ext      *Extension
//	provider hibikemediaplayer.MediaPlayer
//}
//
//func NewMediaPlayerExtension(ext *Extension, provider hibikemediaplayer.MediaPlayer) MediaPlayerExtension {
//	return &MediaPlayerExtensionImpl{
//		ext:      ext,
//		provider: provider,
//	}
//}
//
//func (m *MediaPlayerExtensionImpl) GetMediaPlayer() hibikemediaplayer.MediaPlayer {
//	return m.provider
//}
//
//func (m *MediaPlayerExtensionImpl) GetExtension() *Extension {
//	return m.ext
//}
//
//func (m *MediaPlayerExtensionImpl) GetType() Type {
//	return m.ext.Type
//}
//
//func (m *MediaPlayerExtensionImpl) GetID() string {
//	return m.ext.ID
//}
//
//func (m *MediaPlayerExtensionImpl) GetName() string {
//	return m.ext.Name
//}
//
//func (m *MediaPlayerExtensionImpl) GetVersion() string {
//	return m.ext.Version
//}
//
//func (m *MediaPlayerExtensionImpl) GetManifestURI() string {
//	return m.ext.ManifestURI
//}
//
//func (m *MediaPlayerExtensionImpl) GetLanguage() Language {
//	return m.ext.Language
//}
//
//func (m *MediaPlayerExtensionImpl) GetDescription() string {
//	return m.ext.Description
//}
//
//func (m *MediaPlayerExtensionImpl) GetAuthor() string {
//	return m.ext.Author
//}
//
//func (m *MediaPlayerExtensionImpl) GetPayload() string {
//	return m.ext.Payload
//}
//
//func (m *MediaPlayerExtensionImpl) GetWebsite() string {
//	return m.ext.Website
//}
//
//func (m *MediaPlayerExtensionImpl) GetIcon() string {
//	return m.ext.Icon
//}
//
//func (m *MediaPlayerExtensionImpl) GetScopes() []string {
//	return m.ext.Scopes
//}
//
//func (m *MediaPlayerExtensionImpl) GetConfig() Config {
//	return m.ext.Config
//}
