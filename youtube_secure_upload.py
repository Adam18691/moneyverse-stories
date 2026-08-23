import os, glob
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

# مؤقتا بدون Fernet لحد ما نحدث الـ secret
refresh_token = os.environ['YT_TOKEN_ENC'] # حط فيه التوكن الخام مؤقتا

creds = Credentials(
    None,
    refresh_token=refresh_token,
    client_id=os.environ['YT_CLIENT_ID'],
    client_secret=os.environ['YT_CLIENT_SECRET'],
    token_uri='https://oauth2.googleapis.com/token'
)
youtube = build('youtube', 'v3', credentials=creds)
print("دخول مباشر - هنرجع التشفير بعد ما يشتغل")

videos = glob.glob("output/*.mp4")
if videos:
    file_path = videos[0]
    req = youtube.videos().insert(
        part="snippet,status",
        body={"snippet":{"title":"Tayyibat - تلاوة","description":"#قران","categoryId":"22"},"status":{"privacyStatus":"public"}},
        media_body=MediaFileUpload(file_path)
    )
    res = req.execute()
    print(f"اترفع: https://youtu.be/{res['id']}")
