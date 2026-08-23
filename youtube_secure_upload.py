import os, glob, subprocess
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

videos = glob.glob("output/*.mp4")
if not videos:
    print("مفيش فيديو")
    exit(0)

orig = videos[0]
fixed = "output/final_youtube.mp4"

print(f"بعالج: {orig}")

# تحويل شامل لكل الملف
cmd = [
    "ffmpeg", "-y",
    "-i", orig,
    "-vf", "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2",
    "-r", "30",
    "-c:v", "libx264", "-profile:v", "high", "-pix_fmt", "yuv420p",
    "-preset", "medium", "-crf", "20",
    "-c:a", "aac", "-b:a", "192k", "-ar", "48000",
    "-movflags", "+faststart",
    fixed
]
subprocess.run(cmd, check=True)
print("✅ تمت المعالجة الكاملة")

# رفع
creds = Credentials(
    None,
    refresh_token=os.environ['YT_TOKEN_ENC'].strip(),
    client_id=os.environ['YT_CLIENT_ID'].strip(),
    client_secret=os.environ['YT_CLIENT_SECRET'].strip(),
    token_uri='https://oauth2.googleapis.com/token'
)
yt = build('youtube', 'v3', credentials=creds)
req = yt.videos().insert(
    part="snippet,status",
    body={"snippet":{"title":"Tayyibat - تلاوة خاشعة | الطب الملعون","description":"Cursed Medicine | تلاوة","categoryId":"22"},"status":{"privacyStatus":"public"}},
    media_body=MediaFileUpload(fixed, resumable=True)
)
res = req.execute()
print(f"✅ اترفع: https://youtu.be/{res['id']}")
