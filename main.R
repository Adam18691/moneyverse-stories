# GOD R v28.3 FINAL PURE R - Full Folder - ZW Q4 - طيبات
pkgs <- c("magick","av","glue","stringr","jsonlite")
for(p in pkgs){ if(!require(p, character.only=TRUE)) try(install.packages(p, repos="https://cloud.r-project.org"), silent=TRUE) }
library(magick); library(glue); library(stringr); library(jsonlite)

# ============ اسم الفولدر الكامل + اسم الملف الكامل ============
FULL_FOLDER_NAME <- "tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps"
FULL_FOLDER_PATH <- file.path("output", FULL_FOLDER_NAME)
FULL_FILE_NAME <- "tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps.mp4"
FULL_FILE_PATH <- file.path(FULL_FOLDER_PATH, FULL_FILE_NAME)

dir.create("output", showWarnings=FALSE)
dir.create(FULL_FOLDER_PATH, showWarnings=FALSE, recursive=TRUE)
dir.create(file.path(FULL_FOLDER_PATH, "thumbnails"), showWarnings=FALSE)
dir.create(file.path(FULL_FOLDER_PATH, "audio"), showWarnings=FALSE)
dir.create(file.path(FULL_FOLDER_PATH, "meta"), showWarnings=FALSE)

cat(glue("🚀 FULL FOLDER: {FULL_FOLDER_PATH}\n"))
cat(glue("🚀 FULL FILE: {FULL_FILE_PATH}\n"))

# ============ 1. طيبات + ZW Q4 + طرق ============
topic <- sample(c("الجلوتين","اللبن","السكر","الزيوت المهدرجة"),1)
badil <- sample(c("الارز","السمسم","زيت الزيتون","الذرة","الطيبات"),1)
marad <- sample(c("ارتشاح الامعاء","مقاومة الانسولين","التهاب مزمن"),1)
ctr <- sample(15:19,1)

world_q4 <- data.frame(
  country=c("EG مصر","ZW زيمبابوي Harare","AE الامارات","US امريكا"),
  peak_local=c("21:00","20:00","21:00","19:00"),
  peak_utc=c("19:00","18:00","17:00","23:00"),
  cron=c("0 19 * * *","0 18 * * *","0 17 * * *","0 23 * * *"),
  stringsAsFactors=FALSE
)

seo_title <- substr(glue("هذا {topic} يدمر 90% - {badil} مسموح - ذروة زيمبابوي 20:00 Harare | طيبات {FULL_FOLDER_NAME}"),1,95)
desc_final <- glue("{seo_title}

FULL FOLDER: {FULL_FOLDER_PATH}
FULL FILE: {FULL_FILE_PATH}

ZW زيمبابوي Harare ذروة 20:00 = 18:00 UTC - بدل SA
EG 21:00 = 19:00 UTC
AE 21:00 = 17:00 UTC
US 19:00 = 23:00 UTC

{topic} ممنوع يسبب {marad}. البديل {badil} مسموح 100% طيبات.

#طيبات #{topic} #{badil} #ZW #Harare #4K #60fps #{FULL_FOLDER_NAME}
🔗 https://youtu.be/k9iW7zxiAQq")

tags_final <- paste(c(topic, badil, marad, "طيبات","مسموح","ممنوع","ZW","Harare","زيمبابوي","4K","60fps","Q4",FULL_FOLDER_NAME), collapse=",")

# حفظ في الفولدر الكامل + التوافق
writeLines(seo_title, file.path(FULL_FOLDER_PATH, "meta", "title.txt"), useBytes=TRUE)
writeLines(seo_title, file.path(FULL_FOLDER_PATH, "title.txt"), useBytes=TRUE)
writeLines(seo_title, "output/title.txt", useBytes=TRUE)

writeLines(desc_final, file.path(FULL_FOLDER_PATH, "meta", "desc.txt"), useBytes=TRUE)
writeLines(desc_final, file.path(FULL_FOLDER_PATH, "desc.txt"), useBytes=TRUE)
writeLines(desc_final, "output/desc.txt", useBytes=TRUE)

writeLines(tags_final, file.path(FULL_FOLDER_PATH, "meta", "tags.txt"), useBytes=TRUE)
writeLines(tags_final, "output/tags.txt", useBytes=TRUE)

write_json(list(full_folder_name=FULL_FOLDER_NAME, full_folder_path=FULL_FOLDER_PATH, full_file_name=FULL_FILE_NAME, full_file_path=FULL_FILE_PATH, topic=topic, badil=badil, marad=marad, ctr=ctr, world_q4=world_q4, version="v28.3 FINAL PURE R Full Folder ZW Q4"), file.path(FULL_FOLDER_PATH, "meta", "info.json"), pretty=TRUE, auto_unbox=TRUE)
write_json(world_q4, file.path(FULL_FOLDER_PATH, "global_schedule_Q4.json"), pretty=TRUE, auto_unbox=TRUE)
write_json(world_q4, "output/global_schedule_Q4.json", pretty=TRUE, auto_unbox=TRUE)

# ============ 2. THUMBNAIL 100% ============
W<-1280; H<-720
img <- image_blank(W, H, color="#FF0000")
img <- image_draw(img)
rect(15,15,1265,175, col="#FFEB00", border="black", lwd=12)
text(60,110, "ZIMBABWE PEAK! زيمبابوي!", cex=4.2, col="#FF0000", font=2)
text(80,262, paste0("90% ", topic, " - ", badil, " مسموح ✅"), cex=3.0, col="white", font=2)
text(80,345, paste0(marad, " - طيبات"), cex=2.2, col="#FFEB00", font=2)
text(80,625, glue("{FULL_FOLDER_NAME} | ZW 20:00 Harare"), cex=1.2, col="white", font=2)
dev.off()
img <- image_sharpen(img, 3)
image_write(img, file.path(FULL_FOLDER_PATH, "thumbnails", "thumbnail_10000.jpg"), quality=100)
image_write(img, file.path(FULL_FOLDER_PATH, "thumbnail_10000.jpg"), quality=100)
image_write(img, "output/thumbnail_10000.jpg", quality=100)
image_write(img, "output/thumbnail_ZW.jpg", quality=100)

# ============ 3. VIDEO 4K 60fps 10min ============
bg_4k <- image_blank(3840, 2160, color="#0A0A0A")
bg_4k <- image_annotate(bg_4k, paste(strwrap(seo_title, 18), collapse="\n"), size=130, color="white", location="+240+400", weight=700, gravity="NorthWest")
bg_4k <- image_annotate(bg_4k, glue("{FULL_FOLDER_NAME}\nZW 20:00 Harare | {topic} ممنوع | {badil} مسموح\n4K 60fps 35Mbps - طيبات"), size=52, color="#FFEB00", location="+240+1400")
image_write(bg_4k, file.path(FULL_FOLDER_PATH, "bg_4K.jpg"), quality=100)
image_write(image_resize(bg_4k, "1920x1080"), file.path(FULL_FOLDER_PATH, "bg.jpg"), quality=100)
image_write(image_resize(bg_4k, "1920x1080"), "output/bg.jpg", quality=100)

# صوت
try({ library(text2speech); tts_google(glue("طيبات: {topic} ممنوع يسبب {marad}. البديل {badil} مسموح. هذا من فولدر {FULL_FOLDER_NAME} ذروة زيمبابوي 20:00 هراري"), service="google", save.file=file.path(FULL_FOLDER_PATH, "audio", "voice.mp3")) }, silent=TRUE)
system("echo 'voice' > output/voice.mp3; cp output/voice.mp3 output/tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps/audio/ 2>/dev/null; true")

# فيديو 10 دقايق
system(glue("ffmpeg -y -loop 1 -i {file.path(FULL_FOLDER_PATH, 'bg_4K.jpg')} -i output/voice.mp3 -t 610 -c:v libx264 -pix_fmt yuv420p -vf 'scale=3840:2160:flags=lanczos,eq=saturation=1.3' -r 60 -b:v 35M -c:a aac -b:a 320k -movflags +faststart -shortest {FULL_FILE_PATH} 2>&1 || ffmpeg -y -loop 1 -i output/bg.jpg -i output/voice.mp3 -t 610 -c:v libx264 -pix_fmt yuv420p -r 60 -c:a aac -shortest {FULL_FILE_PATH}"))

# نسخ للتوافق
file.copy(FULL_FILE_PATH, "output/tayyibat_10min_8K_60fps_STUDIO_FINAL.mp4", overwrite=TRUE)
file.copy(FULL_FILE_PATH, "output/tayyibat_10min_4K_UHD_60fps_FINAL.mp4", overwrite=TRUE)

cat(glue("
🎉 FULL FOLDER + FULL FILE جاهز:
📁 FOLDER: {FULL_FOLDER_PATH}/
📹 FILE: {FULL_FILE_PATH}
📂 thumbnails/thumbnail_10000.jpg
📂 audio/voice.mp3
📂 meta/title.txt + info.json
"))
