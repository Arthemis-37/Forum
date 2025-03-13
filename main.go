package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

// liens entre bddd et go
func ConnectDB() {
	// Format: "user:password@tcp(host:port)/dbname"
	dsn := "root:@tcp(localhost:3306)/forum?parseTime=true"

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ Erreur de connexion à la base de données:", err)
	}

	DB.SetConnMaxLifetime(5 * time.Minute) // connecter pour 5 min
	DB.SetMaxOpenConns(10)                 // 10 personnes connecter en même temps
	DB.SetMaxIdleConns(5)                  // 5 connection innactive ouverte en même temps

	if err = DB.Ping(); err != nil {
		log.Fatal("❌ Impossible de pinger la base de données:", err)
	}

	fmt.Println("Connexion réussie à MySQL !")

}

func Register(username, surnom, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //hash les mdp
	if err != nil {
		return err
	}

	query := "INSERT INTO utilisateur (nom, surnom, email, MDP) VALUES (?, ?, ?, ?)" //query pour requête sql
	_, err = DB.Exec(query, username, surnom, email, string(hash))

	return err
}

func Login(email, password string) error {
	query := "SELECT password FROM users WHERE email = ?"
	var hashpass string
	err := DB.QueryRow(query, email).Scan(&hashpass)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("utilisateur non trouvé")
		}
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		return errors.New("mot de passe incorrect")
	}
	return nil
}

func post(titre, auteur, categorie, contenu string) error {
	query := "INSERT INTO post (titre, nom_auteur, catégorie, contenu) VALUES (?, ?, ?, ?)" //query pour requête sql
	_, err := DB.Exec(query, titre, auteur, categorie, contenu)
	return err
}

type Post struct {
	ID         int
	Titre      string
	Nom_auteur string
	Catégorie  string
	Contenu    string
}

func getPost() error {
	rows, err := DB.Query("SELECT ID, Titre, Nom_auteur, Catégorie, Contenu FROM post")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	var infopost []Post

	for rows.Next() {
		var posts Post
		err := rows.Scan(&posts.ID, &posts.Titre, &posts.Nom_auteur, &posts.Catégorie, &posts.Contenu)
		if err != nil {
			return err
		}
		infopost = append(infopost, posts)
	}
	fmt.Println(infopost)
	return nil
}

func main() {
	ConnectDB()
	Register("nom", "surnom", "email", "password")
	Login("email", "password")
	post("titre", "auteur", "categorie", "contenu")
	fmt.Println(getPost())
}
