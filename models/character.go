package models

type Character struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Class string `json:"class"`
    HP    int    `json:"hp"`
    MP    int    `json:"mp"`
    Alive bool   `json:"alive"`
}