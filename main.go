package main

import (
"context"
"fmt"
"math/rand"
"os"
"os/exec"
"time"
"golang.org/x/oauth2"
"google.golang.org/api/option"
"google.golang.org/api/youtube/v3"
)

type Country struct{ Code, Name, Lang, TZ, Flag string; Peak int }

func main(){
ctx:=context.Background()
rand.Seed(time.Now().UnixNano())
now:=time.Now()
fmt.Println("🔥 v48 4 FROM 16 COUNTRIES PEAK", now.Format("2006-01-02"))

countries := []Country{
{"US","الولايات المتحدة","en","America/New_York",20,"🇺🇸"},
{"AU","أستراليا","en","Australia/Sydney",20,"🇦🇺"},
{"CH","سويسرا","de","Europe/Zurich",19,"🇨🇭"},
{"NO","النرويج","no","Europe/Oslo",19,"🇳🇴"},
{"NZ","نيوزيلندا","en","Pacific/Auckland",20,"🇳🇿"},
{"CA","كندا","en","America/Toronto",20,"🇨🇦"},
{"DE","ألمانيا","de","Europe/Berlin",19,"🇩🇪"},
{"DK","الدنمارك","da","Europe/Copenhagen",19,"🇩🇰"},
{"GB","المملكة المتحدة","en","Europe/London",19,"🇬🇧"},
{"NL","هولندا","nl","Europe/Amsterdam",19,"🇳🇱"},
{"FI","فنلندا","fi","Europe/Helsinki",19,"🇫🇮"},
{"SE","السويد","sv","Europe/Stockholm",19,"🇸🇪"},
{"AT","النمسا","de","Europe/Vienna",19,"🇦🇹"},
{"BE","بلجيكا","nl","Europe/Brussels",19,"🇧🇪"},
{"FR","فرنسا","fr","Europe/Paris",20,"🇫🇷"},
{"EG","مصر","ar","Africa/Cairo",20,"🇪🇬"},
}

os.MkdirAll("output/thumbnails",0755); os.MkdirAll("output/audio",0755); os.MkdirAll("output/meta",0755)
rand.Shuffle(len(countries), func(i,j int){countries[i],countries[j]=countries[j],countries[i]})
selected := countries[:4]
fmt.Printf("🌍 4 من 16 دولة اليوم: %s %s %s %s\n", selected[0].Flag+" "+selected[0].Code, selected[1].Flag+" "+selected[1].Code, selected[2].Flag+" "+selected[2].Code, selected[3].Flag+" "+selected[3].Code)

for idx, c := range selected {
    loc,_ := time.LoadLocation(c.TZ)
    peakLocal := time.Now().In(loc).Format("15:04")
    title := fmt.Sprintf("%s [%s] 90%% الجلوتين ممنوع! الارز مسموح بقوة 100 حصان | %s ذروة %d:00 %s", c.Flag, c.Code, c.Name, c.Peak, peakLocal)
    desc := fmt.Sprintf(`%s %s %s - ذروة المشاهدة %d:00 توقيت %s - %s
Tayyibat 15min - 90%% Gluten Free Rice Allowed 100HP
الفصول:
00:00 %s
02:30 لماذا الجلوتين ممنوع؟
06:00 الارز مسموح
10:00 سر 100 حصان
13:00 الخلاصة

#طيبات #%s #Peak%d #GlutenFree`, c.Flag, c.Name, c.Code, c.Peak, c.TZ, peakLocal, c.Name, c.Code, c.Peak)

    os.WriteFile(fmt.Sprintf("output/title_%s.txt", c.Code), []byte(title), 0644)
    script := fmt.Sprintf("اهلا %s %s %s ذروة %d - تسعين في المئة ممنوع عنهم الجلوتين والارز مسموح بقوة مئة حصان طيبات", c.Flag, c.Name, c.Code, c.Peak)
    os.WriteFile(fmt.Sprintf("output/audio/script_%s.txt", c.Code), []byte(script), 0644)

    // فيديو وصوت وصورة
    exec.Command("bash","-c", fmt.Sprintf("python3 generate_thumb.py; espeak -v %s -s 125 -w output/audio/voice_%s.wav -f output/audio/script_%s.txt || espeak -v ar -w output/audio/voice_%s.wav -f output/audio/script_%s.txt; ffmpeg -y -stream_loop 35 -i output/audio/voice_%s.wav -t 900 output/audio/final_%s.wav -y || ffmpeg -y -f lavfi -i anullsrc=r=44100:cl=stereo -t 900 output/audio/final_%s.wav -y; ffmpeg -y -loop 1 -r 30 -i output/thumbnails/thumbnail_10000.jpg -i output/audio/final_%s.wav -t 900 -c:v libx264 -c:a aac -pix_fmt yuv420p -shortest output/final_%s_%s_%d_15min.mp4 -y", c.Lang, c.Code, c.Code, c.Code, c.Code, c.Code, c.Code, c.Code, c.Code, c.Code, idx+1)).Run()

    // رفع يوتيوب ذروة
    if os.Getenv("YOUTUBE_CLIENT_ID")!=""{
        conf:=&oauth2.Config{ClientID:os.Getenv("YOUTUBE_CLIENT_ID"), ClientSecret:os.Getenv("YOUTUBE_CLIENT_SECRET"), Endpoint:oauth2.Endpoint{AuthURL:"https://accounts.google.com/o/oauth2/auth", TokenURL:"https://oauth2.googleapis.com/token"}, Scopes:[]string{youtube.YoutubeUploadScope}}
        tok:=&oauth2.Token{RefreshToken:os.Getenv("YOUTUBE_REFRESH_TOKEN")}
        client:=conf.Client(ctx, tok)
        ytService,_:=youtube.NewService(ctx, option.WithHTTPClient(client))
        file,_:=os.Open(fmt.Sprintf("output/final_%s_%s_%d_15min.mp4", c.Flag, c.Code, idx+1))
        v:=&youtube.Video{Snippet:&youtube.VideoSnippet{Title:title, Description:desc, CategoryId:"22", Tags:[]string{"طيبات",c.Code,c.Name,"Peak"}}, Status:&youtube.VideoStatus{PrivacyStatus:"public"}}
        if res,err:=ytService.Videos.Insert([]string{"snippet","status"}, v).Media(file).Do(); err==nil{
            fmt.Printf("✅ %s %s UPLOADED https://youtu.be/%s Peak %d %s %s\n", c.Flag, c.Code, res.Id, c.Peak, c.TZ, peakLocal)
        }
    }
}
os.WriteFile("output/upload_result.json", []byte(fmt.Sprintf(`{"16_countries":true,"selected":"%s %s %s %s","peak":true}`, selected[0].Code, selected[1].Code, selected[2].Code, selected[3].Code)), 0644)
fmt.Println("✅ DONE 4 FROM 16 PEAK")
}
