# Projet Forum

Bienvenue sur **Projet Forum** ! 🚀 Ce projet vise à créer un espace d'échange en ligne où les utilisateurs peuvent discuter, partager des idées et interagir autour de sujets variés.

## Fonctionnalités

- 📖 **Lecture** : Tout le monde peut parcourir les sujets et les posts.
- 🔐 **Authentification** : Inscription et connexion sécurisées avec mots de passe hashés.
- 📝 **Publication** : Les utilisateurs connectés peuvent créer des sujets et publier des posts.
- 👍👎 **Interactions** : Like, dislike et commentaires sur les posts.
- 🎯 **Filtrage** : Possibilité de trier les sujets par catégorie et par interactions de l'utilisateur.
- ⏳ **Sessions** : Gestion via un cookie avec expiration pour une connexion sécurisée.

## Contraintes techniques

- 🖥️ **Serveur web** : Développé en Golang.
- 🗄️ **Base de données** : SQLite, administrée via SQLite3.
- 🌍 **Navigation** : Une URL unique par page.

## Packages utilisés

- 🔑 **bcrypt** : Hashage sécurisé des mots de passe.
- 🗃️ **sqlite3** : Gestion de la base de données.
- 🆔 **uuid** : Gestion des sessions utilisateur via cookies.

## Installation et exécution

Prêt à tester le forum ? Suivez ces étapes simples :

1. **Cloner le dépôt** :
   ```sh
   git clone https://github.com/ton-utilisateur/Projet_Forum.git
   cd Projet_Forum
   ```
2. **Installer les dépendances** :
   ```sh
   go mod tidy
   ```
3. **Lancer le serveur** :
   ```sh
   go run main.go
   ```

🚀 Amusez-vous bien et bon développement !

