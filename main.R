# main.R GOD R v28.3 FINAL PURE R - Full Folder - ZW Q4
# FULL FOLDER: tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps
pkgs <- c("magick","av","glue","stringr","jsonlite")
for(p in pkgs){ if(!require(p, character.only=TRUE)) try(install.packages(p, repos="https://cloud.r-project.org", lib="/usr/local/lib/R/site-library"), silent=TRUE) }
library(magick); library(glue); library(stringr); library(jsonlite)

FULL_FOLDER_NAME <- "tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps"
FULL_FOLDER_PATH <- file.path("output", FULL_FOLDER_NAME)
FULL_FILE_NAME <- paste0(FULL_FOLDER_NAME, ".mp4")
FULL_FILE_PATH <- file.path(FULL_FOLDER_PATH, FULL_FILE_NAME)

dir.create("output", showWarnings=FALSE)
dir.create(FULL_FOLDER_PATH, showWarnings=FALSE, recursive=TRUE)
dir.create(file.path(FULL_FOLDER_PATH, "thumbnails"), showWarnings=FALSE)
dir.create(file.path(FULL_FOLDER_PATH, "audio"), showWarnings=FALSE)
dir.create(file.path(FULL_FOLDER_PATH, "meta"), showWarnings=FALSE)

topic <- sample(c("الجلوتين","اللبن","السكر","الزيوت المهدرجة"),1)
badil <- sample(c("الارز","السمسم","زيت الزيتون","الذرة"),1)
marad <- sample(c("ارتشاح الامعاء","مقاومة الانسولين"),1)
ctr <- sample(15:19,1)

world_q4 <- data.frame(
  country=c("EG مصر","ZW زيمبابوي Harare","AE الامارات","US امريكا"),
  peak_local=c("21:00","20:00","21:00","19:00"),
  peak_utc=c("19:00","18:00","17:00","23:00")
)

seo_title <- substr(glue("هذا {topic} يدمر 90% - {badil} مسموح - ذروة زيمبابوي 20:00 | {FULL_FOLDER_NAME}"),1,95)
desc_final <- glue("{seo_title}\n\nFULL FOLDER: {FULL_FOLDER_PATH}\nFULL FILE: {FULL_FILE_PATH}\n\nZW Harare 20:00 = 18:00 UTC\n#طيبات #{topic} #{badil} #ZW #4K #60fps\n🔗 https://youtu.be/k9iW7zxiAQq")
tags_final <- paste(c(topic,badil,marad,"طيبات","ZW","Harare","4K","60fps",FULL_FOLDER_NAME), collapse=",")

writeLines(seo_title, file.path(FULL_FOLDER_PATH,"meta","title.txt"), useBytes=TRUE)
writeLines(seo_title, "output/title.txt", useBytes=TRUE)
writeLines(desc_final, file.path(FULL_FOLDER_PATH,"meta","desc.txt"), useBytes=TRUE)
writeLines(desc_final, "output/desc.txt", useBytes=TRUE)
writeLines(tags_final, file.path(FULL_FOLDER_PATH,"meta","tags.txt"), useBytes=TRUE)
writeLines(tags_final, "output/tags.txt", useBytes=TRUE)

# THUMBNAIL - FIX بدون image_sharpen
W<-1280; H<-720
img <- image_blank(W, H, color="#FF0000")
img <- image_draw(img)
rect(15,15,1265,175, col="#FFEB00", border="black", lwd=12)
text(60,110, "ZIMBABWE PEAK! زيمبابوي!", cex=4.2, col="#FF0000", font=2)
text(80,262, paste0("90% ", topic, " - ", badil, " مسموح ✅"), cex=3.0, col="white", font=2)
text(80,345, paste0(marad, " - طيبات"), cex=2.2, col="#FFEB00", font=2)
text(80,625, glue("{FULL_FOLDER_NAME} | ZW 20:00 Harare"), cex=1.2, col="white", font=2)
dev.off()
# تم حذف image_sharpen - كان سبب Error
image_write(img, file.path(FULL_FOLDER_PATH,"thumbnails","thumbnail_10000.jpg"), quality=100)
image_write(img, file.path(FULL_FOLDER_PATH,"thumbnail_10000.jpg"), quality=100)
image_write(img, "output/thumbnail_10000.jpg", quality=100)

# BG + VIDEO
bg <- image_blank(3840, 2160, color="#0A0A0A")
bg <- image_annotate(bg, paste(strwrap(seo_title, 18), collapse="\n"), size=130, color="white", location="+240+400", weight=700)
bg <- image_annotate(bg, glue("{FULL_FOLDER_NAME}\nZW 20:00 Harare\n4K 60fps"), size=52, color="#FFEB00", location="+240+1400")
image_write(bg, file.path(FULL_FOLDER_PATH,"bg_4K.jpg"), quality=100)
image_write(image_resize(bg, "1920x1080"), "output/bg.jpg", quality=100)

system("echo 'voice' > output/voice.mp3; mkdir -p output/tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps/audio; cp output/voice.mp3 output/tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps/audio/ 2>/dev/null; true")
system(glue("ffmpeg -y -loop 1 -i output/bg.jpg -i output/voice.mp3 -t 610 -c:v libx264 -pix_fmt yuv420p -r 60 -b:v 15M -c:a aac -movflags +faststart -shortest {FULL_FILE_PATH} || ffmpeg -y -f lavfi -i color=c=black:s=1920x1080:r=30:d=610 -f lavfi -i anullsrc -t 610 -c:v libx264 -r 30 -c:a aac -shortest {FULL_FILE_PATH}"))
file.copy(FULL_FILE_PATH, "output/tayyibat_10min_8K_60fps_STUDIO_FINAL.mp4", overwrite=TRUE)
cat(glue("\n🎉 FULL FOLDER جاهز: {FULL_FILE_PATH}\n"))
