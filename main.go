package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"

	"github.com/gorilla/securecookie"
)

// User details

type user struct {
	Id        int
	FirstName string
	lastName  string
	userName  string
	password  string
}

var s user

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

var t *template.Template
var db *sql.DB

// Connecting to the database and initializing the variable to load the static files

func init() {
	t = template.Must(template.ParseGlob("static/*.html"))
	var err error
	db, err = sql.Open("mysql", "root:Stebin@333@tcp(localhost:3306)/Project")
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
}

// Login page serving

func indexpage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	s.userName = getUsername(r)
	if s.userName == "" {
		err := t.ExecuteTemplate(w, "login.html", "Please login")
		if err != nil {
			fmt.Fprint(w, err)
		}
	} else {
		http.Redirect(w, r, "/homepage", http.StatusFound)
	}
}

//Verifying the user

func loginhandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("email")
	password := r.FormValue("password")
	if check(username, password) {
		setSession(username, w)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		err := t.ExecuteTemplate(w, "login.html", "Invalid Credintials")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
	}
}


func check(username, password string) bool {
	row := db.QueryRow("SELECT * FROM Users WHERE Password = ? AND UserName=?", password, username)
	if err := row.Scan(&s.Id, &s.FirstName, &s.lastName, &s.userName, &s.password); err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		fmt.Println(err)
	}
	return true
}

//SignUp page serving

func signup(w http.ResponseWriter, r *http.Request) {
	s.userName = getUsername(r)
	if s.userName == "" {
		err := t.ExecuteTemplate(w, "signup.html", nil)
		if err != nil {
			fmt.Fprint(w, err)
		}
	} else {
		http.Redirect(w, r, "/homepage", http.StatusFound)
	}
}

//Adding the new user to the database

func signupHandler(w http.ResponseWriter, r *http.Request) {
	fName := r.FormValue("fName")
	lName := r.FormValue("lName")
	eMail := r.FormValue("email")
	password := r.FormValue("Password")
	confirmPassword := r.FormValue("confirmPassword")
	if fName == "" || lName == "" || eMail == "" || password == "" || confirmPassword == "" {
		err := t.ExecuteTemplate(w, "signup.html", "Fill the required fields")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
	} else if password != confirmPassword {
		err := t.ExecuteTemplate(w, "signup.html", "Password mismatch")
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
	} else {
		if check(eMail, password) {
			err := t.ExecuteTemplate(w, "signup.html", "User Already Exsists")
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
		} else {
			_, err := db.Exec("INSERT INTO Users (FirstName, LastName, UserName,Password) VALUES (?, ?, ?,?)", fName, lName, eMail, password)
			if err != nil {
				fmt.Println(err)
			}
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}
}

//Session is set once a user loged In

func setSession(username string, w http.ResponseWriter) {
	value := map[string]string{
		"name": username,
	}
	encoded, err := cookieHandler.Encode("session", value)
	if err == nil {
		cookie := http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(w, &cookie)
	}
}

//home page serving

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	if s.userName == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	err := t.ExecuteTemplate(w, "home.html", s)
	if err != nil {
		fmt.Fprint(w, err)
	}
}

//checking if a user is already loged in

func getUsername(r *http.Request) (userName string) {
	cookie, err := r.Cookie("session")
	if err == nil {
		Value := make(map[string]string)
		err = cookieHandler.Decode("session", cookie.Value, &Value)
		if err == nil {
			userName = Value["name"]
		}
	}
	return
}

//Cookies are cleared once a user logout

func logouthandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", http.StatusFound)
}
func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}


func main() {
	http.HandleFunc("/", indexpage)
	http.HandleFunc("/login-submit", loginhandler)
	http.HandleFunc("/signup-submit", signupHandler)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/homepage", homeHandler)
	http.HandleFunc("/logout", logouthandler)
	http.ListenAndServe(":2671", nil)
}
