import os
import glob
from cryptography.fernet import Fernet
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

# فك التشفير العالي AES
key = os.environ['YT_KEY'].encode()
enc_token = os.environ['YT_TOKEN_ENC'].encode()

f = Fernet(key)
refresh_token = f.decrypt(enc_token).decode()

# تسجيل دخول يوتيوب بتشفير TLS 1.3
creds = Credentials(
    None,
    refresh_token=refresh_token,
    client_id=os.environ['YT_CLIENT_ID'],
    client_secret=os.environ['YT_CLIENT_SECRET'],
    token_uri='https://oauth2.googleapis.com/token'
)

youtube = build('youtube', 'v3', credentials=creds)
print("تم فك التشفير والدخول ليوتيوب بنجاح - التشفير عالي!")

# رفع الفيديو
videos = glob.glob("output/*.mp4")
if videos:
    file_path = videos[0]
    request = youtube.videos().insert(
        part="snippet,status",
        body={
            "snippet": {"title": "Tayyibat - تلاوة طيبة", "description": "#قران #Tayyibat", "categoryId": "22"},
            "status": {"privacyStatus": "public"}
        },
        media_body=MediaFileUpload(file_path)
    )
    response = request.execute()
    print(f"تم الرفع: https://youtu.be/{response['id']}")
else:
    print("مفيش فيديو في output/")
