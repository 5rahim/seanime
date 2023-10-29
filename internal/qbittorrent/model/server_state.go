package qbittorrent_model

import (
	"encoding/json"
	"strconv"
)

type ServerState struct {
	TransferInfo
	AlltimeDl            int     `json:"alltime_dl"`
	AlltimeUl            int     `json:"alltime_ul"`
	AverageTimeQueue     int     `json:"average_time_queue"`
	FreeSpaceOnDisk      int     `json:"free_space_on_disk"`
	GlobalRatio          float64 `json:"global_ratio"`
	QueuedIoJobs         int     `json:"queued_io_jobs"`
	ReadCacheHits        float64 `json:"read_cache_hits"`
	ReadCacheOverload    float64 `json:"read_cache_overload"`
	TotalBuffersSize     int     `json:"total_buffers_size"`
	TotalPeerConnections int     `json:"total_peer_connections"`
	TotalQueuedSize      int     `json:"total_queued_size"`
	TotalWastedSession   int     `json:"total_wasted_session"`
	WriteCacheOverload   float64 `json:"write_cache_overload"`
}

func (s *ServerState) UnmarshalJSON(data []byte) error {
	var raw rawServerState
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	globalRatio, err := strconv.ParseFloat(raw.GlobalRatio, 64)
	if err != nil {
		return err
	}
	readCacheHits, err := strconv.ParseFloat(raw.ReadCacheHits, 64)
	if err != nil {
		return err
	}
	readCacheOverload, err := strconv.ParseFloat(raw.ReadCacheOverload, 64)
	if err != nil {
		return err
	}
	writeCacheOverload, err := strconv.ParseFloat(raw.WriteCacheOverload, 64)
	if err != nil {
		return err
	}
	*s = ServerState{
		TransferInfo:         raw.TransferInfo,
		AlltimeDl:            raw.AlltimeDl,
		AlltimeUl:            raw.AlltimeUl,
		AverageTimeQueue:     raw.AverageTimeQueue,
		FreeSpaceOnDisk:      raw.FreeSpaceOnDisk,
		GlobalRatio:          globalRatio,
		QueuedIoJobs:         raw.QueuedIoJobs,
		ReadCacheHits:        readCacheHits,
		ReadCacheOverload:    readCacheOverload,
		TotalBuffersSize:     raw.TotalBuffersSize,
		TotalPeerConnections: raw.TotalPeerConnections,
		TotalQueuedSize:      raw.TotalQueuedSize,
		TotalWastedSession:   raw.TotalWastedSession,
		WriteCacheOverload:   writeCacheOverload,
	}
	return nil
}

type rawServerState struct {
	TransferInfo
	AlltimeDl            int    `json:"alltime_dl"`
	AlltimeUl            int    `json:"alltime_ul"`
	AverageTimeQueue     int    `json:"average_time_queue"`
	FreeSpaceOnDisk      int    `json:"free_space_on_disk"`
	GlobalRatio          string `json:"global_ratio"`
	QueuedIoJobs         int    `json:"queued_io_jobs"`
	ReadCacheHits        string `json:"read_cache_hits"`
	ReadCacheOverload    string `json:"read_cache_overload"`
	TotalBuffersSize     int    `json:"total_buffers_size"`
	TotalPeerConnections int    `json:"total_peer_connections"`
	TotalQueuedSize      int    `json:"total_queued_size"`
	TotalWastedSession   int    `json:"total_wasted_session"`
	WriteCacheOverload   string `json:"write_cache_overload"`
}
