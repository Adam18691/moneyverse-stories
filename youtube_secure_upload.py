import os, glob
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

files = glob.glob("output/*.mp4")
if not files:
    print("مفيش فيديو")
    exit(0)

file = files[0]
print(f"بيرفع: {file}")

creds = Credentials(None,
    refresh_token=os.environ['YT_TOKEN_ENC'].strip(),
    client_id=os.environ['YT_CLIENT_ID'].strip(),
    client_secret=os.environ['YT_CLIENT_SECRET'].strip(),
    token_uri='https://oauth2.googleapis.com/token')

yt = build('youtube','v3',credentials=creds)

# عنوان من اسم الملف
title = "Cursed Medicine | " + os.path.basename(file)[:50]

req = yt.videos().insert(
    part="snippet,status",
    body={
        "snippet":{"title":title,"description":"الطب الملعون - Cursed Medicine | Mostafa Mahmoud #الطب_الملعون","categoryId":"22","tags":["الطب الملعون","Cursed Medicine"]},
        "status":{"privacyStatus":"public","selfDeclaredMadeForKids":False}
    },
    media_body=MediaFileUpload(file,resumable=True,chunksize=-1)
)
res = req.execute()
print(f"✅ ارفع: https://youtu.be/{res['id']}")
