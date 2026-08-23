import os
import glob
from google.oauth2.credentials import Credentials
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

# === 1. دور على الفيديو ===
videos = glob.glob("output/*.mp4")
if not videos:
    print("❌ مفيش فيديو في output/")
    exit(1)

file = videos[0]
print(f"📹 لقيت الفيديو: {file}")

# === 2. العنوان والوصف والتاجز - مع fallback ===
try:
    if os.path.exists("output/title.txt"):
        title = open("output/title.txt", encoding="utf-8").read().strip()[:95]
    else:
        raise FileNotFoundError

    if os.path.exists("output/desc.txt"):
        desc = open("output/desc.txt", encoding="utf-8").read()
    else:
        desc = f"{title}\n\nالطب الملعون - دكتور ضياء العوضي - كلام بدون فلترة"

    if os.path.exists("output/tags.txt"):
        tags = open("output/tags.txt", encoding="utf-8").read().split(",")
    else:
        tags = ["الطب الملعون","دكتور ضياء العوضي","الجلوتين","اللبن","السكر","Cursed Medicine"]

except Exception as e:
    print(f"⚠️ بيستخدم عنوان افتراضي بسبب: {e}")
    title = "الطب الملعون - 13 ثانية قد تنقذك | TAYYIBAT GOD R"
    desc = "الطب الملعون - دكتور ضياء العوضي - الطب الملعون كامل بدون حذف\n\n#الطب_الملعون #دكتور_ضياء_العوضي #CursedMedicine"
    tags = ["الطب الملعون","دكتور ضياء العوضي","الجلوتين","ارتشاح الامعاء","اللبن","السكر","Cursed Medicine","ترند مصر"]

print(f"🚀 العنوان: {title}")

# === 3. الاتصال بيوتيوب ===
creds = Credentials(
    None,
    refresh_token=os.environ['YT_TOKEN_ENC'].strip(),
    client_id=os.environ['YT_CLIENT_ID'].strip(),
    client_secret=os.environ['YT_CLIENT_SECRET'].strip(),
    token_uri='https://oauth2.googleapis.com/token'
)
yt = build('youtube', 'v3', credentials=creds)

# === 4. رفع الفيديو ===
try:
    res = yt.videos().insert(
        part="snippet,status",
        body={
            "snippet": {
                "title": title,
                "description": desc,
                "categoryId": "22",
                "tags": [t.strip() for t in tags[:15] if t.strip()]
            },
            "status": {
                "privacyStatus": "public",
                "selfDeclaredMadeForKids": False
            }
        },
        media_body=MediaFileUpload(file, resumable=True, chunksize=-1)
    ).execute()

    vid = res['id']
    url = f"https://youtu.be/{vid}"
    print(f"✅ بيرفع: {title}")
    print(f"✅ اترفع: {url}")

    # === 5. رفع الصورة المصغرة - اصلاح الخطأ اللي في الصورة ===
    thumbs = glob.glob("output/*.jpg") + glob.glob("output/*.jpeg") + glob.glob("output/*.png")
    # استبعد الصور الصغيرة
    thumbs = [t for t in thumbs if os.path.getsize(t) > 5000]

    if thumbs:
        thumb = thumbs[0]
        print(f"🖼️ بيرفع صورة مصغرة: {thumb}")
        try:
            yt.thumbnails().set(
                videoId=vid,
                media_body=MediaFileUpload(thumb, mimetype='image/jpeg')
            ).execute()
            print("✅ صورة مصغرة اترفعت")
        except Exception as e:
            print(f"⚠️ فشل رفع الصورة المصغرة: {e} - الفيديو اترفع عادي")
    else:
        print("⚠️ مفيش صورة مصغرة - الفيديو اترفع بدونها")

    print(f"🎉 خلص: {url}")

except Exception as e:
    print(f"❌ خطأ في الرفع: {e}")
    raise
