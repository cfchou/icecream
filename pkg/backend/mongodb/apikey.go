package mongodb

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	pe "github.com/pkg/errors"
)

type mAPIKey struct {
	ID bson.ObjectId `bson:"_id,omitempty" json:"_id,omitempty"`
	// ApiKey is mandatory.
	APIKey      string `json:"apikey"`
}

type MongoAPIKeyBackend struct {
	session *mgo.Session
}

func (h *MongoAPIKeyBackend) Authenticate(apiKey string) error {
	q := h.session.DB("").C("apikey").Find(
		&bson.M{"apikey": apiKey})

	var keys []mAPIKey
	if err := q.All(&keys); err != nil {
		log.Error("Query.All failed", "err", err)
		return pe.WithStack(err)
	}
	if len(keys) == 0 {
		return pe.WithStack(mgo.ErrNotFound)
	} else if len(keys) > 1 {
		// By design this should not happen. Most likely a duplicated product is
		// added in an out-of-band fashion.
		log.Error("Find gets more than 1", "err", ErrInconsistent)
		return pe.WithStack(ErrInconsistent)
	}
	log.Debug(fmt.Sprintf("Find _id=%s", keys[0].ID.Hex()))
	return nil
}
