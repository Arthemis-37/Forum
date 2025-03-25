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
		log.Fatal("❌Erreur de connexion à la base de données:", err)
	}

	DB.SetConnMaxLifetime(5 * time.Minute) // connecter pour 5 min
	DB.SetMaxOpenConns(10)                 // 10 personnes connecter en même temps
	DB.SetMaxIdleConns(5)                  // 5 connection innactive ouverte en même temps

	if err = DB.Ping(); err != nil {
		log.Fatal("❌Impossible de pinger la base de données:", err)
	}

	fmt.Println("Connexion réussie à MySQL !")

}

// ---------------------------------------------------------------inscription + connection sécurisé----------------------------------------------------
func Register(username, surnom, email, password string) (string, error) {
	if !ValideEmail(email) {
		return "", errors.New("❌Format d'email invalide")
	}

	existe, err := verifemail(DB, email)
	if err != nil {
		return "", err
	}
	if existe {
		return "", errors.New("❌Cet email est déjà utilisé")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) //hash les mdp
	if err != nil {
		return "", err
	}

	query := "INSERT INTO utilisateur (nom, surnom, email, MDP) VALUES (?, ?, ?, ?)" //query pour requête sql
	res, err := DB.Exec(query, username, surnom, email, string(hash))

	if err != nil {
		return "", errors.New("❌Erreur lors de l'inscription")
	}

	userID, err := res.LastInsertId()
	if err != nil {
		return "", errors.New("❌Erreur lors de la récupération de l'ID utilisateur")
	}

	// session creer avec inscrption info
	sessionID, err := createSession(int(userID), 24*time.Hour) //valide 24h
	if err != nil {
		return "", errors.New("❌Erreur lors de la création de session")
	}

	fmt.Println("✅Inscription réussie ! Session créée.")
	return sessionID, nil
}

func verifemail(db *sql.DB, email string) (bool, error) {
	var id int
	err := db.QueryRow("SELECT * FROM utilisateur WHERE email = ?", email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
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
			return "", errors.New("❌Utilisateur non trouvé")
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
	if err != nil {
		return "", errors.New("❌Mot de passe incorrect")
	}

	// connection session après connexion réussie
	sessionID, err := createSession(userID, 24*time.Hour)
	if err != nil {
		return "", errors.New("❌Erreur lors de la création de session")
	}

	fmt.Println("✅Connexion réussie ! Session créée.")
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

// -----------------------------------------------fonction autour du post + filtrage + likes/dislikes----------------------------------------------------
type Post struct {
	id          int
	auteurid    int
	contenu     string
	picture     string
	dislikes    int
	datepost    time.Time
	categorieid int
}

func getPost() ([]Post, error) {
	rows, err := DB.Query("SELECT ID, auteurid, contenu, picture, datepost, categorieid FROM post")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var infopost []Post
	for rows.Next() {
		var posts Post
		err := rows.Scan(&posts.id, &posts.auteurid, &posts.contenu, &posts.picture, &posts.datepost, &posts.categorieid)
		if err != nil {
			return nil, err
		}
		infopost = append(infopost, posts)
	}
	return infopost, nil
}

func Getcategorypost(categorieID int) ([]Post, error) {
	rows, err := DB.Query("SELECT ID, auteurid, contenu, picture, dislikes, datepost, categorieid FROM post WHERE categorieid = ? ORDER BY datepost DESC", categorieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.id, &p.auteurid, &p.contenu, &p.picture, &p.dislikes, &p.datepost, &p.categorieid)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

type Comments struct {
	id          int
	auteurid    int
	postid      int
	contenu     string
	commentdate time.Time
}

func Getcomments(postID int) ([]Comments, error) {
	query := "SELECT id, auteurid, postid, contenu, commentdate FROM commentaires WHERE postid = ? ORDER BY datecomment ASC"
	rows, err := DB.Query(query, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []Comments
	for rows.Next() {
		var c Comments
		err := rows.Scan(&c.id, &c.postid, &c.auteurid, &c.contenu, &c.commentdate)
		if err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	return comments, nil
}

// -----------------------------------------------------utilisateur connecter------------------------------------------------------------------------------
func post(titre, auteur, categorie, contenu string) error {
	query := "INSERT INTO post (titre, nom_auteur, catégorie, contenu) VALUES (?, ?, ?, ?)" //query pour requête sql
	_, err := DB.Exec(query, titre, auteur, categorie, contenu)
	return err
}

// même choses qu'avec les categorie mais pour les utilisateurs
func Getuserposts(categorieID int) ([]Post, error) {
	rows, err := DB.Query("SELECT ID, auteurid, contenu, picture, dislikes, datepost, categorieid FROM post WHERE auteurid = ? ORDER BY datepost DESC", categorieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.id, &p.auteurid, &p.contenu, &p.picture, &p.dislikes, &p.datepost, &p.categorieid)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func Getuserlikes(userID int) ([]Post, error) {
	query := `SELECT p.ID, p.auteurid, p.contenu, p.picture, p.dislikes, p.datepost, p.categorieid FROM post p JOIN likes l ON p.ID = l.postid WHERE l.userid = ? ORDER BY l.likesdate DESC`
	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var p Post
		err := rows.Scan(&p.id, &p.auteurid, &p.contenu, &p.picture, &p.dislikes, &p.datepost, &p.categorieid)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}

func Adddislikes(postID int) error {
	query := "UPDATE post SET dislikes = dislikes + 1 WHERE ID = ?"
	_, err := DB.Exec(query, postID)
	if err != nil {
		return err
	}

	fmt.Println("👎Dislike ajouté au post", postID)
	return nil
}

func Addlikes(userID, postID int) error {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM likes WHERE userid = ? AND postid = ?)", userID, postID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("❌L'utilisateur a déjà liké ce post")
	}
	query := "INSERT INTO likes (userid, postid, likesdate) VALUES (?, ?, CURRENT_TIMESTAMP)"
	_, err = DB.Exec(query, userID, postID)
	if err != nil {
		return err
	}
	fmt.Printf("✅L'utilisateur %d a liké le post %d\n", userID, postID)
	return nil
}


// mettre un commentaire
//repondre a un commentaire
//verif email dans login
//---------------------------------------------------------------------hébergeur--------------------------------------------------------------------------

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
	deleteSession("sessionID")
	http.HandleFunc("/", IndexHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	fmt.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
