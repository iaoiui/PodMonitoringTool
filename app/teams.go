package main

import (
	"fmt"
	log "log"
	"net/http"
	"strings"
)

func sendAlertToTeams(title, msg, endpoint string) {
	b := fmt.Sprintf(`{ "title": "%v", "text": "%v"}`, title, msg)
	body := strings.NewReader(b)
	req, err := http.NewRequest("POST", endpoint, body)
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
}
