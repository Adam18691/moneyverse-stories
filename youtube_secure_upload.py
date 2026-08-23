import os, glob
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

file = glob.glob("output/*.mp4")[0]

# يقرأ الجديد ولو مش موجود يقرأ القديم
try:
    title = open("output/title.txt", encoding="utf-8").read().strip()[:95]
    desc = open("output/desc.txt", encoding="utf-8").read()
    tags = open("output/tags.txt", encoding="utf-8").read().split(",")
except FileNotFoundError:
    desc = open("output/desc.txt", encoding="utf-8").read() if os.path.exists("output/desc.txt") else "الطب الملعون - دكتور ضياء العوضي"
    title = desc.splitlines()[0][:95]
    tags = ["الطب الملعون","دكتور ضياء العوضي","الجلوتين","اللبن","السكر","Cursed Medicine"]

creds = Credentials(None, refresh_token=os.environ['YT_TOKEN_ENC'].strip(), client_id=os.environ['YT_CLIENT_ID'].strip(), client_secret=os.environ['YT_CLIENT_SECRET'].strip(), token_uri='https://oauth2.googleapis.com/token')
yt = build('youtube','v3', credentials=creds)

print(f"🚀 بيرفع: {title}")

res = yt.videos().insert(
    part="snippet,status",
    body={"snippet":{"title":title,"description":desc,"categoryId":"22","tags":tags[:15]},"status":{"privacyStatus":"public"}},
    media_body=MediaFileUpload(file, resumable=True, chunksize=-1)
).execute()

vid = res['id']
print(f"✅ اترفع: https://youtu.be/{vid}")

thumb = "output/thumbnail_pro_max.jpg"
if not os.path.exists(thumb):
    thumb = glob.glob("output/*.jpg")[0]

yt.thumbnails().set(videoId=vid, media_body=MediaFileUpload(thumb, mimetype='image/jpeg')).execute()
print("✅ صورة مصغرة اترفعت")
