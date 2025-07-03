package main

import (
	"log"
	"net/http"

	db "golang-game/database"
	handlers "golang-game/handlers"
)

func main() {

	db.InitDB()

	http.HandleFunc("/characters", handlers.CharactersHandler)
	http.HandleFunc("/characters/", handlers.CharacterByIDHandler)
	http.HandleFunc("/battle", handlers.BattleHandler)

	log.Println("Server listening in :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
