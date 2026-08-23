import os, glob
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

print("TAYYIBAT - رفع مباشر بدون تشفير")

# هيقرا التوكن الخام مباشر من Secrets
refresh_token = os.environ['YT_TOKEN_ENC'].strip()
client_id = os.environ['YT_CLIENT_ID'].strip()
client_secret = os.environ['YT_CLIENT_SECRET'].strip()

creds = Credentials(
    None,
    refresh_token=refresh_token,
    client_id=client_id,
    client_secret=client_secret,
    token_uri='https://oauth2.googleapis.com/token'
)

youtube = build('youtube', 'v3', credentials=creds)
print("تم تسجيل الدخول ليوتوب بنجاح!")

videos = glob.glob("output/*.mp4")
if not videos:
    print("مفيش فيديو في output/")
    exit(0)

file_path = videos[0]
print(f"هنرفع: {file_path}")

request = youtube.videos().insert(
    part="snippet,status",
    body={
        "snippet": {
            "title": "Tayyibat - تلاوة خاشعة",
            "description": "تلاوة من مشروع الطيبات #قران",
            "categoryId": "22",
            "tags": ["قران", "تلاوة"]
        },
        "status": {"privacyStatus": "public"}
    },
    media_body=MediaFileUpload(file_path, chunksize=-1, resumable=True)
)

response = request.execute()
print(f"✅ اترفع بنجاح: https://youtu.be/{response['id']}")
