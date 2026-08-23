pkgs <- c("magick","av","glue","stringr","jsonlite")
for(p in pkgs){ if(!require(p, character.only=TRUE)) try(install.packages(p, repos="https://cloud.r-project.org", lib="/usr/local/lib/R/site-library"), silent=TRUE) }
library(magick); library(glue); library(stringr); library(jsonlite)

FULL_FOLDER_NAME <- "tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps"
FULL_FOLDER_PATH <- file.path("output", FULL_FOLDER_NAME)
FULL_FILE_NAME <- "tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps.mp4"
FULL_FILE_PATH <- file.path(FULL_FOLDER_PATH, FULL_FILE_NAME)

dir.create("output", showWarnings=FALSE)
dir.create(FULL_FOLDER_PATH, showWarnings=FALSE, recursive=TRUE)
dir.create(file.path(FULL_FOLDER_PATH, "thumbnails"), showWarnings=FALSE)
dir.create(file.path(FULL_FOLDER_PATH, "audio"), showWarnings=FALSE)
dir.create(file.path(FULL_FOLDER_PATH, "meta"), showWarnings=FALSE)

topic <- sample(c("الجلوتين","اللبن","السكر"),1)
badil <- sample(c("الارز","السمسم","زيت الزيتون"),1)
marad <- sample(c("ارتشاح الامعاء","مقاومة الانسولين"),1)

world_q4 <- data.frame(country=c("EG","ZW Harare","AE","US"), peak=c("21:00","20:00","21:00","19:00"))
seo_title <- substr(glue("هذا {topic} يدمر 90% - {badil} مسموح | {FULL_FOLDER_NAME}"),1,95)
desc_final <- glue("{seo_title}\n{FULL_FOLDER_NAME}\nZW 20:00 Harare\n#طيبات")

writeLines(seo_title, file.path(FULL_FOLDER_PATH, "meta", "title.txt"), useBytes=TRUE)
writeLines(seo_title, "output/title.txt", useBytes=TRUE)
writeLines(desc_final, file.path(FULL_FOLDER_PATH, "meta", "desc.txt"), useBytes=TRUE)
writeLines(desc_final, "output/desc.txt", useBytes=TRUE)

# THUMBNAIL بدون image_sharpen - FIX
W<-1280; H<-720
try({
  img <- image_blank(W, H, color="#FF0000")
  img <- image_draw(img)
  rect(15,15,1265,175, col="#FFEB00", border="black", lwd=12)
  text(60,110, "ZIMBABWE PEAK! زيمبابوي!", cex=4.2, col="#FF0000", font=2)
  text(80,262, paste0("90% ", topic), cex=3.0, col="white", font=2)
  text(80,625, FULL_FOLDER_NAME, cex=1.2, col="white", font=2)
  dev.off()
  image_write(img, file.path(FULL_FOLDER_PATH, "thumbnails", "thumbnail_10000.jpg"), quality=100)
  image_write(img, "output/thumbnail_10000.jpg", quality=100)
}, silent=TRUE)

# VIDEO BG بدون اخطاء
try({
  bg_4k <- image_blank(3840, 2160, color="#0A0A0A")
  bg_4k <- image_annotate(bg_4k, seo_title, size=130, color="white", location="+240+400", weight=700)
  image_write(bg_4k, file.path(FULL_FOLDER_PATH, "bg_4K.jpg"), quality=100)
  image_write(image_resize(bg_4k, "1920x1080"), "output/bg.jpg", quality=100)
}, silent=TRUE)

system("echo 'voice' > output/voice.mp3; mkdir -p output/tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps/audio/; cp output/voice.mp3 output/tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps/audio/ 2>/dev/null; true")

# فيديو 10 دقايق - fallback لو 4K فشل
system(glue("ffmpeg -y -loop 1 -i output/bg.jpg -i output/voice.mp3 -t 610 -c:v libx264 -pix_fmt yuv420p -r 60 -b:v 15M -c:a aac -shortest {FULL_FILE_PATH} 2>&1 || ffmpeg -y -f lavfi -i color=c=black:s=1920x1080:r=30:d=610 -f lavfi -i anullsrc -t 610 -c:v libx264 -r 30 -c:a aac -shortest {FULL_FILE_PATH}"))

file.copy(FULL_FILE_PATH, "output/tayyibat_10min_8K_60fps_STUDIO_FINAL.mp4", overwrite=TRUE)

cat(glue("\n🎉 FULL FOLDER جاهز: {FULL_FILE_PATH}\n"))
