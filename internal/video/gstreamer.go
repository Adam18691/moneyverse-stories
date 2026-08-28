package video

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

type GStreamerProcessor struct {
	logger *zap.Logger
	mu     sync.Mutex
	cmd    *exec.Cmd
}

// NewGStreamerProcessor ينشئ معالج GStreamer جديد
func NewGStreamerProcessor(logger *zap.Logger) *GStreamerProcessor {
	return &GStreamerProcessor{
		logger: logger,
	}
}

// ProcessPipeline معالجة فيديو عبر خط أنابيب GStreamer
func (g *GStreamerProcessor) ProcessPipeline(ctx context.Context, pipeline string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.logger.Info("بدء معالجة GStreamer", zap.String("pipeline", pipeline))

	g.cmd = exec.CommandContext(ctx, "gst-launch-1.0", "-e", pipeline)
	
	// تشغيل مع معالجة الأخطاء
	if err := g.cmd.Run(); err != nil {
		g.logger.Error("فشل خط أنابيب GStreamer", zap.Error(err))
		return err
	}

	g.logger.Info("اكتملت معالجة GStreamer بنجاح")
	return nil
}

// ProcessVideo معالجة ملف فيديو كامل
func (g *GStreamerProcessor) ProcessVideo(ctx context.Context, inputPath, outputPath string, filters map[string]string) error {
	// بناء خط الأنابيب ديناميكياً
	pipeline := fmt.Sprintf(
		"filesrc location=%s ! decodebin ! videoconvert ! %s ! x264enc bitrate=%s ! mux. filesink location=%s",
		inputPath,
		g.buildFilters(filters),
		filters["bitrate"],
		outputPath,
	)

	return g.ProcessPipeline(ctx, pipeline)
}

// buildFilters بناء سلسلة الفلاتر
func (g *GStreamerProcessor) buildFilters(filters map[string]string) string {
	filterStr := "videoconvert"

	// Color correction
	if brightness, ok := filters["brightness"]; ok {
		filterStr += fmt.Sprintf(" ! colorcurve brightness=%s", brightness)
	}

	// Scaling
	if width, ok := filters["width"]; ok {
		if height, ok := filters["height"]; ok {
			filterStr += fmt.Sprintf(" ! videoscale width=%s height=%s", width, height)
		}
	}

	// Frame rate
	if fps, ok := filters["fps"]; ok {
		filterStr += fmt.Sprintf(" ! videorate ! video/x-raw,framerate=%s/1", fps)
	}

	return filterStr
}

// Stop إيقاف العملية الحالية
func (g *GStreamerProcessor) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.cmd != nil && g.cmd.Process != nil {
		g.cmd.Process.Kill()
	}
}
