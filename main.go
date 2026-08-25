package main

import (
	"fmt"
	"os"
	"os/exec"
	"math"
	"encoding/binary"
	"time"
)

func genCalm(fn string, freqs []float64){
	f,_:=os.Create(fn)
	defer f.Close()
	sr:=44100
	dur:=900 // 15 دقيقة
	n:=sr*dur
	f.Write([]byte("RIFF"))
	binary.Write(f,binary.LittleEndian,uint32(36+n*2))
	f.Write([]byte("WAVEfmt "))
	binary.Write(f,binary.LittleEndian,uint32(16))
	binary.Write(f,binary.LittleEndian,uint16(1))
	binary.Write(f,binary.LittleEndian,uint32(sr))
	binary.Write(f,binary.LittleEndian,uint32(sr*2))
	binary.Write(f,binary.LittleEndian,uint16(2))
	binary.Write(f,binary.LittleEndian,uint16(16))
	f.Write([]byte("data"))
	binary.Write(f,binary.LittleEndian,uint32(n*2))
	for i:=0;i<n;i++{
		t:=float64(i)/float64(sr)
		idx:=int(math.Floor(t*0.8))%len(freqs)
		s:=math.Sin(2*math.Pi*freqs[idx]*t)*0.32 + math.Sin(2*math.Pi*freqs[idx]*0.5*t)*0.18
		s*= (1+0.12*math.Sin(2*math.Pi*5*t)) * math.Pow(math.Sin(math.Pi*float64(i)/float64(n)),0.25) *0.35
		binary.Write(f,binary.LittleEndian,int16(s*32767*0.6))
	}
}

func makeHook(out string){
	exec.Command("melt", "color:black", "out=70",
		"-filter", "pango:text='$1000 -> $10000\nحلال 100%':family=Sans:size=115:fgcolour=gold:weight=bold:align=center",
		"-filter", "affine:from=0=50%=100%=50%:to=0=0=100%=100%:0.08",
		"-consumer", "avformat:hook1.mp4", "s=3840x2160", "fps=30", "vcodec=libx264", "vb=30M").Run()
	exec.Command("melt", "color:0x0a192f", "out=70",
		"-filter", "pango:text='بدون ربا\n[DUTCH 15°]':family=Sans:size=95:fgcolour=white:weight=bold:align=center",
		"-filter", "affine:from=0=0=100%=100%:to=-5%=-5%=105%=105%:15",
		"-consumer", "avformat:hook2.mp4", "s=3840x2160", "fps=30", "vcodec=libx264", "vb=30M").Run()
	exec.Command("melt", "color:0x1a3a5f", "out=70",
		"-filter", "pango:text='السر 15 دقيقة\n3-2-1':family=Sans:size=105:fgcolour=#00ff88:weight=bold:align=center",
		"-consumer", "avformat:hook3.mp4", "s=3840x2160", "fps=30", "vcodec=libx264", "vb=30M").Run()
	exec.Command("melt", "hook1.mp4", "hook2.mp4", "-transition", "luma:duration=10",
		"hook3.mp4", "-transition", "luma:duration=10",
		"-consumer", "avformat:"+out, "s=3840x2160", "fps=30", "vcodec=libx264", "vb=30M").Run()
}

func makeAngle(out, color, text, angle string, sec int){
	filters := map[string]string{
		"low_hero": "affine:from=0=50%=100%=50%:to=0=0=100%=100%:0.001",
		"bird_eye": "affine:from=0=0=100%=100%:to=10%=10%=80%=80%:0.001",
		"worm_eye": "affine:from=0=100%=100%=100%:to=0=0=100%=100%:0.001",
		"dutch_15": "affine:from=0=0=100%=100%:to=-5%=-5%=105%=105%:15",
		"pov": "affine:from=0=0=100%=100%:to=2%=2%=102%=102%:0",
		"closeup": "affine:from=25%=25%=75%=75%:to=0=0=100%=100%:0.002",
		"wide": "affine:from=0=0=100%=100%:to=-10%=-10%=110%=110%:0.0005",
		"tracking": "affine:from=0=0=100%=100%:to=-8%=0=92%=100%:0.001",
	}
	exec.Command("melt", fmt.Sprintf("color:%s", color), fmt.Sprintf("out=%d", sec*30),
		"-filter", filters[angle],
		"-filter", "lift_gamma_gain:lift_b=0.12:gamma_r=1.06:gain_r=1.1",
		"-filter", "vignette:radius=75%:softness=40%:opacity=0.4",
		"-filter", fmt.Sprintf("pango:text='%s\n[%s]':family=Sans:size=58:fgcolour=#f0d78c:weight=bold:align=center:pad=30:shadow=3", text, angle),
		"-consumer", "avformat:"+out, "s=3840x2160", "fps=30", "vcodec=libx264", "vb=28M").Run()
}

func main(){
	fmt.Println("7 SEC HOOK + 10 HIDDEN ANGLES - 15 MIN - FIXED V44")
	genCalm("calm.wav", []float64{432,540,648,528,396})

	for vid:=1; vid<=4; vid++{
		fmt.Printf("⚡ VIDEO %d START\n", vid)
		makeHook(fmt.Sprintf("INTRO_7SEC_%d.mp4", vid))
		angles := []string{"low_hero","bird_eye","worm_eye","dutch_15","pov","closeup","wide","tracking"}
		colors := []string{"0x0a192f","0x132a4c","0x1a3a5f","0x0f2d4d","0x1e1a3a","0x0f3d2e","0x2a1a1a","0x1a2a1a"}
		names := []string{"LOW HERO - هيبة","BIRD EYE - سيطرة","WORM EYE - ضخامة","DUTCH 15° - توتر","POV - انت البطل","CLOSE-UP GOLD","WIDE - 195 دولة","TRACKING DOLLY"}
		for idx, ang := range angles {
			makeAngle(fmt.Sprintf("angle%d_%s.mp4", vid, ang), colors[idx], fmt.Sprintf("VIDEO %d - %d/8\n%s", vid, idx+1, names[idx]), ang, 110)
		}
		final := fmt.Sprintf("FINAL_15MIN_7SEC_HOOK_%d_%s.mp4", vid, time.Now().Format("2006-01-02"))
		args := []string{fmt.Sprintf("INTRO_7SEC_%d.mp4", vid)}
		for _, ang := range angles {
			args = append(args, fmt.Sprintf("angle%d_%s.mp4", vid, ang), "-transition", "luma:duration=25")
		}
		args = append(args, "-track", "calm.wav", "-consumer", "avformat:"+final, "s=3840x2160", "fps=30", "vcodec=libx264", "vb=28M", "acodec=aac", "ab=128k")
		exec.Command("melt", args...).Run()
		fmt.Printf("✅ VIDEO %d READY - 15 MIN - 7 SEC HOOK + 8 ANGLES\n", vid)
	}
}
