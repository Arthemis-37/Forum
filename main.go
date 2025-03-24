package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
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

func Register(username, surnom, email, password string) string, error {
	if !ValideEmail(email) {
		return "", errors.New("❌ Format d'email invalide")
	}
	
	existe, err := verifemail(DB, email)
	if err != nil {
		return "", err
	}
	if existe {
		return "", errors.New("❌ Cet email est déjà utilisé")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //hash les mdp
	if err != nil {
		return "", err
	}

	query := "INSERT INTO utilisateur (nom, surnom, email, MDP) VALUES (?, ?, ?, ?)" //query pour requête sql
	_, err = DB.Exec(query, username, surnom, email, string(hash))
	
	if err != nil {
		return "", errors.New("❌ Erreur lors de l'inscription")
	}
	
	userID, err := res.LastInsertId()
	if err != nil {
		return "", errors.New("❌ Erreur lors de la récupération de l'ID utilisateur")
	}

	// session creer avec inscrption info
	sessionID, err := createSession(int(userID), 24*time.Hour) //valide 24h
	if err != nil {
		return "", errors.New("❌ Erreur lors de la création de session")
	}

	fmt.Println("✅ Inscription réussie ! Session créée.")
	return sessionID, nil
}

func verifemail(db *sql.DB, email string) (bool, error) {
	var id int
	err := db.QueryRow("SELECT * FROM utilisateur WHERE email = ?", email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows){
		return false, err
	}
	return true, err
}

func ValideEmail(email string) bool {
	regex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)

	return re.MatchString(email)
}

func Login(email, password string) (string, error) {
	var userID int
	var hashpass string

	query := "SELECT id, MDP FROM utilisateur WHERE email = ?"
	err := DB.QueryRow(query, email).Scan(&userID, &hashpass)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("❌ Utilisateur non trouvé")
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		return "", errors.New("❌ Mot de passe incorrect")
	}

	// connection session après connexion réussie
	sessionID, err := createSession(userID, 24*time.Hour)
	if err != nil {
		return "", errors.New("❌ Erreur lors de la création de session")
	}

	fmt.Println("✅ Connexion réussie ! Session créée.")
	return sessionID, nil
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func createSession(userID int, expiration time.Duration) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", err
	}

	expirationTime := time.Now().Add(expiration) // Calcul de l'expiration

	query := "INSERT INTO sessions (id, user_id, expiration) VALUES (?, ?, ?)"
	_, err = DB.Exec(query, sessionID, userID, expirationTime)
	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func deleteSession(sessionID string) error {
	query := "DELETE FROM sessions WHERE id = ?"
	_, err := DB.Exec(query, sessionID)
	return err
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
	Date_crea  int
}

func getPost() ([]Post, error) {
	rows, err := DB.Query("SELECT ID, Titre, Nom_auteur, Catégorie, Contenu FROM post")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var infopost []Post

	for rows.Next() {
		var posts Post
		err := rows.Scan(&posts.ID, &posts.Titre, &posts.Nom_auteur, &posts.Catégorie, &posts.Contenu)
		if err != nil {
			return nil, err
		}
		infopost = append(infopost, posts)
	}
	return infopost, err
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		Titre := r.FormValue("Titre")
		Catégorie := r.FormValue("Catégorie")
		Contenu := r.FormValue("Contenu")

		post(Titre, "admin", Catégorie, Contenu)
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template : %v", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
	tmpl.Execute(w, r)
}

func main() {
	ConnectDB()
	Register("nom", "surnom", "email", "password")
	Login("email", "password")
	post("titre", "auteur", "categorie", "contenu")
	getPost()
	http.HandleFunc("/", IndexHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fmt.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
	verifemail(DB, "email")
}
