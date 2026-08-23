import os, glob, hashlib, base64
from cryptography.fernet import Fernet
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

# تحويل أي كلمة سر لمفتاح تشفير عالي صحيح
raw_key = os.environ['YT_KEY'].encode()
fernet_key = base64.urlsafe_b64encode(hashlib.sha256(raw_key).digest())
f = Fernet(fernet_key)

refresh_token = f.decrypt(os.environ['YT_TOKEN_ENC'].encode()).decode()

creds = Credentials(
    None,
    refresh_token=refresh_token,
    client_id=os.environ['YT_CLIENT_ID'],
    client_secret=os.environ['YT_CLIENT_SECRET'],
    token_uri='https://oauth2.googleapis.com/token'
)

youtube = build('youtube', 'v3', credentials=creds)
print("تم فك التشفير العالي بنجاح!")

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
