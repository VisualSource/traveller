package handler

import (
	"log"
	"net/http"
	"visualsource/traveller/internal/model"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

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

	statement, _ := h.db.Prepare("INSERT INTO user (id,username,password) VALUES (?,?,?);")

	userid := bson.NewObjectId()

	_, err = statement.Exec(userid.String(), formData.Username, string(psd))
	if err != nil {
		log.Println(err.Error())
		return c.String(http.StatusBadRequest, "Failed to insert")
	}

	err = createSession(userid.String(), c)
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

func (h *Handler) UserJoinSession(c echo.Context) (err error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	log.Printf("User %s\n", sess.Values["sub"])

	//TODO: check if current user can join this session
	// if so add user to session redirect to player creation

	c.Response().Header().Add("HX-Redirect", "/session/in/example-session-id?create=true")
	return c.NoContent(204)
}

func (h *Handler) UserJoinedSessions(c echo.Context) (err error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	log.Printf("User %s\n", sess.Values["sub"])

	return c.Render(http.StatusOK, "session_joined.html", map[string]interface{}{
		"sessions": []model.Session{
			model.Session{
				ID:    bson.NewObjectId(),
				Admin: "USER_ID",
				Name:  "TestSession",
			},
		},
	})
}
