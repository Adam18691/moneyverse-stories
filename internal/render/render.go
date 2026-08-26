package render

import (
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
)

// Build: GStreamer pipeline (بدون FFmpeg) + fallback melt
func Build(scriptKey, audioTrack, outPath string) error {
	h := fnv.New32a()
	h.Write([]byte(scriptKey))

	cmd := exec.Command("gst-launch-1.0", "-e",
		"multifilesrc", fmt.Sprintf("location=scenes/%d_%%03d.png", h.Sum32()%10000),
		"!", "image/png", "width=1920,height=1080", "framerate=2/1",
		"!", "videoconvert", "!", "videoscale",
		"!", "x264enc", "speed-preset=ultrafast", "quantizer=28", "threads=0",
		"key-int-max=120",
		"!", "mp4mux", "!", "filesink", "location="+outPath,
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err == nil {
		return nil
	}

	// fallback: MLT melt
	m := exec.Command("melt", "audio="+audioTrack,
		"-consumer", "avformat:"+outPath,
		"vcodec=libx264", "preset=ultrafast", "crf=28", "threads=0")
	m.Stdout, m.Stderr = os.Stdout, os.Stderr
	return m.Run()
}
