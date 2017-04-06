package lost

import "time"

type Serials []Serial

type Serial struct {
	ID       int        `json:"id"`
	Name     string     `json:"name"`
	Episodes []*Episode `json:"episodes"`
}

type Episode struct {
	Season int       `json:"season"`
	Number int       `json:"number"`
	Name   string    `json:"name"`
	Date   time.Time `json:"date"`
	Link   string    `json:"link"`
}
