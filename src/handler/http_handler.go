// Author: Pirakalan

package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"html/template"
	"time"
	"log"
	"definition"
	"settings"
	"helper"
	"session"
	"strings"
)

// This struct is used to interact with the HTML while rendering 
type HtmlResponse struct{
	Result string
	Action string
	Username string
	UserStatus string
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	var val = HtmlResponse{}
	if isValidSession, _ := session.VerifySession(w, r); isValidSession {
		val.UserStatus = "loggedin" // To show the log in or log out icon in the HTML
	}

	t, _ := template.ParseFiles("templates/hello.html")
	t.Execute(w, val)
}

// Displays 'Hello World!' text
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

// This handler will respond based on the city search query with an array of JSON as result
// The jQuery autocomplete plugin uses the result to show suggestions within the search box
func CityHandler(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("search")
	if len(searchQuery) > 0 {
		output := helper.CityQuery(strings.ToLower(searchQuery))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("["+strings.TrimSuffix(output, ",")+"]"))
	}
}

// Handle requests to display weather data based on the cities searched by the user or random 
// unique cities when "I’m feeling lucky button" is clicked
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var val = HtmlResponse{}

	// Check to see if the session is vaild
	if isValidSession, _ := session.VerifySession(w, r); isValidSession {
		val.UserStatus = "loggedin"
		if r.Method == "GET" {
			t, _ := template.ParseFiles("templates/search.html")
			t.Execute(w, val)
		} else if r.Method == "POST" {
			r.ParseForm()

			var luckyCityId string

			if r.Form["type"][0] == "feelinglucky" {
				username, _ := session.ReadCookieHandler(w, r)
				// Retrieves random unique cities by utilizing double hashing function to resolve collisions
				luckyCityId = helper.GetRandomCity(username)
			}

			// Depending on whether the user wants to search by using the autocompleted city, custom city search
			// or I'm Feeling Lucky feature the appropriate request is made to OpenWeatherMap API
			var query string
			if luckyCityId != "" {
				query = fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?id=%s&type=accurate&units=metric&mode=json&APPID=%s", luckyCityId, settings.APIKEY)
			} else if r.Form["cityautocomplete"][0] != "" {
				query = fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?id=%s&type=accurate&units=metric&mode=json&APPID=%s", r.Form["cityautocomplete"][0], settings.APIKEY)
			} else {
				query = fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&type=accurate&units=metric&mode=json&APPID=%s", r.Form["city"][0], settings.APIKEY)
			}

			apiResult := getCurrentWeather(query)
			
			apiResult.UserStatus = val.UserStatus

			// If there is an error code resulted based on the query then the appropriate message is displayed
			if apiResult.Code == 400 || apiResult.Code == 404 {
				var val = HtmlResponse{}
		
				val.Result = fmt.Sprintf("Please try again. Error message: '%s'", apiResult.Message)
				t, _ := template.ParseFiles("templates/result.html")
				t.Execute(w, val)
			} else {
				t, _ := template.ParseFiles("templates/weather_results.html")
				t.Execute(w, apiResult)
			}
		}
	} else {
		// New users or users with expired sessions are asked to log in
		http.Redirect(w, r, "/login", 302)
	}
}

// Creates new users based on the registration form
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var val = HtmlResponse{}
	if isValidSession, _ := session.VerifySession(w, r); isValidSession {
		val.UserStatus = "loggedin"
	}
	if r.Method == "GET" {
		t, _ := template.ParseFiles("templates/create_user.html")
		t.Execute(w, val)
	} else if r.Method == "POST" {
		r.ParseForm()
		if helper.CheckUsernameExists(r.Form["username"][0]) {
			val.Result = fmt.Sprintf("The username: '%s' is already registered, please try again.", r.Form["username"][0])
			t, _ := template.ParseFiles("templates/result.html")
			t.Execute(w, val)
		} else {
			helper.CreateUser(r.Form["username"][0], r.Form["password"][0], r.Form["fullname"][0], r.Form["question"][0], r.Form["answer"][0])
			
			val.Result = "User created successfully!"
			t, _ := template.ParseFiles("templates/result.html")
			t.Execute(w, val)
		}
	}
}

// Allows users to log in, if authenticated successfully then assigns a new session cookie
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var val = HtmlResponse{}
	isValidSession, message := session.VerifySession(w, r)

	// Users who have a valid session is not allow to log in again
	if isValidSession {
		val.UserStatus = "loggedin"
		val.Result = "Already logged in"
		t, _ := template.ParseFiles("templates/result.html")
		t.Execute(w, val)
	} else {
		if r.Method == "GET" {
			val.Result = message
			t, _ := template.ParseFiles("templates/login.html")
			t.Execute(w, val)
		} else if r.Method == "POST" {
			r.ParseForm()

			if helper.IsValidUser(r.Form["username"][0], r.Form["password"][0], "passwordhash") {
				val.UserStatus = "loggedin"
				session.SetSession(w, r, r.Form["username"][0], r.Form["password"][0])
				val.Result = fmt.Sprintf("Hello %s!\n", r.Form["username"][0])
				t, _ := template.ParseFiles("templates/result.html")
				t.Execute(w, val)
			} else{
				val.Result = "Incorrect credentials!"
				t, _ := template.ParseFiles("templates/login.html")
				t.Execute(w, val)
			}
		}
	}
}

// Authenticated users will be able to log out and this is done by removing user's session key
// from the 'usersSession' table
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var val = HtmlResponse{}
	if isValidSession, _ := session.VerifySession(w, r); isValidSession {
		session.ClearSession(w, r)
		val.Result = "Logged out successfully!"
		t, _ := template.ParseFiles("templates/result.html")
		t.Execute(w, val)
	} else {
		val.Result = "Not logged in"
		t, _ := template.ParseFiles("templates/result.html")
		t.Execute(w, val)
	}
}

// User can reset their password by answering their secret question
func PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	var val = HtmlResponse{}
	if isValidSession, _ := session.VerifySession(w, r); isValidSession {
		val.UserStatus = "loggedin"
		val.Result = "Already logged in, please log out to reset password."
		t, _ := template.ParseFiles("templates/result.html")
		t.Execute(w, val)
	} else {
		if r.Method == "GET" {
			// The user is asked for their username first
			val.Action = "username"
			t, err := template.ParseFiles("templates/password_reset.html")
			checkErr("Template parsefile error", err)
			t.Execute(w, val)
		} else if r.Method == "POST" {
			r.ParseForm()
			if r.Form["result"][0] == "usernamesubmitted" {
				if !helper.CheckUsernameExists(r.Form["val"][0]) {
					val.Action = "message"
					val.Result = fmt.Sprintf("The username: '%s' is not registered", r.Form["val"][0])
					t, err := template.ParseFiles("templates/password_reset.html")
					checkErr("Template parsefile error", err)
					t.Execute(w, val)
				} else {
					// Once their username is verified, their secret question is displayed
					val.Action = "question"
					val.Result = helper.GetSecretQuestion(r.Form["val"][0])
					val.Username = r.Form["val"][0]
					t, err := template.ParseFiles("templates/password_reset.html")
					checkErr("Template parsefile error", err)
					t.Execute(w, val)
				}
			} else if r.Form["result"][0] == "answersubmitted" {
				if helper.IsValidUser(r.Form["username"][0], r.Form["val"][0], "secretanswer") {
					// Once the answer to secret question is validated, the user can enter the new password
					val.Action = "resetpass"
					val.Username = r.Form["username"][0]
					t, err := template.ParseFiles("templates/password_reset.html")
					checkErr("Template parsefile error", err)
					t.Execute(w, val)
				} else {
					val.Action = "message"
					val.Result = "The answer to the secret question is incorrect."
					t, err := template.ParseFiles("templates/password_reset.html")
					checkErr("Template parsefile error", err)
					t.Execute(w, val)
				}
			} else if r.Form["result"][0] == "passwordsubmitted" {
				// The 'users' table is updated with the new password hash
				helper.ResetPassword(r.Form["username"][0], r.Form["password"][0])
				val.Action = "message"
				val.Result = "Password is reset!"
				t, err := template.ParseFiles("templates/password_reset.html")
				checkErr("Template parsefile error", err)
				t.Execute(w, val)
			}
		}
	}
}

// The OpenWeatherMap API is sent with the request with user's query
// and the JSON result is decoded into the 'CurrentWeather' stuct
func getCurrentWeather(apiUrl string) definition.CurrentWeather {
	var apiClient = &http.Client{Timeout: 10 * time.Second}
	response, err := apiClient.Get(apiUrl)
	checkErr("HTTP Get error", err)

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()

	var reqWeather = definition.CurrentWeather{}
	err = json.Unmarshal(body, &reqWeather)
	
	if err != nil {
		// Query not found in the OpenWeatherMap API
		reqWeather.Code = 404
	}

	return reqWeather
}

func checkErr(message string, err error) {
	if err != nil {
		log.Printf("%s> %s", message, err.Error())
	}
}
