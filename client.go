package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	gomail "gopkg.in/mail.v2"
)

const (
	WATCH_LATER_URL = "https://api.bilibili.com/x/v2/history/toview"
	VIDEOS_PER_MAIL = 5
	SENDER_EMAIL    = "stevenjxhc@gmail.com"
)

type video struct {
	Title     string `json:"title"`
	Bvid      string `json:"bvid"`
	Pic       string `json:"pic"`
	TimeAdded int64  `json:"add_at"`
}

type watchLaterData struct {
	Count int     `json:"count"`
	List  []video `json:"list"`
}

type watchLaterResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Ttl     int            `json:"ttl"`
	Data    watchLaterData `json:"data"`
}

func main() {
	lambda.Start(sendReminderEmail)
}

func sendReminderEmail() error {
	videos, err := fetchWatchLater()
	if err != nil {
		return err
	}

	if len(videos) > 0 {
		err = sendMail(videos)
		if err != nil {
			return err
		}
	}

	return nil
}

func fetchWatchLater() ([]video, error) {
	req, err := http.NewRequest("GET", WATCH_LATER_URL, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set cookie
	sessdataCookie := os.Getenv("SESSDATA")
	if sessdataCookie == "" {
		return nil, fmt.Errorf("SESSDATA environment variable not set")
	}
	req.Header.Set("Cookie", fmt.Sprintf("SESSDATA=%s", sessdataCookie))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var watchLater watchLaterResponse
	err = json.Unmarshal(responseBody, &watchLater)
	if err != nil {
		log.Fatal(err)
	}

	// Choose earliest videos to send
	videos := watchLater.Data.List
	if len(videos) > VIDEOS_PER_MAIL {
		videos = videos[len(videos)-VIDEOS_PER_MAIL:]
	}

	return videos, nil
}

func sendMail(videos []video) error {
	// Create a new message
	message := gomail.NewMessage()

	// Set email headers
	message.SetHeader("From", SENDER_EMAIL)
	message.SetHeader("To", "stevenjxhc@gmail.com")
	message.SetHeader("Subject", "Don't Let These Videos Collect Dust!")

	// Build body
	body := `
        <html>
            <body>
                <h1>Time to watch these!</h1>
				<div>`

	for _, video := range videos {
		body += fmt.Sprintf(`
				<div style="display: inline-flex">
					<a href="bilibili.com/video/%s">
                		<img src="%s" alt="Video thumbnail" height="80px" width="120px">
            		</a>
					<div>
						<h3 style="margin-left: 40px;">%s</h3>
						<p style="margin-left: 40px;">Added on %s</p>
					</div>
				</div>`, video.Bvid, video.Pic, video.Title, time.Unix(video.TimeAdded, 0).Format("2006-01-02"))

		body += "<br />"
	}

	body += `
				</div>
			</body>
		</html>`

	// Set email body
	message.SetBody("text/html", body)

	// Set up the SMTP dialer
	gmailPassword := os.Getenv("GMAIL_PASSWORD")
	if gmailPassword == "" {
		return fmt.Errorf("GMAIL_PASSWORD environment variable not set")
	}
	dialer := gomail.NewDialer("smtp.gmail.com", 587, "stevenjxhc@gmail.com", gmailPassword)

	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		return err
	} else {
		fmt.Println("Email sent successfully!")
		return nil
	}
}
