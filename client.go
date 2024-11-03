package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	gomail "gopkg.in/mail.v2"
)

const (
	WATCH_LATER_URL = "https://api.bilibili.com/x/v2/history/toview"
	VIDEOS_PER_MAIL = 5
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
	cfg, err := loadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	sendReminderEmail(cfg)
}

func sendReminderEmail(cfg config) {
	videos, err := fetchWatchLater(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if len(videos) > 0 {
		err = sendMail(videos, cfg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func fetchWatchLater(cfg config) ([]video, error) {
	req, err := http.NewRequest("GET", WATCH_LATER_URL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Cookie", fmt.Sprintf("SESSDATA=%s", cfg.Sessdata))

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

func sendMail(videos []video, cfg config) error {
	// Create a new message
	message := gomail.NewMessage()

	// Set email headers
	message.SetHeader("From", "stevenjxhc@gmail.com")
	message.SetHeader("To", "stevenjxhc@gmail.com")
	message.SetHeader("Subject", "Just testing")

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
					<h3 style="margin-left: 40px;">%s</h3>
				</div>`, video.Bvid, video.Pic, video.Title)

		body += "<br />"
	}

	body += `
				</div>
			</body>
		</html>`

	// Set email body
	message.SetBody("text/html", body)

	// Set up the SMTP dialer
	dialer := gomail.NewDialer("smtp.gmail.com", 587, "stevenjxhc@gmail.com", cfg.GmailPassword)

	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		return err
	} else {
		fmt.Println("Email sent successfully!")
		return nil
	}
}
