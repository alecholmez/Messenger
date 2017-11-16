package users

import (
	"crypto/sha512"
	"encoding/hex"
	"net/http"

	"gopkg.in/mgo.v2/bson"

	"github.com/deciphernow/gm-fabric-go/dbutil"
	"github.com/rs/zerolog/hlog"
)

// User ...
type User struct {
	Name     string `json:"name" bson:"name"`
	Username string `json:"username" bson:"username"`
	Password string `json:"password" bson:"password"`
	Email    string `json:"email" bson:"email"`
	ID       string `json:"_id" bson:"_id"`
}

// Signup ...
// Not the greatest solution but it'll work for a quick prototype
func Signup(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)

	var user User
	err := dbutil.ReadReqest(r, &user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request")
		return
	}
	user.ID = dbutil.CreateHash()

	hasher := sha512.New()
	hasher.Write([]byte(user.Password))
	hash := hex.EncodeToString(hasher.Sum(nil))

	user.Password = hash

	sess, err := dbutil.GetMongo(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve mongo session from context")
		return
	}
	col := sess.DB("messenger").C("users")

	log.Info().Msg("Creating new user: " + user.Username)
	err = col.Insert(user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert user into mongo")
		return
	}

	err = dbutil.WriteJSON(w, user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write json")
		return
	}
}

// Login ...
// Not the greatest solution but it'll work for a quick prototype
func Login(w http.ResponseWriter, r *http.Request) {
	log := hlog.FromRequest(r)

	var l struct {
		Username string `json:"username" bson:"username"`
		Password string `json:"password" bson:"password"`
	}
	err := dbutil.ReadReqest(r, &l)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request")
		return
	}

	hasher := sha512.New()
	hasher.Write([]byte(l.Password))
	hash := hex.EncodeToString(hasher.Sum(nil))

	sess, err := dbutil.GetMongo(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve mongo session from context")
		return
	}
	col := sess.DB("messenger").C("users")

	var users []User
	err = col.Find(bson.M{"username": l.Username}).All(&users)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve potential users from mongo")
		return
	}

	for _, u := range users {
		if u.Password == hash {
			err = dbutil.WriteJSON(w, u)
			if err != nil {
				log.Error().Err(err).Msg("Failed to write json")
			}
			return
		}
	}
}
