import os, subprocess, random
from datetime import datetime

os.makedirs("output", exist_ok=True)

TITLES = [
    "الطب الملعون | اشارة خطيرة في الكلى",
    "Cursed Medicine | سر التعب المستمر",
    "جسمك يصرخ | علامة لا تتجاهلها",
    "KIDNEYS - Warning Sign You Miss",
    "الطب الملعون | 13 ثانية قد تنقذك"
]

title = random.choice(TITLES)
out = "output/tayyibat_10min_8K_60fps_STUDIO_FINAL.mp4"

print(f"TAYYIBAT GOD R - {title}")

cmd = [
    "ffmpeg","-y",
    "-f","lavfi","-i","color=c=0x0a192f:s=1920x1080:r=30:d=600",
    "-f","lavfi","-i","anullsrc=r=48000:cl=stereo:d=600",
    "-vf", f"drawtext=text='{title}':fontcolor=white:fontsize=55:x=(w-text_w)/2:y=900:box=1:boxcolor=black@0.5:boxborderw=12",
    "-c:v","libx264","-profile:v","high","-pix_fmt","yuv420p","-preset","medium","-crf","22","-r","30",
    "-c:a","aac","-b:a","192k","-ar","48000","-movflags","+faststart",
    out
]
subprocess.run(cmd, check=True)
print(f"✅ تم: {out}")
