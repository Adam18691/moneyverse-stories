package video

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
)

type MLTRenderer struct {
	logger      *zap.Logger
	mu          sync.Mutex
	xmlPath     string
	outputPath  string
	maxWorkers  int
	semaphore   chan struct{}
}

// MLTProject تمثيل ملف المشروع XML
type MLTProject struct {
	XMLName   xml.Name      `xml:"mlt"`
	Version   string        `xml:"version,attr"`
	Producer  MLTProducer   `xml:"producer"`
	Playlist  MLTPlaylist   `xml:"playlist"`
	Tractor   MLTTractor    `xml:"tractor"`
	Consumers []MLTConsumer `xml:"consumer"`
}

type MLTProducer struct {
	ID       string         `xml:"id,attr"`
	Resource string         `xml:"property[name=resource]"`
	Clips    []MLTClip      `xml:"property[name=resource]"`
}

type MLTPlaylist struct {
	ID    string      `xml:"id,attr"`
	Entry []MLTEntry  `xml:"entry"`
}

type MLTEntry struct {
	Producer string `xml:"producer,attr"`
	In       string `xml:"in,attr"`
	Out      string `xml:"out,attr"`
}

type MLTTractor struct {
	ID    string        `xml:"id,attr"`
	Track []MLTTrack    `xml:"track"`
}

type MLTTrack struct {
	Producer string `xml:"producer,attr"`
	Hide     string `xml:"hide,attr"`
}

type MLTConsumer struct {
	ID       string              `xml:"id,attr"`
	Resource string              `xml:"property[name=resource]"`
	Target   string              `xml:"property[name=target]"`
	Profiles []MLTProfile        `xml:"property"`
}

type MLTProfile struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

type MLTClip struct {
	Path    string
	In
