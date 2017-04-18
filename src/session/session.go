// Author: Pirakalan

package session

import (
	"github.com/gorilla/securecookie"
	"net/http"
	"settings"
	"helper"
	"time"
)

var (
	s = securecookie.New([]byte(settings.COOKIEHASHKEY), nil)
)

// Set cookie into user's browser with username and sessionKey hash
// Reference: http://www.gorillatoolkit.org/pkg/securecookie
func setCookieHandler(w http.ResponseWriter, r *http.Request, username string, sessionKey string) {
    value := map[string]string{
    	"u": username,
        "key": sessionKey,
    }
    if encoded, err := s.Encode("session", value); err == nil {
        cookie := &http.Cookie{
            Name:  "session",
            Value: encoded,
            Path:  "/",
        }
        http.SetCookie(w, cookie)
    }
}

// Retrieve cookie information from user's browser
// Reference: http://www.gorillatoolkit.org/pkg/securecookie
func ReadCookieHandler(w http.ResponseWriter, r *http.Request) (string, string) {
	cookie, _ := r.Cookie("session")
    value := make(map[string]string)
    if err := s.Decode("session", cookie.Value, &value); err == nil {
        return value["u"], value["key"]
    }
    return "", ""
}

// Check to see if the user's session is valid
func VerifySession(w http.ResponseWriter, r *http.Request) (bool, string) {
	if _, err := r.Cookie("session"); err == nil {
		username, key := ReadCookieHandler(w,r)
		if helper.IsValidSessionKey(username, key) {
			return true, ""
		} else {
			return false, "Session expired, please login"
		}
	}
	return false, ""
}

// Remove the session key when a user logs out
func ClearSession(w http.ResponseWriter, r *http.Request) {
	_, key := ReadCookieHandler(w,r)
	helper.DeleteSessionKey(key)
	helper.UpdateSessionHistory(key, "User logged out", false)
}

// After a user logs in a session key is created. The key is stored into the user's cookie and 
// 'usersSession' table 
func SetSession(w http.ResponseWriter, r *http.Request, username string, password string) {
	sessionKey := helper.CreateSession(username, password)
	setCookieHandler(w, r, username, sessionKey)
	helper.CreateSessionHistory(username, sessionKey)
}

// Periodically clear old sessions (greater than 5 hours)
func CleanSessions(period time.Duration, quit <- chan struct{}) {
	timer := time.NewTicker(period)

	for {
		select {
		case <- timer.C:
			helper.DeleteOldSessions()
		case <-quit:
			timer.Stop()
			return
		}
	}
}
