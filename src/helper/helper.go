// Author: Pirakalan

package helper

import (
	"log"
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"settings"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"time"
	"math/rand"
	"strconv"
)

var (
	db *sql.DB
)

func CreateUser(username string, password string, fullname string, question string, answer string) {
	// A password hash is created based on the pepper defined in settings
	passHash, err := bcrypt.GenerateFromPassword([]byte(password + settings.PEPPER), bcrypt.DefaultCost)
	checkErr("Bcrypt generate password hash error in CreateUser()", err)

	// Answer to the secret question is encrpted since the password can be reset with this
	answerHash, err := bcrypt.GenerateFromPassword([]byte(answer + settings.PEPPER), bcrypt.DefaultCost)
	checkErr("Bcrypt generate answer hash error CreateUser()", err)

	query, errDb := db.Prepare("INSERT INTO users (username, fullname, passwordhash, secretquestion, secretanswer) VALUES (?, ?, ?, ?, ?);")
	checkErr("Db error in CreateUser()", errDb)

	_, errDBexec := query.Exec(username, fullname, string(passHash), question, answerHash)
	checkErr("Db exec error in CreateUser()", errDBexec)

	query.Close()
}

// Password hash is updated in the 'users' based on the new password provided by the user
func ResetPassword(username string, password string) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password + settings.PEPPER), bcrypt.DefaultCost)
	checkErr("Bcrypt generate password hash error in ResetPassword()", err)

	query, errDb := db.Prepare("UPDATE users SET passwordhash = ? where username = ?;")
	checkErr("Db error in ResetPassword()", errDb)

	_, errDBexec := query.Exec(string(passHash), username)
	checkErr("Db exec error in ResetPassword()", errDBexec)

	query.Close()
}

// Verify the username is already registered by checking the database
func CheckUsernameExists(username string) bool {
	query, err := db.Query("SELECT username FROM users WHERE username = ?;", username)
	checkErr("Db query error in CheckUsernameExists()", err)
 
	usernameExists := false

	if query.Next() {
		usernameExists = true
	}
	query.Close()

	return usernameExists
}

// Verify the user entered the correct credentials
func IsValidUser(username string, input string, column string) bool {
	query, err := db.Query(fmt.Sprintf("SELECT %s FROM users WHERE username = '%s';", column, username))
	checkErr("Db query error in IsValidUser()", err)

	var hash string
	if query.Next() {
		err = query.Scan(&hash)
		checkErr("Query scan error in IsValidUser()", err)
	}

	query.Close()

	pass_error := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input + settings.PEPPER))

	if pass_error == nil {
		return true
	} else {
		return false
	}
}

// Retrieve the corresponding secret question based on the username
func GetSecretQuestion(username string) string {
	query, err := db.Query("SELECT secretquestion FROM users WHERE username = ?;", username)
	checkErr("Db query error in GetSecretQuestion()", err)

	var secretQuestion string

	for query.Next() {
		err = query.Scan(&secretQuestion)
		checkErr("Query scan error in GetSecretQuestion()", err)
	}

	query.Close()
	
	return secretQuestion
}

// Check to see if the logged in user has a valid session in their cookie
func IsValidSessionKey(username string, key string) bool {
	query, err := db.Query("SELECT sessionkey, logintime FROM usersSession WHERE username = ?;", username)
	checkErr("Db query error in IsValidSessionKey()", err)
	
	deleteSession := false
	validSession := false

	var sessionKey string
	var loginTime time.Time

	for query.Next() {
		err = query.Scan(&sessionKey, &loginTime)
		checkErr("Query scan error IsValidSessionKey()", err)

		// Allow a vaild session to be 6 hours long
		// There can be multiple sessions for the same user, therefore this loop will go
		// through the retrieved results from the database to check if the session key from
		// cookie is valid
		if time.Since(loginTime).Hours() <= 6 && key == sessionKey {
			validSession = true
		} else if key == sessionKey { // Session expired
			deleteSession = true
		}
	}

	query.Close()

	if deleteSession {
		DeleteSessionKey(sessionKey)
	}

	return validSession
}

// Delete the entry from 'userSession' table based on the session key
func DeleteSessionKey(key string) {
	queryDelete, errDb := db.Prepare("DELETE FROM usersSession where sessionkey = ?")
	checkErr("Db error in deleteSessionKey()", errDb)

	_, errDeleteExec := queryDelete.Exec(key)
	checkErr("Db errDeleteExec error deleteSessionKey()", errDeleteExec)

	queryDelete.Close()
}

// A new valid session key is generated and inserted into the table
func CreateSession(username string, password string) string {
	sessionKeyHash, err := bcrypt.GenerateFromPassword([]byte(username + settings.SESSIONPEPPER + password), bcrypt.DefaultCost)
	checkErr("Bcrypt generate answer hash error in CreateSession()", err)

	queryInsert, errDbInsert := db.Prepare("INSERT INTO usersSession (username, sessionkey) VALUES (?, ?);")
	checkErr("Db error in CreateSession()", errDbInsert)

	_, errDBInsertExec := queryInsert.Exec(username, string(sessionKeyHash))
	checkErr("Db exec error in CreateSession()", errDBInsertExec)

	queryInsert.Close()

	return string(sessionKeyHash)
}

// The goroutine that runs in the background calls this function every hour 
// to clear out old sessions (greater than 6 hours) from 'usersSession' table
func DeleteOldSessions() {
	// Delete sessions greater than 6 hours long
	queryDelete, errDbDelete := db.Prepare("DELETE FROM usersSession WHERE ((STRFTIME('%s','NOW') - STRFTIME('%s',logintime))/60/60) >= 6;")
	checkErr("Db error in DeleteOldSessions()", errDbDelete)

	_, errDbDeleteExec := queryDelete.Exec()
	checkErr("Db exec error in DeleteOldSessions()", errDbDeleteExec)

	queryDelete.Close()

	UpdateSessionHistory("", "Session cleaned up", true)
}

// The 'usersHistory' table is created to display user session history within the command line
// A new entry is inserted everytime a user logs in
func CreateSessionHistory(username string, key string) {
	queryInsert, errDbInsert := db.Prepare("INSERT INTO usersHistory (username, sessionkey, status) VALUES (?, ?, ?);")
	checkErr("Db error in CreateSessionHistory()", errDbInsert)

	_, errDBInsertExec := queryInsert.Exec(username, key, "Active user")
	checkErr("Db exec error in CreateSessionHistory()", errDBInsertExec)

	queryInsert.Close()
}

// Session history is updated when a user logs out or when old sessions are cleanup by the goroutine
func UpdateSessionHistory(key string, status string, oldSessions bool) {
	if oldSessions {
		queryUpdate, errDbUpdate := db.Prepare("UPDATE usersHistory SET statusupdatedat = current_timestamp, status = ? WHERE status = 'Active user' AND ((STRFTIME('%s','NOW') - STRFTIME('%s',logintime))/60/60) >= 6;")
		checkErr("Db error in UpdateSessionHistory()", errDbUpdate)

		_, errDBUpdateExec := queryUpdate.Exec(status)
		checkErr("Db exec error in UpdateSessionHistory()", errDBUpdateExec)
		queryUpdate.Close()
	} else {
		queryUpdate, errDbUpdate := db.Prepare("UPDATE usersHistory SET statusupdatedat = current_timestamp, status = ? WHERE sessionkey = ?;")
		checkErr("Db error in UpdateSessionHistory()", errDbUpdate)

		_, errDBUpdateExec := queryUpdate.Exec(status, key)
		checkErr("Db exec error in UpdateSessionHistory()", errDBUpdateExec)
		queryUpdate.Close()
	}
}

// Output the session history onto the command line by running: 'go run main.go --display=history'
func PrintSessionHistory(command string){
	if command == "history" {
		fmt.Printf("User Session History:\n\n Username | Session Key | Logged In Time | Status Updated At | Status\n")
		fmt.Println("____________________________________________________________________")
		query, err := db.Query("SELECT username, sessionkey, logintime, statusupdatedat, status FROM usersHistory;")
		checkErr("Db query error in PrintSessionHistory()", err)

		var username, sessionkey, status string
		var loginTime, statusUpdated time.Time
		for query.Next() {
			err = query.Scan(&username, &sessionkey, &loginTime, &statusUpdated, &status)
			fmt.Printf("%s | %s | %s | %s | %s\n\n", username, sessionkey, loginTime, statusUpdated, status)
			checkErr("Query scan error in PrintSessionHistory()", err)
		}

		query.Close()
	}
}

// The table 'city' was created by using the following JSON file and it contains all possible cities 
// that the API supports: http://bulk.openweathermap.org/sample/city.list.json.gz
// This function outputs JSON formatted string with region and id based on the search term.
// It is called when the HTTP handler (CityHandler) receives a request via: "/citylist.json?search=".
// The jQuery autocomplete library utlizes this to output region suggestions as the user enters value
func CityQuery(searchTerm string) string {
	query, err := db.Query("SELECT key, region FROM city WHERE LOWER(region) like '%%"+searchTerm+"%%' limit 5;")
	checkErr("Db query error in CityQuery()", err)

	var key int
	var region string
	var output string

	for query.Next() {
		err = query.Scan(&key, &region)
		checkErr("Query scan error in CityQuery()", err)
		output += fmt.Sprintf("{\"value\":%d,\"label\":\"%s\"},",key,region)
	}

	query.Close()
	return output
}

// "I’m feeling lucky button" feature utilizes this function to retrieve random city. For each user a random 
// number is initially chosen from a range of total number of regions from the 'city' table. Last shown 
// region key is stored in the 'luckyTracker' table based on the username. To avoid collision a double
// hashing function is used
func GetRandomCity(username string) string {
	query, err := db.Query("SELECT lastshownkey FROM luckyTracker WHERE username = ?;", username)
	checkErr("Db query error in GetRandomCity()", err)

	var lastshownkey int
	found := false

	for query.Next() {
		err = query.Scan(&lastshownkey)
		found = true
		checkErr("Query scan error in GetRandomCity()", err)
	}
	query.Close()

	var r int

	if !found {
		// Assign a new starting point for each user using "I’m feeling lucky button" feature for the first time
		rand.Seed(time.Now().Unix())
		r = rand.Intn(209490-1) + 1 // Adding 1 as there is no city entry at 0 (referring to 'city' table)
	} else {
		r = doubleHashing(lastshownkey)
	}

	if !found {
		// Insert into 'luckyTracker' table if there is no entries
		queryInsert, errDbInsert := db.Prepare("INSERT INTO luckyTracker (username, lastshownkey) VALUES (?, ?);")
		checkErr("Db error in GetRandomCity()", errDbInsert)

		_, errDBInsertExec := queryInsert.Exec(username, r)
		checkErr("Db exec error in GetRandomCity()", errDBInsertExec)

		queryInsert.Close()
	} else {
		// Update the last shown city index into the table
		queryUpdate, errDbUpdate := db.Prepare("UPDATE luckyTracker SET lastshownkey = ? WHERE username = ?;")
		checkErr("Db error in GetRandomCity()", errDbUpdate)

		_, errDBUpdateExec := queryUpdate.Exec(r, username)
		checkErr("Db exec error in GetRandomCity()", errDBUpdateExec)
		queryUpdate.Close()
	}

	// Get City Key based on r
	query, err = db.Query("SELECT key FROM city WHERE id = ?;", r)
	checkErr("Db query error in GetRandomCity()", err)

	var cityKey int

	for query.Next() {
		err = query.Scan(&cityKey)
		checkErr("Query scan error in GetRandomCity()", err)
	}
	query.Close()

	return strconv.Itoa(cityKey)
}

// Utilize double hashing to avoid collision which will result in showing unique cities to user using 
// 'Feeling Lucky' feature
func doubleHashing(r int) int {
	// 7 was chosen arbitrarily and also there is no common factor with total number of cities (209497) 
	// which is relatively prime
	decrement := 7
	cityTableSize := 209497

	r = r - decrement

	// To correct the index to wrap around the table
	if r <= 0 {
		r = r + cityTableSize
	}

	return r
}

func CloseDB() {
	db.Close()
}

func checkErr(message string, err error) {
	if err != nil {
		log.Printf("%s> %s", message, err.Error())
	}
}

func init(){
	var err error
	db, err = sql.Open("sqlite3", "userdb.sqlite?cache=shared&mode=rwc")
	checkErr("Sqlite3 open error", err)
}