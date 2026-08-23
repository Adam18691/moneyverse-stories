import os
import subprocess
import random
from datetime import datetime

# إعدادات الذروة
OUTPUT_DIR = "output"
os.makedirs(OUTPUT_DIR, exist_ok=True)

# عناوين ذروة مصر - تجيب مشاهدات
TITLES = [
    "الطب الملعون | سر لا يخبرك به طبيبك",
    "Cursed Medicine | علامة في الكلى تنذر بالخطر",
    "جسمك يصرخ وهذه الإشارة | الطب الملعون",
    "KIDNEYS Warning Sign You Must Know",
    "السر وراء التعب المستمر | Cursed Medicine"
]

title = random.choice(TITLES)
filename = f"{OUTPUT_DIR}/tayyibat_10min_8K_60fps_STUDIO_FINAL.mp4"

print(f"""

TAYYIBAT MEGA V14 - GOD R - 60

العنوان: {title}
الوقت: {datetime.now()}
وقت الذروة: 7م - 11م بتوقيت القاهرة
""")

# فيديو 10 دقايق = 600 ثانية - متوافق 100% مع يوتيوب
# 1080p 30fps - H264 + yuv420p = يوتيوب يقبله فوراً
ffmpeg_cmd = [
    "ffmpeg", "-y",
    "-f", "lavfi", "-i", f"color=c=0x0a192f:s=1920x1080:r=30:d=600",
    "-f", "lavfi", "-i", "anullsrc=r=48000:cl=stereo:d=600",
    "-vf", f"drawtext=text='{title}':fontcolor=white:fontsize=52:x=(w-text_w)/2:y=900:box=1:boxcolor=black@0.6:boxborderw=10",
    "-c:v", "libx264",
    "-profile:v", "high",
    "-pix_fmt", "yuv420p",
    "-preset", "medium",
    "-crf", "22",
    "-r", "30",
    "-c:a", "aac",
    "-b:a", "192k",
    "-ar", "48000",
    "-movflags", "+faststart",
    filename
]

try:
    subprocess.run(ffmpeg_cmd, check=True)
    size = os.path.getsize(filename) / (1024*1024)
    print(f"✅ جاهز: {filename} - {size:.1f} MB")
    print(f"✅ العنوان: {title}")
except Exception as e:
    print(f"❌ فشل: {e}")
    exit(1)
