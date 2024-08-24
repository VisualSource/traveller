package handler

import (
	"log"
	"net/http"
	"slices"
	"strings"
	"visualsource/traveller/internal/model"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/lucsky/cuid"
	"golang.org/x/crypto/bcrypt"
)

const Session_ID = "sub"

type userFormLogin struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

type userFormSignup struct {
	Username   string `form:"username"`
	Password   string `form:"password"`
	Repassword string `form:"password"`
}

func createSession(id string, c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	sess.Values["sub"] = id

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return nil
}

func (h *Handler) UserSignup(c echo.Context) error {
	var formData userFormSignup
	err := c.Bind(&formData)
	if err != nil {
		return c.String(http.StatusBadRequest, "missing form fields")
	}

	if formData.Password != formData.Repassword {
		return c.String(http.StatusBadRequest, "password does not match")
	}

	psd, err := bcrypt.GenerateFromPassword([]byte(formData.Password), 12)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "Failed to hash password")
	}
	statement, _ := h.db.Prepare("INSERT INTO user (id,username,password) VALUES (?,?,?);")

	userid := cuid.New()

	_, err = statement.Exec(userid, formData.Username, string(psd))
	if err != nil {
		log.Println(err.Error())
		return c.String(http.StatusBadRequest, "Failed to insert")
	}

	err = createSession(userid, c)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "session error")
	}

	c.Response().Header().Add("HX-Redirect", "/session/select")
	return c.NoContent(204)
}

func (h *Handler) UserLogin(c echo.Context) error {
	var formData userFormLogin
	err := c.Bind(&formData)
	if err != nil {
		return c.String(http.StatusBadRequest, "missing form fields")
	}

	statement, _ := h.db.Prepare("SELECT id,password FROM user WHERE username = ? LIMIT 1;")

	var psdHash string
	var id string
	err = statement.QueryRow(formData.Username).Scan(&id, &psdHash)
	if err != nil {
		log.Println(err.Error())
		return c.HTML(http.StatusBadRequest, "Invalid login")
	}

	err = bcrypt.CompareHashAndPassword([]byte(psdHash), []byte(formData.Password))
	if err != nil {
		log.Println(err)
		return c.String(http.StatusBadRequest, "Bad Request")
	}

	err = createSession(id, c)
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "Failed to create session")
	}

	c.Response().Header().Add("HX-Redirect", "/session/select")

	return c.NoContent(204)
}

func (h *Handler) UserLogout(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusPermanentRedirect, "/")
}

type sessionJoinCode struct {
	SessionCode string `form:"session_code"`
}

func (h *Handler) UserJoinSession(c echo.Context) (err error) {
	var formData sessionJoinCode
	err = c.Bind(&formData)
	if err != nil {
		return c.HTML(http.StatusBadRequest, "<span>missing form fields</span>")
	}

	sess, err := session.Get("session", c)
	if err != nil {
		return c.String(http.StatusUnauthorized, "Unauthorized")
	}

	var user model.User

	err = user.GetUser(h.db, sess.Values[Session_ID].(string))
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusNotFound, "Failed to find user")
	}

	var session model.Session
	err = session.GetSession(h.db, formData.SessionCode)
	if err != nil {
		return c.HTML(http.StatusBadRequest, "<span>No session found!</span>")
	}

	if session.Admin == user.Id {
		return c.HTML(http.StatusBadRequest, "<span>Admin can not join a session</span>")
	}

	if !slices.Contains(session.Players, user.Id) {
		return c.HTML(http.StatusBadRequest, "<span>You are not allowed to join this session!</span>")
	}

	if slices.Contains(user.Sessions, session.Id) {
		return c.HTML(http.StatusBadRequest, "<span>You have already joined this session!</span>")
	}

	stmt, err := h.db.Prepare("UPDATE user SET sessions = json_insert(sessions,'$[#]',?) WHERE id = ?;")
	if err != nil {
		return c.HTML(http.StatusInternalServerError, "<span>Failed to update user</span>")
	}

	_, err = stmt.Exec(session.Id, user.Id)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, "<span>Failed to update user</span>")
	}

	sess.AddFlash("NEW_PLAYER", "TRUE")

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	c.Response().Header().Add("HX-Redirect", strings.Join([]string{"/session/in/", session.Id}, ""))
	return c.NoContent(204)
}

func (h *Handler) UserJoinedSessions(c echo.Context) (err error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	userId := sess.Values[Session_ID]

	stmt, err := h.db.Prepare(strings.Join([]string{"SELECT * FROM 'session' WHERE players LIKE '%\"", userId.(string), "\"%'"}, ""))
	if err != nil {
		log.Println(err)
		return c.String(http.StatusInternalServerError, "Failed to load sessions")
	}

	rows, err := stmt.Query()
	if err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, "<span>Failed to load</span>")
	}

	items := []model.Session{}
	for rows.Next() {
		var session model.Session

		err = session.Scan(rows)
		if err != nil {
			log.Panicln(err)
			return c.HTML(http.StatusInternalServerError, "<span>Failed to load</span>")
		}

		items = append(items, session)
	}

	return c.Render(http.StatusOK, "session_joined.html", map[string]interface{}{
		"sessions": items,
		"empty":    len(items) == 0,
	})
}
