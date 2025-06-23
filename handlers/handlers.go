package handlers

import (
	"encoding/json"
	"fmt"
	"golang-game/database"
	"golang-game/models"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

// Response helpers
func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, status int) {
	http.Error(w, message, status)
}

func CharactersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getAllCharacters(w)
	case "POST":
		createCharacter(w, r)
	default:
		writeError(w, "No support method", http.StatusMethodNotAllowed)
	}
}

func getAllCharacters(w http.ResponseWriter) {
	rows, err := database.DB.Query("SELECT id, name, class, hp, mp, alive FROM characters")

	if err != nil {
		writeError(w, "Error finding characters", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var characters []models.Character
	for rows.Next() {
		var c models.Character
		if err := rows.Scan(&c.ID, &c.Name, &c.Class, &c.HP, &c.MP, &c.Alive); err != nil {
			continue
		}
		characters = append(characters, c)
	}

	writeJSON(w, characters, http.StatusOK)
}

func createCharacter(w http.ResponseWriter, r *http.Request) {
	var c models.Character
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Default values
	c.HP = 100
	c.MP = 50
	c.Alive = true

	result, err := database.DB.Exec("INSERT INTO characters(name, class, hp, mp, alive) VALUES(?, ?, ?, ?, ?)",
		c.Name, c.Class, c.HP, c.MP, c.Alive)
	if err != nil {
		writeError(w, "Erro ao inserir personagem", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	c.ID = int(id)

	writeJSON(w, c, http.StatusCreated)
}

func CharacterByIDHandler(w http.ResponseWriter, r *http.Request) {
	id, err := extractIDFromPath(r.URL.Path)
	if err != nil {
		writeError(w, "ID inválido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		getCharacterByID(w, id)
	case "DELETE":
		deleteCharacter(w, id)
	case "PATCH":
		handleCharacterAction(w, r, id)
	default:
		writeError(w, "No support method", http.StatusMethodNotAllowed)
	}
}

func extractIDFromPath(path string) (int, error) {
	idStr := strings.TrimPrefix(path, "/characters/")
	return strconv.Atoi(idStr)
}

func getCharacterByID(w http.ResponseWriter, id int) {
	var c models.Character
	err := database.DB.QueryRow("SELECT id, name, class, hp, mp, alive FROM characters WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Class, &c.HP, &c.MP, &c.Alive)

	if err != nil {
		writeError(w, "Character not found", http.StatusNotFound)
		return
	}

	writeJSON(w, c, http.StatusOK)
}

func deleteCharacter(w http.ResponseWriter, id int) {
	_, err := database.DB.Exec("DELETE FROM characters WHERE id = ?", id)
	if err != nil {
		writeError(w, "Error deleting character", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleCharacterAction(w http.ResponseWriter, r *http.Request, id int) {
	var req struct {
		Action string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	character, err := getCharacter(id)
	if err != nil || !character.Alive {
		writeError(w, "Character invalid or dead", http.StatusBadRequest)
		return
	}

	switch req.Action {
	case "attack":
		performAttack(w, character)
	case "heal":
		performHeal(w, character)
	case "meditate":
		performMeditate(w, character)
	default:
		writeError(w, "Invalid action", http.StatusBadRequest)
	}
}

func getCharacter(id int) (models.Character, error) {
	var c models.Character
	err := database.DB.QueryRow("SELECT id, name, class, hp, mp, alive FROM characters WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Class, &c.HP, &c.MP, &c.Alive)
	return c, err
}

func performAttack(w http.ResponseWriter, c models.Character) {
	if c.MP < 10 {
		writeError(w, "Insufficient MP", http.StatusBadRequest)
		return
	}

	targets := getAliveTargets(c.ID)
	if len(targets) == 0 {
		fmt.Fprintln(w, "No targets available")
		return
	}

	targetID := targets[rand.Intn(len(targets))]

	database.DB.Exec("UPDATE characters SET mp = mp - 10 WHERE id = ?", c.ID)
	database.DB.Exec("UPDATE characters SET hp = hp - 15 WHERE id = ?", targetID)
	database.DB.Exec("UPDATE characters SET alive = 0 WHERE id = ? AND hp <= 0", targetID)

	fmt.Fprintf(w, "%s hit character %d\n", c.Name, targetID)
}

func performHeal(w http.ResponseWriter, c models.Character) {
	if c.MP < 10 {
		writeError(w, "Insufficient MP", http.StatusBadRequest)
		return
	}

	newHP := c.HP + 20
	if newHP > 100 {
		newHP = 100
	}

	database.DB.Exec("UPDATE characters SET mp = mp - 10, hp = ? WHERE id = ?", newHP, c.ID)
	fmt.Fprintf(w, "%s heal yourself\n", c.Name)
}

func performMeditate(w http.ResponseWriter, c models.Character) {
	newMP := c.MP + 15
	if newMP > 50 {
		newMP = 50
	}

	database.DB.Exec("UPDATE characters SET mp = ? WHERE id = ?", newMP, c.ID)
	fmt.Fprintf(w, "%s meditate\n", c.Name)
}

func getAliveTargets(excludeID int) []int {
	rows, _ := database.DB.Query("SELECT id FROM characters WHERE alive = 1 AND id != ?", excludeID)
	defer rows.Close()

	var targets []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		targets = append(targets, id)
	}
	return targets
}

func BattleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeError(w, "No support method", http.StatusMethodNotAllowed)
		return
	}

	fighters := getAliveFighters()
	if len(fighters) < 2 {
		writeError(w, "No character available", http.StatusBadRequest)
		return
	}

	i, j := rand.Intn(len(fighters)), rand.Intn(len(fighters))
	for i == j {
		j = rand.Intn(len(fighters))
	}

	winner := rand.Intn(2)
	result := fmt.Sprintf("Battle between %s and %s!\n", fighters[i].Name, fighters[j].Name)

	if winner == 0 {
		result += fmt.Sprintf("%s win!\n", fighters[i].Name)
	} else {
		result += fmt.Sprintf("%s win!\n", fighters[j].Name)
	}

	fmt.Fprint(w, result)
}

type Fighter struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func getAliveFighters() []Fighter {
	rows, _ := database.DB.Query("SELECT id, name FROM characters WHERE alive = 1")
	defer rows.Close()

	var fighters []Fighter
	for rows.Next() {
		var f Fighter
		rows.Scan(&f.ID, &f.Name)
		fighters = append(fighters, f)
	}
	return fighters
}
