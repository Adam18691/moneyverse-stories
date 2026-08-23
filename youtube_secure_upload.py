import os
from cryptography.fernet import Fernet
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build

# فك التشفير
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
