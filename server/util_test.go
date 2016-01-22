package server

import (
	"bytes"
	"errors"
	. "gopkg.in/check.v1"
	"log"
	"net/http"
	"os"
	"strings"
)

type Util struct{}

var _ = Suite(&Util{})

func (u *Util) TestWrapError(c *C) {
	err := errors.New("foo")
	wrapped := wrapError{"bar", err}
	c.Assert(wrapped.Error(), Equals, "bar: foo")
}

func (u *Util) TestThrowNoError(c *C) {
	defer c.Assert(recover(), IsNil)
	throw(nil)
}

func (u *Util) TestThrowWithError(c *C) {
	err := errors.New("foo")
	defer func() {
		c.Assert(recover(), DeepEquals, err)
	}()
	throw(err)
}

func (u *Util) TestCheckAuth(c *C) {
	config = serverConfigs{}
	config.Staff.Classes = make(map[string]staffClass, 1)
	config.Staff.Classes["admin"] = staffClass{
		Rights: map[string]bool{
			"canFoo": true,
			"canBar": false,
		},
	}

	// Staff with rights
	ident := Ident{Auth: "admin"}
	c.Assert(checkAuth("canFoo", ident), Equals, true)

	// Staff without rights
	c.Assert(checkAuth("canBar", ident), Equals, false)
	c.Assert(checkAuth("canBaz", ident), Equals, false)

	// Non-existant staff
	ident = Ident{Auth: "butler"}
	c.Assert(checkAuth("canFoo", ident), Equals, false)

	// Not staff
	ident = Ident{}
	c.Assert(checkAuth("canFoo", ident), Equals, false)
}

func (u *Util) TestLookupIdent(c *C) {
	const ip = "::1"
	ident := Ident{IP: ip}
	c.Assert(lookUpIdent(ip), DeepEquals, ident)
}

func (u *Util) TestCanAccessBoard(c *C) {
	config.Boards.Enabled = []string{"a", "staff"}
	config.Boards.Staff = "staff"
	config.Staff.Classes = make(map[string]staffClass, 1)
	config.Staff.Classes["admin"] = staffClass{
		Rights: map[string]bool{
			"accessStaffBoard": true,
		},
	}
	ident := Ident{}

	// Board exists
	c.Assert(canAccessBoard("a", ident), Equals, true)

	// Board doesn't exist
	c.Assert(canAccessBoard("b", ident), Equals, false)

	// /all/ board
	c.Assert(canAccessBoard("all", ident), Equals, true)

	// Staff board
	c.Assert(canAccessBoard("staff", ident), Equals, false)
	ident.Auth = "admin"
	c.Assert(canAccessBoard("staff", ident), Equals, true)

	// Banned
	ident = Ident{Banned: true}
	c.Assert(canAccessBoard("a", ident), Equals, false)
}

func (u *Util) TestHashBuffer(c *C) {
	c.Assert(hashBuffer([]byte{1, 2, 3}), Equals, "5289df737df57326")
}

type jsonSample struct {
	A int    `json:"a"`
	B string `json:"b"`
}

func (u *Util) TestMarshalJSON(c *C) {
	s := jsonSample{1, "b"}
	c.Assert(string(marshalJSON(s)), Equals, `{"a":1,"b":"b"}`)
}

func (u *Util) TestUnmarshalJSON(c *C) {
	const json = `{"a":1,"b":"b"}`
	var store jsonSample
	result := jsonSample{1, "b"}
	unmarshalJSON([]byte(json), &store)
	c.Assert(store, DeepEquals, result)
}

func (u *Util) TestCopyFile(c *C) {
	buf := new(bytes.Buffer)
	copyFile("./test/frontpage.html", buf)
	c.Assert(buf.String(), Equals, "<!doctype html><html></html>\n")
}

func (u *Util) TestIDToString(c *C) {
	c.Assert(idToString(1), Equals, "1")
}

func (u *Util) TestChooseLang(c *C) {
	const (
		def     = "lv_LV"
		enabled = "en_GB"
	)
	config = serverConfigs{}
	config.Lang.Enabled = []string{enabled}
	config.Lang.Default = def
	req := newRequest(c)

	// No cookie
	c.Assert(chooseLang(req), Equals, def)

	// Non-enabled language
	req = newRequest(c)
	req.AddCookie(&http.Cookie{
		Name:  "lang",
		Value: "pt_BR",
	})
	c.Assert(chooseLang(req), Equals, def)

	// Enabled language
	req = newRequest(c)
	req.AddCookie(&http.Cookie{
		Name:  "lang",
		Value: enabled,
	})
	c.Assert(chooseLang(req), Equals, enabled)
}

func (u *Util) TestLogError(c *C) {
	req := newRequest(c)
	err := errors.New("foo")
	req.RemoteAddr = "::1"
	buf := new(bytes.Buffer)
	log.SetOutput(buf)
	logError(req, err)
	log.SetOutput(os.Stdout)
	c.Assert(
		strings.Split(buf.String(), "\n")[0],
		Matches,
		`\d+/\d+/\d+ \d+:\d+:\d+ panic serving ::1: foo`,
	)
}
