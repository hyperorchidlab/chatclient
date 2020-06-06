package config

import (
	"crypto/ed25519"
	"encoding/json"
	"github.com/kprc/chat-protocol/protocol"
	"github.com/kprc/nbsnetwork/tools"
	"log"
	"os"
	"path"
	"strconv"
	"sync"
)

const (
	CC_HomeDir      = ".chatclient"
	CC_CFG_FileName = "chatclient.json"
)

type ChatClientConfig struct {
	RemoteHttpPort   int    `json:"remotehttpport"`
	RemoteHttpServer string `json:"remotehttpserver"`

	CmdListenPort string `json:"cmdlistenport"`

	RemoteChatPort int `json:"chatport"`

	KeyFile  string `json:"keyfile"`
	UserFile string `json:"userfile"`

	PrivKey ed25519.PrivateKey `json:"-"`
	PubKey  ed25519.PublicKey  `json:"-"`

	SP *protocol.SignPack `json:"-"`

	UpdateFriendTime int64 `json:"updatefriendtime"`

	ChatDataPath string		`json:"chat_data_path"`
	MetaDataPath string     `json:"meta_data_path"`

}

var (
	cccfgInst     *ChatClientConfig
	cccfgInstLock sync.Mutex
)

func (bc *ChatClientConfig) InitCfg() *ChatClientConfig {
	bc.RemoteHttpPort = 50101
	bc.CmdListenPort = "127.0.0.1:50100"

	bc.RemoteChatPort = 50102
	bc.KeyFile = "chat_client.key"
	bc.UserFile = "chat_user.info"
	bc.ChatDataPath = "chat-data"
	bc.MetaDataPath = "meta-data"

	return bc
}

func (bc *ChatClientConfig) Load() *ChatClientConfig {
	if !tools.FileExists(GetCCCFGFile()) {
		return nil
	}

	jbytes, err := tools.OpenAndReadAll(GetCCCFGFile())
	if err != nil {
		log.Println("load file failed", err)
		return nil
	}

	//bc1:=&BASDConfig{}

	err = json.Unmarshal(jbytes, bc)
	if err != nil {
		log.Println("load configuration unmarshal failed", err)
		return nil
	}

	return bc

}

func newCCCfg() *ChatClientConfig {

	bc := &ChatClientConfig{}

	bc.InitCfg()

	return bc
}

func GetCCC() *ChatClientConfig {
	if cccfgInst == nil {
		cccfgInstLock.Lock()
		defer cccfgInstLock.Unlock()
		if cccfgInst == nil {
			cccfgInst = newCCCfg()
		}
	}

	return cccfgInst
}

func PreLoad() *ChatClientConfig {
	bc := &ChatClientConfig{}

	return bc.Load()
}

func LoadFromCfgFile(file string) *ChatClientConfig {
	bc := &ChatClientConfig{}

	bc.InitCfg()

	bcontent, err := tools.OpenAndReadAll(file)
	if err != nil {
		log.Fatal("Load Config file failed")
		return nil
	}

	err = json.Unmarshal(bcontent, bc)
	if err != nil {
		log.Fatal("Load Config From json failed")
		return nil
	}

	cccfgInstLock.Lock()
	defer cccfgInstLock.Unlock()
	cccfgInst = bc

	return bc

}

func LoadFromCmd(initfromcmd func(cmdbc *ChatClientConfig) *ChatClientConfig) *ChatClientConfig {
	cccfgInstLock.Lock()
	defer cccfgInstLock.Unlock()

	lbc := newCCCfg().Load()

	if lbc != nil {
		cccfgInst = lbc
	} else {
		lbc = newCCCfg()
	}

	cccfgInst = initfromcmd(lbc)

	return cccfgInst
}

func GetCCCHomeDir() string {
	curHome, err := tools.Home()
	if err != nil {
		log.Fatal(err)
	}

	return path.Join(curHome, CC_HomeDir)
}

func GetCCCFGFile() string {
	return path.Join(GetCCCHomeDir(), CC_CFG_FileName)
}

func (bc *ChatClientConfig) Save() {
	jbytes, err := json.MarshalIndent(*bc, " ", "\t")

	if err != nil {
		log.Println("Save BASD Configuration json marshal failed", err)
	}

	if !tools.FileExists(GetCCCHomeDir()) {
		os.MkdirAll(GetCCCHomeDir(), 0755)
	}

	err = tools.Save2File(jbytes, GetCCCFGFile())
	if err != nil {
		log.Println("Save BASD Configuration to file failed", err)
	}

}

func (bc *ChatClientConfig) GetKeyPath() string {
	return path.Join(GetCCCHomeDir(), bc.KeyFile)
}

func (bc *ChatClientConfig) GetUserFile() string {
	return path.Join(GetCCCHomeDir(), bc.UserFile)
}

func (bc *ChatClientConfig) GetMetaPath() string  {
	mtp:=path.Join(GetCCCHomeDir(),bc.MetaDataPath)

	if !tools.FileExists(mtp){
		os.MkdirAll(mtp,0755)
	}

	return mtp
}


func (bc *ChatClientConfig) GetAjaxPath() string {
	host := bc.RemoteHttpServer + ":" + strconv.Itoa(bc.RemoteHttpPort)
	return "http://" + host + "/ajax"
}

func (bc *ChatClientConfig) GetRegUrl() string {
	return bc.GetAjaxPath() + "/userreg"
}

func (bc *ChatClientConfig) GetCmdUrl() string {
	return bc.GetAjaxPath() + "/cmd"
}

func IsInitialized() bool {
	if tools.FileExists(GetCCCFGFile()) {
		return true
	}

	return false
}
