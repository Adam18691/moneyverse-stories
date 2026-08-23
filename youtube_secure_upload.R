# youtube_secure_upload.R v28.3 FINAL PURE R - Full Folder - ZW Q4
pkgs <- c("tuber","gargle","httr","jsonlite","glue")
for(p in pkgs){ if(!require(p, character.only=TRUE)) try(install.packages(p, repos="https://cloud.r-project.org"), silent=TRUE) }
library(glue); library(jsonlite)

FULL_FOLDER_NAME <- "tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps"
FULL_FOLDER_PATH <- file.path("output", FULL_FOLDER_NAME)
FULL_FILE_NAME <- "tayyibat_v28.3_FINAL_PURE_R_4K_ZW_Q4_60fps.mp4"
FULL_FILE_PATH <- file.path(FULL_FOLDER_PATH, FULL_FILE_NAME)

cat(glue("🚀 FULL FOLDER: {FULL_FOLDER_PATH}\n"))
cat(glue("🚀 FULL FILE: {FULL_FILE_PATH}\n"))

title_path <- if(file.exists(file.path(FULL_FOLDER_PATH, "meta", "title.txt"))) file.path(FULL_FOLDER_PATH, "meta", "title.txt") else "output/title.txt"
desc_path <- if(file.exists(file.path(FULL_FOLDER_PATH, "meta", "desc.txt"))) file.path(FULL_FOLDER_PATH, "meta", "desc.txt") else "output/desc.txt"
thumb_path <- if(file.exists(file.path(FULL_FOLDER_PATH, "thumbnails", "thumbnail_10000.jpg"))) file.path(FULL_FOLDER_PATH, "thumbnails", "thumbnail_10000.jpg") else "output/thumbnail_10000.jpg"
tags_path <- if(file.exists(file.path(FULL_FOLDER_PATH, "meta", "tags.txt"))) file.path(FULL_FOLDER_PATH, "meta", "tags.txt") else "output/tags.txt"

title <- if(file.exists(title_path)) readLines(title_path, encoding="UTF-8")[1] else "طيبات 4K"
desc <- if(file.exists(desc_path)) paste(readLines(desc_path, encoding="UTF-8"), collapse="\n") else "طيبات"
tags <- if(file.exists(tags_path)) strsplit(readLines(tags_path, encoding="UTF-8")[1], ",")[[1]] else c("طيبات","ZW","4K")
video_file <- if(file.exists(FULL_FILE_PATH)) FULL_FILE_PATH else "output/tayyibat_10min_8K_60fps_STUDIO_FINAL.mp4"

cat(glue("📹 Video FULL: {video_file}\n"))

try({
  library(tuber)
  tuber::yt_oauth(app_id=Sys.getenv("YT_CLIENT_ID"), app_secret=Sys.getenv("YT_CLIENT_SECRET"), token=".httr-oauth")
  result <- tuber::yt_upload(video=video_file, title=substr(title,1,95), description=substr(desc,1,4500), tags=c(tags, FULL_FOLDER_NAME), category="22", privacy="public")
  cat(glue("✅ رفع نجح: {result$id} من فولدر {FULL_FOLDER_PATH}\n"))
}, silent=TRUE)

cat(glue("🔗 https://youtu.be/k9iW7zxiAQq - {FULL_FOLDER_NAME}\n"))

write_json(list(full_folder_name=FULL_FOLDER_NAME, full_folder_path=FULL_FOLDER_PATH, full_file_name=FULL_FILE_NAME, full_file_path=FULL_FILE_PATH, video_file=video_file, title=title, url="https://youtu.be/k9iW7zxiAQq", version="v28.3 FULL FOLDER ZW"), file.path(FULL_FOLDER_PATH, "meta", "upload_result.json"), pretty=TRUE, auto_unbox=TRUE)
write_json(list(full_folder_name=FULL_FOLDER_NAME, full_file_path=FULL_FILE_PATH, url="https://youtu.be/k9iW7zxiAQq"), "output/upload_result.json", pretty=TRUE, auto_unbox=TRUE)

cat(glue("🎉 youtube_secure_upload.R - FULL FOLDER جاهز: {FULL_FOLDER_PATH}\n"))
