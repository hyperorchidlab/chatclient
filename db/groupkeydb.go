package db

import (
	"encoding/json"
	"github.com/hyperorchidlab/chatserver/db"
	"sync"

	"crypto/sha256"
	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperorchidlab/chatclient/chatcrypt"
	"github.com/hyperorchidlab/chatclient/config"
)

type GroupKeysDb struct {
	db.NbsDbInter
	dbLock sync.Mutex
	cursor *db.DBCusor
}

var (
	gksStore     *GroupKeysDb
	gksStoreLock sync.Mutex
)

func newChatGroupkeysDb() *GroupKeysDb {
	cfg := config.GetCCC()

	db := db.NewFileDb(cfg.GetGrpKeysDbPath()).Load()

	return &GroupKeysDb{NbsDbInter: db}
}

func GetChatGrpKeysDb() *GroupKeysDb {
	if gksStore == nil {
		gksStoreLock.Lock()
		defer gksStoreLock.Unlock()

		if gksStore == nil {
			gksStore = newChatGroupkeysDb()
		}

	}

	return gksStore
}

type GroupKeys struct {
	PubKeys   []string `json:"pub_keys"`
	GroupKeys []string `json:"group_keys"`
}

func (gk *GroupKeys) GenKey() string {

	var pks [][]byte
	for i := 0; i < len(gk.PubKeys); i++ {
		pks = append(pks, base58.Decode(gk.PubKeys[i]))
	}

	r := chatcrypt.InsertionSortDArray(pks)
	r = append(r, base58.Decode(gk.GroupKeys[0]))

	var shabytes []byte

	for i := 0; i < len(r); i++ {
		shabytes = append(shabytes, r[i]...)
	}

	hash := sha256.Sum256(shabytes)

	return base58.Encode(hash[:])
}

func GenKey(pks [][]byte, gpks [][]byte) string {
	r := chatcrypt.InsertionSortDArray(pks)

	r = append(r, gpks[0])

	var shabytes []byte

	for i := 0; i < len(r); i++ {
		shabytes = append(shabytes, r[i]...)
	}

	hash := sha256.Sum256(shabytes)

	return base58.Encode(hash[:])
}

func GenKeyByPubKeys(pks [][]byte, owner []byte) string {
	r := chatcrypt.InsertionSortDArray(pks)

	r = append(r, owner)

	var shabytes []byte

	for i := 0; i < len(r); i++ {
		shabytes = append(shabytes, r[i]...)
	}

	hash := sha256.Sum256(shabytes)

	return base58.Encode(hash[:])

}

func (gkdb *GroupKeysDb) Insert(gks [][]byte, pks [][]byte) (key string) {
	gk := &GroupKeys{}

	for i := 0; i < len(gks); i++ {
		gk.GroupKeys = append(gk.GroupKeys, base58.Encode(gks[i]))
	}

	for i := 0; i < len(pks); i++ {
		gk.PubKeys = append(gk.PubKeys, base58.Encode(pks[i]))
	}

	key = gk.GenKey()

	gkdb.dbLock.Lock()
	defer gkdb.dbLock.Unlock()
	if _, err := gkdb.NbsDbInter.Find(key); err != nil {
		j, _ := json.Marshal(*gk)
		gkdb.NbsDbInter.Insert(key, string(j))
	}

	return
}

func (gkdb *GroupKeysDb) Insert2(gks []string, pks []string) (key string) {
	gk := &GroupKeys{}

	gk.PubKeys = pks
	gk.GroupKeys = gks

	key = gk.GenKey()

	gkdb.dbLock.Lock()
	defer gkdb.dbLock.Unlock()
	if _, err := gkdb.NbsDbInter.Find(key); err != nil {
		j, _ := json.Marshal(*gk)
		gkdb.NbsDbInter.Insert(key, string(j))
	}

	return
}

func (gkdb *GroupKeysDb) Find(key string) *GroupKeys {
	gkdb.dbLock.Lock()
	defer gkdb.dbLock.Unlock()
	if v, err := gkdb.NbsDbInter.Find(key); err != nil {
		return nil
	} else {
		gk := &GroupKeys{}
		json.Unmarshal([]byte(v), gk)

		return gk
	}
}
