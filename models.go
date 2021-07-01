package main

type SignalMessage struct {
	Channel     string   `json:"kanaal"`
	Subscribers []string `json:"p2"`
	MessageBody string   `json:"meldingtekst"`
}
type Payload struct {
	Payload interface{} `json:"payload"`
}
