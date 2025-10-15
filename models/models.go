package models

import (
	"encoding/json"
	"time"
)

// Add your model structs here

type FileChunk struct {
	FileName    string `json:"fileName"`
	ChunkId     int    `json:"chunkId"`
	TotalChunks int    `json:"totalChunks"`
	Content     string `json:"content"`
}

type File struct {
	FID       int       `json:"fid"`
	UserID    int       `json:"userid"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	FileSize  int64     `json:"fileSize"`
	Hash      string    `json:"hash"`
	UploadAt  time.Time `json:"uploadAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToJsonString(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}
