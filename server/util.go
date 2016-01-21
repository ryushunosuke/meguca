/*
 Contains various general helper functions
*/

package server

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

// Wrapper type for compund errors errors
type wrapError struct {
	text  string
	inner error
}

func (e wrapError) Error() string {
	text := e.text
	if e.inner != nil {
		text += ": " + e.inner.Error()
	}
	return text
}

// throw panics, if there is an error. Rob Pike must never know.
func throw(err error) {
	if err != nil {
		panic(err)
	}
}

// checkAuth checks if the suplied Ident is priveledged to perform an action
func checkAuth(action string, ident Ident) bool {
	if class, ok := config.Staff.Classes[ident.Auth]; ok {
		return class.Rights[action]
	}
	return false
}

// Determine access rights of an IP
func lookUpIdent(ip string) Ident {
	ident := Ident{IP: ip}

	// TODO: BANS

	return ident
}

// Confirm client has rights to access board
func canAccessBoard(board string, ident Ident) bool {
	if board == config.Boards.Staff && !checkAuth("accessStaffBoard", ident) {
		return false
	}
	_, ok := config.Boards.Boards[board]
	if !ok && board == "all" {
		ok = true
	}
	return !ident.Banned && ok
}

// Confirm thread exists and client has rights to access it's board
func canAccessThread(id uint64, board string, ident Ident) bool {
	if !canAccessBoard(board, ident) {
		return false
	}
	var deleted bool
	db().Do(getThread(id).Field("deleted").Default(false)).One(&deleted)
	if deleted && !checkAuth("seeModeration", ident) {
		return false
	}
	return true
}

// Compute a truncated MD5 hash from a buffer
func hashBuffer(buf []byte) string {
	hasher := md5.New()
	hasher.Write(buf)
	return hex.EncodeToString(hasher.Sum(nil))[16:]
}

// Shorthand for marshaling JSON and handling the error
func marshalJSON(input interface{}) []byte {
	data, err := json.Marshal(input)
	throw(err)
	return data
}

// Shorthand for unmarshalling JSON
func unmarshalJSON(data []byte, store interface{}) {
	throw(json.Unmarshal(data, store))
}

// copyFile reads a file from disk and copies it into the writer
func copyFile(path string, writer io.Writer) {
	file, err := os.Open(path)
	throw(err)
	defer file.Close()
	_, err = io.Copy(writer, file)
	throw(err)
}

// Shorthand for converting a post ID to a string for JSON keys
func idToString(id uint64) string {
	return strconv.FormatUint(id, 10)
}

// chooseLang selects the language to use in responses to the client, by
// checking the language setting of the request's cookies and verifying it
// against the available selection on the server. Defaults to the server's
// default language.
func chooseLang(req *http.Request) string {
	cookie, err := req.Cookie("lang")
	if err == http.ErrNoCookie {
		return config.Lang.Default
	} else if err != nil {
		panic(err)
	}
	for _, lang := range config.Lang.Enabled {
		if cookie.Value == lang {
			return lang
		}
	}
	return config.Lang.Default
}

// Log an error with its stack trace
func logError(req *http.Request, err interface{}) {
	const size = 64 << 10
	buf := make([]byte, size)
	buf = buf[:runtime.Stack(buf, false)]
	log.Printf("panic serving %v: %v\n%s", req.RemoteAddr, err, buf)
}
