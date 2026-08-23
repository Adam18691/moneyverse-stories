import os, glob
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

file = glob.glob("output/*.mp4")[0]
title = open("output/title.txt", encoding="utf-8").read().strip()[:98]
desc = open("output/desc.txt", encoding="utf-8").read()
tags = open("output/tags.txt", encoding="utf-8").read().split(",")

creds = Credentials(None, refresh_token=os.environ['YT_TOKEN_ENC'].strip(), client_id=os.environ['YT_CLIENT_ID'].strip(), client_secret=os.environ['YT_CLIENT_SECRET'].strip(), token_uri='https://oauth2.googleapis.com/token')
yt = build('youtube','v3', credentials=creds)

res = yt.videos().insert(
    part="snippet,status",
    body={
        "snippet":{"title":title,"description":desc,"categoryId":"22","tags":tags[:15]},
        "status":{"privacyStatus":"public","selfDeclaredMadeForKids":False}
    },
    media_body=MediaFileUpload(file, resumable=True, chunksize=-1)
).execute()

vid = res['id']
yt.thumbnails().set(videoId=vid, media_body=MediaFileUpload("output/thumbnail_pro_max.jpg", mimetype='image/jpeg')).execute()
print(f"✅ ULTIMATE SEO اترفع - اعلى CTR: https://youtu.be/{vid}\n{title}")
