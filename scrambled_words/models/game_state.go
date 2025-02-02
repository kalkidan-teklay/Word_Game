package models

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name" bson:"username"`

	Word  string `bson:"word"`
	Score int    `json:"score"`
}

type GameState struct {
	Word     string   `json:"word"`
	Shuffled string   `json:"shuffled"`
	Players  []Player `json:"players"`
	Started  bool     `json:"started"`
	Winner   *Player  `json:"winner"`
}
