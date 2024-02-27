package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var err error

type User struct {
	gorm.Model
	Role      string `json:"role"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Class     int    `json:"class"`
	Password  string `gorm:"type:string;default:User@123" json:"password"`
}

type UserAttd struct {
	gorm.Model
	UserId  int       `json:"userid"` //refers id in user
	Date    time.Time `gorm:"type:date;default:current_date" json:"date"`
	TimeIn  time.Time `gorm:"default:current_timestamp" json:"timein"`
	TimeOut time.Time `gorm:"default:null" json:"timeout"`
	Active  bool      `gorm:"default:true" json:"active"`
}

var jwtKey = []byte(",secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func initialiseMigration() {
	dsn := "user=postgres password=admin dbname=userDB host=localhost port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	fmt.Println("DB Connected")

	db.AutoMigrate(&User{})
	db.AutoMigrate(&UserAttd{})

	db.Exec("ALTER TABLE user_attds ADD FOREIGN KEY (user_id) REFERENCES users(id);")
}

func initialiseRouter() {
	r := mux.NewRouter()

	//routes / endpoints
	r.HandleFunc("/api/users", getUsers).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/user/{userid}", getUser).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/users/{role}", getUsersByRole).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/addUser", addUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/delUser/{userid}", delUser).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/punchIn", punchIn).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/punchOut", punchOut).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/changePassword/{userid}", changePassword).Methods("PUT", "OPTIONS")
	r.HandleFunc("/api/userAttendance/{userid}/{month}/{year}", getUserAttendance).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/classAttendance/{class}/{month}/{year}", getClassAttendance).Methods("GET", "OPTIONS")

	r.HandleFunc("/api/login", login).Methods("POST", "OPTIONS")

	r.Use(mux.CORSMethodMiddleware(r))
	r.Use(CORSMiddleware)

	log.Fatal(http.ListenAndServe(":8000", r))
}

func main() {
	//initialise migration
	initialiseMigration()

	//initialise the router
	initialiseRouter()
}
