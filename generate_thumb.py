from PIL import Image, ImageDraw, ImageFont
import sys, os
os.makedirs("output/thumbnails", exist_ok=True)
code = sys.argv[1] if len(sys.argv)>1 else "EG"
name = sys.argv[2] if len(sys.argv)>2 else "مصر"

img = Image.new('RGB', (1280,720), (8,8,20))
draw = ImageDraw.Draw(img)
# هالة ذهبية بروفيشنال
for r in range(300,0,-6):
    c = int(180 + r*0.2)
    draw.ellipse([450-r, 360-r, 450+r, 360+r], outline=(c, int(c*0.55), 0))
# سيلويت عضلي
draw.rectangle([320,70,580,660], fill=(12,12,12))
# سهم نار برتقالي طالع
draw.polygon([(450,480),(420,530),(480,530)], fill=(255,200,0))
draw.rectangle([440,180,460,500], fill=(255,140,0))
# رجل بدلة بيشاور
draw.ellipse([820,150,1180,680], fill=(25,25,25))
# نصوص بروفيشنال
try:
    font = ImageFont.truetype("/usr/share/fonts/truetype/noto/NotoNaskhArabic-Bold.ttf", 52)
    font2 = ImageFont.truetype("/usr/share/fonts/truetype/noto/NotoNaskhArabic-Bold.ttf", 38)
except:
    font = ImageFont.load_default()
    font2 = font
draw.text((30,20), "90% الجلوتين", fill=(255,255,255), font=font)
draw.text((30,90), "ممنوع!", fill=(255,255,255), font=font)
draw.text((30,170), "الارز مسموح", fill=(255,220,0), font=font)
draw.rectangle([25,245,330,325], fill=(170,0,0))
draw.text((40,260), "بقوة 100 حصان", fill=(255,255,0), font=font2)
draw.text((30,340), f"{name} {code}", fill=(0,255,200), font=font2)
draw.text((30,400), "15 دقيقة | ذروة", fill=(255,255,255), font=font2)
img.save("output/thumbnails/thumbnail_10000.jpg")
img.save("output/thumbnail_10000.jpg")
img.save("output/bg.jpg")
img.save(f"output/thumbnails/thumb_{code}.jpg")
print(f"✅ AI Thumb PRO {code} {name} - same context")
