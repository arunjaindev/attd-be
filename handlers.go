package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var users []User
	db.Find(&users)
	json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	params := mux.Vars(r)
	db.First(&user, params["userid"])
	if user.ID == 0 {
		json.NewEncoder(w).Encode("NO SUCH USER")
		return
	}
	json.NewEncoder(w).Encode(user)
}

func getUsersByRole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var users []User
	params := mux.Vars(r)
	db.Find(&users, "role = ?", params["role"])
	json.NewEncoder(w).Encode(users)
}

func addUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	json.NewDecoder(r.Body).Decode(&user)
	db.Create(&user)
	json.NewEncoder(w).Encode(user)
}

func delUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["userid"]

	var cnt int
	err := db.Raw("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&cnt).Error
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	if cnt == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User Not Found!"))
		return
	}

	err = db.Exec("DELETE FROM users WHERE id = ?", id).Error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error deleting user"))
		return
	}

	json.NewEncoder(w).Encode("User Deleted")
}

func punchIn(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    
    var user UserAttd
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    userID := user.UserId
    currentDate := time.Now().Format("2006-01-02")
    
    // Check if the user has punched in for the current date
    var existingUser UserAttd
    result := db.First(&existingUser, "user_id = ? AND date = ?", userID, currentDate)
    if result.Error != nil { // If no record found, create a new row
        result = db.Create(&user)
        if result.Error != nil {
            http.Error(w, "Unable to Punch In", http.StatusInternalServerError)
            return
        }
		json.NewEncoder(w).Encode("Punched In")
    } else { // If a record found, update the existing row
        if existingUser.Active {
            json.NewEncoder(w).Encode("Already Punched In")
            return
        } else {
            existingUser.Active = true
            // Update the existing user's activity status
            result = db.Save(&existingUser)
            if result.Error != nil {
                http.Error(w, "Unable to Punch In", http.StatusInternalServerError)
                return
            }
			json.NewEncoder(w).Encode("Punched In")
        }
    }
    
}

func punchOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    
    var user UserAttd
    err := json.NewDecoder(r.Body).Decode(&user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    userID := user.UserId
    currentDate := time.Now().Format("2006-01-02")
    
    // Check if the user has punched in for the current date
    var existingUser UserAttd
    result := db.First(&existingUser, "user_id = ? AND date = ?", userID, currentDate)
    if result.Error != nil { // If no record found, create a new row
		json.NewEncoder(w).Encode("You are Not Punched In")
    } else { // If a record found, update the existing row
        if !existingUser.Active {
            json.NewEncoder(w).Encode("You are Not Punched In")
            return
        } else {
            existingUser.Active = false
            // Update the existing user's activity status
            result = db.Save(&existingUser)
            if result.Error != nil {
                http.Error(w, "Unable to Punch Out", http.StatusInternalServerError)
                return
            }
			json.NewEncoder(w).Encode("Punched Out")
        }
    }
}

func changePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)

	var u User
	var user User
	json.NewDecoder(r.Body).Decode(&u)
	newPassword := u.Password // todo - hash password

	db.First(&user, "id = ?", params["userid"])
	db.Model(&user).Updates(map[string]interface{}{
		"Password": newPassword,
	})
	json.NewEncoder(w).Encode(user)
}

func getUserAttendance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	userID := params["userid"]
	month := params["month"]
	year := params["year"]

	var days []time.Time

	db.Model(&UserAttd{}).Where("user_id = ? AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?", userID, month, year).Pluck("date", &days)

	json.NewEncoder(w).Encode(days)
}

func getClassAttendance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	class := params["class"]
	month := params["month"]
	year := params["year"]

	m, _ := strconv.Atoi(month)
	y, _ := strconv.Atoi(year)

	startDate := time.Date(y, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)
	startDay := startDate.Day()
	endDay := endDate.Day()

	mp := make(map[int]int)

	for i := startDay; i <= endDay; i++ {
		currDate := i
		var cnt int
		db.Raw("SELECT COUNT(*) FROM (SELECT * FROM (SELECT * FROM users INNER JOIN user_attds ON users.id = user_attds.user_id) where EXTRACT(DAY FROM date) = ? AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ? AND class = ?)", currDate, month, year, class).Scan(&cnt)

		mp[currDate] = cnt
	}
	json.NewEncoder(w).Encode(mp)
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var expectedPassword string

	db.Raw("select password from users where email = ?", creds.Username).Scan(&expectedPassword)

	if expectedPassword != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(time.Minute * 5)

	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:    "token",
			Value:   tokenStr,
			Expires: expirationTime,
		})

	var u User
	db.Raw("select * from users where email = ?", creds.Username).Scan(&u)
	json.NewEncoder(w).Encode(u)
}
