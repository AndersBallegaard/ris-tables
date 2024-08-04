package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func GenericHTTPGet(url string) string {
	resp, err := http.Get(url)
	ErrorParserFatal(err)
	s, err := io.ReadAll(resp.Body)
	ErrorParserFatal(err)
	return string(s)
}

func ErrorParserFatal(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

func (c *RRCInfoAPIResp) UnmarshalJSON(body string) {
	json.Unmarshal([]byte(body), &c)
}
