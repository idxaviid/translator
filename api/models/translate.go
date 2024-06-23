package models

type InputTranslate struct {
	Text string `json:"text"`
	From string `json:"from"`
	To   string `json:"to"`
}
