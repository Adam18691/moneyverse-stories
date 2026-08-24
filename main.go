package main

import (
	"context"
	"fmt"
	"os"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func main() {
	// 1. توليد بالذكاء المجاني
	titlePrompt := "اكتب عنوان جذاب لفيديو قرآن سورة البقرة مع ايموجي وهاشتاج الطيبات"
	title := GenerateTitle(titlePrompt)
	desc := GenerateTitle("اكتب وصف + 10 هاشتاجات لفيديو: " + title)

	fmt.Println("✅ Title:", title)

	// 2. رفع يوتيوب
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID: os.Getenv("YOUTUBE_CLIENT_ID"),
		ClientSecret: os.Getenv("YOUTUBE_CLIENT_SECRET"),
		Endpoint: oauth2.Endpoint{
			AuthURL: "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
		Scopes: []string{youtube.YoutubeUploadScope},
	}

	token := &oauth2.Token{
		RefreshToken: os.Getenv("YOUTUBE_REFRESH_TOKEN"),
	}
	client := conf.Client(ctx, token)

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err!= nil {
		fmt.Println("YouTube Error - اتأكد ان CLIENT_ID و SECRET من نفس المشروع المنشور:", err)
		return
	}

	// هنا حط كود الرفع بتاعك - مثال:
	fmt.Println("YouTube Ready:", service!= nil)
	fmt.Println("Description:", desc)
	//... كود الرفع الفعلي...
}
