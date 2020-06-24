package db

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/kprc/chat-protocol/address"
	"github.com/kprc/chat-protocol/groupid"
	"github.com/kprc/chatclient/config"
	"github.com/kprc/nbsnetwork/common/list"
	"github.com/kprc/nbsnetwork/tools"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"path"
	"sync"
)

type MetaDbIntf interface {
	AddFriend(alias string, addr address.ChatAddress, agree int, addTime int64)
	DelFriend(addr address.ChatAddress)
	FindFriend(addr address.ChatAddress) (*Friend,error)

	AddGroup(name string, id groupid.GrpID, isOwner bool, createTime int64)
	DelGroup(id groupid.GrpID) error
	FindGroup(gid groupid.GrpID) (g *Group, err error)

	AddGroupMember(id groupid.GrpID, name string, addr address.ChatAddress, agree int, joinTime int64) error
	DelGroupMember(id groupid.GrpID, addr address.ChatAddress) error

	TraversFriends(arg interface{}, do func(arg, v interface{}) (ret interface{}, err error))
	TraversGroups(arg interface{}, do func(arg, v interface{}) (ret interface{}, err error))
	TraversGroupMembers(gid groupid.GrpID, arg interface{}, do func(arg, v interface{}) (ret interface{}, err error))

	Load() MetaDbIntf
	Save()
}

type CMPIndex interface {
	GetIndex() string
}

type Friend struct {
	AliasName string              `json:"alias_name"`
	Addr      address.ChatAddress `json:"addr"`
	Agree     int                 `json:"agree"`
	AddTime   int64               `json:"add_time"`
}

func (f *Friend) String() string {
	msg := ""
	msg += fmt.Sprintf("%-20s", f.AliasName)
	msg += fmt.Sprintf("%-48s", f.Addr.String())
	msg += fmt.Sprintf("%-4d", f.Agree)
	msg += fmt.Sprintf("%-20s", tools.Int64Time2String(f.AddTime))

	return msg
}

func (f *Friend) GetIndex() string {
	return f.Addr.String()
}

type GroupMember struct {
	AliasName     string              `json:"alias_name"`
	Addr          address.ChatAddress `json:"addr"`
	Agree         int                 `json:"agree"`
	JoinGroupTime int64               `json:"join_group_time"`
}

func (gm *GroupMember) String() string {
	msg := ""

	msg += fmt.Sprintf("%-20s", gm.AliasName)
	msg += fmt.Sprintf("%-48s", gm.Addr.String())
	msg += fmt.Sprintf("%-4d", gm.Agree)
	msg += fmt.Sprintf("%-20s", tools.Int64Time2String(gm.JoinGroupTime))

	return msg
}

func (gm *GroupMember) GetIndex() string {
	return gm.Addr.String()
}

type Group struct {
	GroupName  string        `json:"group_name"`
	GroupId    groupid.GrpID `json:"group_id"`
	IsOwner    bool          `json:"is_owner"`
	CreateTime int64         `json:"create_time"`
	LMember    list.List     `json:"-"`
}

func (g *Group) String() string {
	msg := ""
	msg += fmt.Sprintf("%-20s", g.GroupName)
	msg += fmt.Sprintf("%-48s", g.GroupId.String())
	msg += fmt.Sprintf("%-6v", g.IsOwner)
	msg += fmt.Sprintf("%-20s", tools.Int64Time2String(g.CreateTime))

	return msg
}

func (g *Group) GetIndex() string {
	return g.GroupId.String()
}

type MetaDb struct {
	Lock    sync.Mutex
	LFriend list.List //*Friend
	LGroup  list.List //*Group

	friendFile *os.File
	groupFile  *os.File
	grpMbrFile *os.File
}

func medatacmp(v1, v2 interface{}) int {
	i1, i2 := v1.(CMPIndex), v2.(CMPIndex)

	if i1.GetIndex() == i2.GetIndex() {
		return 0
	}

	return 1

}

func NewMetaDb() MetaDbIntf {
	md := &MetaDb{}

	md.LFriend = list.NewList(medatacmp)

	md.LGroup = list.NewList(medatacmp)

	return md
}

var (
	metaDBStore MetaDbIntf
	metaDBLock  sync.Mutex
)

const (
	friendDbFileName      string = "friend.db"
	groupDbFileName       string = "group.db"
	groupmemberDbFileName string = "grpmbr.db"
)

func GetMetaDb() MetaDbIntf {

	if metaDBStore != nil {
		return metaDBStore
	}

	metaDBLock.Lock()
	defer metaDBLock.Unlock()
	if metaDBStore != nil {
		return metaDBStore
	}

	metaDBStore = NewMetaDb()

	return metaDBStore
}

func RenewMetaDb() MetaDbIntf {
	metaDBStore = nil

	return GetMetaDb()
}

func (md *MetaDb) AddFriend(alias string, addr address.ChatAddress, agree int, addTime int64) {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	frd := md.addFriend(alias, addr, agree, addTime)
	if frd != nil {
		data, _ := json.Marshal(*frd)
		md.appendFriend(data, false)
	}
}

func (md *MetaDb) addFriend(alias string, addr address.ChatAddress, agree int, addTime int64) *Friend {
	frd := &Friend{AliasName: alias, Addr: addr, Agree: agree, AddTime: addTime}

	node := md.LFriend.Find(frd)
	if node == nil {
		md.LFriend.AddValue(frd)
	} else {
		md.LFriend.FindDo(frd, func(arg interface{}, v interface{}) (ret interface{}, err error) {
			f, vf := arg.(*Friend), v.(*Friend)
			vf.Agree = f.Agree
			vf.AliasName = f.AliasName

			f.AddTime = vf.AddTime

			return nil, nil
		})
	}

	return frd

}

func (md *MetaDb) DelFriend(addr address.ChatAddress) {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	f := md.delFriend(addr)
	if f != nil {
		data, _ := json.Marshal(*f)
		md.appendFriend(data, true)
	}

}

func (md *MetaDb)FindFriend(addr address.ChatAddress) (*Friend,error)  {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	farg:=Friend{Addr: addr}

	if f:=md.LFriend.Find(farg);f==nil{
		return nil,errors.New("not found")
	}else{
		return f.Value.(*Friend),nil
	}
}


func (md *MetaDb) delFriend(addr address.ChatAddress) *Friend {
	frd := &Friend{Addr: addr}

	_, err := md.LFriend.FindDo(frd, func(arg interface{}, v interface{}) (ret interface{}, err error) {
		f1, f2 := arg.(*Friend), v.(*Friend)
		f1.AliasName = f2.AliasName
		f1.AddTime = f2.AddTime
		f1.Agree = f2.Agree
		return nil, nil
	})

	if err != nil {
		return nil
	} else {
		md.LFriend.DelValue(frd)
	}

	return frd

}

func (md *MetaDb) AddGroup(name string, id groupid.GrpID, isOwner bool, createTime int64) {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	g := md.addGroup(name, id, isOwner, createTime)
	if g == nil {
		return
	}

	data, _ := json.Marshal(*g)
	md.appendGroup(data, false)

}

func (md *MetaDb) addGroup(name string, id groupid.GrpID, isOwner bool, createTime int64) *Group {
	grp := &Group{GroupName: name, GroupId: id, IsOwner: isOwner, CreateTime: createTime}
	node := md.LGroup.Find(grp)
	if node == nil {
		grp.LMember = list.NewList(medatacmp)
		md.LGroup.AddValue(grp)
	} else {
		md.LGroup.FindDo(grp, func(arg interface{}, v interface{}) (ret interface{}, err error) {
			g1, g2 := arg.(*Group), v.(*Group)

			g2.GroupName = g1.GroupName
			g1.CreateTime = g2.CreateTime
			g1.IsOwner = g2.IsOwner

			return nil, nil
		})
	}

	return grp
}

func (md *MetaDb) DelGroup(id groupid.GrpID) error {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	g, err := md.delGroup(id)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(*g)
	return md.appendGroup(data, true)
}

func (md *MetaDb) delGroup(id groupid.GrpID) (*Group, error) {
	g := &Group{GroupId: id}

	mrbcnt, err := md.LGroup.FindDo(g, func(arg interface{}, v interface{}) (ret interface{}, err error) {
		g1 := v.(*Group)
		g2 := arg.(*Group)

		g2.IsOwner = g1.IsOwner
		g2.CreateTime = g1.CreateTime
		g2.GroupName = g1.GroupName

		return g1.LMember.Count(), nil
	})
	if err != nil {
		return nil, err
	}

	cnt := mrbcnt.(int32)
	if cnt == 0 {
		md.LGroup.DelValue(g)
		return g, nil
	}
	return nil, errors.New("members in the group")
}

func (md *MetaDb) FindGroup(gid groupid.GrpID) (g *Group, err error) {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	gArg := &Group{GroupId: gid}

	node := md.LGroup.Find(gArg)
	if node == nil {
		return nil, errors.New("not find")
	}

	g = node.Value.(*Group)

	return
}

type GroupMbrWithOwner struct {
	GroupId groupid.GrpID `json:"group_id"`
	GrpMbr  *GroupMember  `json:"grp_mbr"`
}

func (gmo *GroupMbrWithOwner) GetIndex() string {
	return gmo.GroupId.String()
}

func (md *MetaDb) AddGroupMember(id groupid.GrpID, name string, addr address.ChatAddress, agree int, joinTime int64) error {
	md.Lock.Lock()
	defer md.Lock.Unlock()
	gmo, err := md.addGroupMember(id, name, addr, agree, joinTime)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(*gmo)
	return md.appendGroupMember(data, false)
}

func (md *MetaDb) addGroupMember(id groupid.GrpID, name string, addr address.ChatAddress, agree int, joinTime int64) (*GroupMbrWithOwner, error) {
	g := &Group{GroupId: id}

	node := md.LGroup.Find(g)
	if node == nil {
		return nil, errors.New("Group Not found")
	}

	gmbr := &GroupMember{}
	gmbr.AliasName = name
	gmbr.Addr = addr
	gmbr.Agree = agree
	gmbr.JoinGroupTime = joinTime

	arg := &GroupMbrWithOwner{id, gmbr}

	md.LGroup.FindDo(arg, func(arg interface{}, v interface{}) (ret interface{}, err error) {
		gv := v.(*Group)
		gmo := arg.(*GroupMbrWithOwner)

		n := gv.LMember.Find(gmo.GrpMbr)
		if n == nil {
			gv.LMember.AddValue(gmo.GrpMbr)
		} else {
			gv.LMember.FindDo(gmo.GrpMbr, func(arg interface{}, v interface{}) (ret interface{}, err error) {
				gmbr1, gmbr2 := arg.(*GroupMember), v.(*GroupMember)

				gmbr2.AliasName = gmbr1.AliasName
				gmbr2.Agree = gmbr1.Agree

				return nil, nil

			})
		}
		return nil, nil
	})

	return arg, nil
}

func (md *MetaDb) DelGroupMember(id groupid.GrpID, addr address.ChatAddress) error {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	gmo, err := md.delGroupMember(id, addr)
	if err != nil {
		return err
	}

	data, _ := json.Marshal(*gmo)

	return md.appendGroupMember(data, true)
}

func (md *MetaDb) delGroupMember(id groupid.GrpID, addr address.ChatAddress) (*GroupMbrWithOwner, error) {
	g := &Group{GroupId: id}
	node := md.LGroup.Find(g)
	if node == nil {
		return nil, errors.New("Group Not Found")
	}

	arg := &GroupMbrWithOwner{GroupId: id, GrpMbr: &GroupMember{Addr: addr}}

	gmo, err := md.LGroup.FindDo(arg, func(arg interface{}, v interface{}) (ret interface{}, err error) {
		gmo := arg.(*GroupMbrWithOwner)
		gv := v.(*Group)

		if n := gv.LMember.Find(gmo.GrpMbr); n == nil {
			return nil, errors.New("group member not found")
		}

		gv.LMember.DelValue(gmo.GrpMbr)

		return gmo, nil

	})

	if err != nil {
		return nil, err
	}

	return gmo.(*GroupMbrWithOwner), nil
}

func (md *MetaDb) loadFriend() error {
	fdbPath := path.Join(config.GetCCC().GetMetaPath(), friendDbFileName)

	flag := os.O_RDWR | os.O_APPEND

	if !tools.FileExists(fdbPath) {
		flag |= os.O_CREATE
	}

	f, err := os.OpenFile(fdbPath, flag, 0755)
	if err != nil {
		return errors.New("Open friend file failed")
	}
	defer f.Close()

	bf := bufio.NewReader(f)

	for {
		if line, _, err := bf.ReadLine(); err != nil {
			if err == io.EOF {
				break
			}

			if err == bufio.ErrBufferFull {
				return err
			}

		} else {
			if len(line) > 0 {

				ul := line

				if line[0] == '-' {
					ul = line[1:]
				}

				f := &Friend{}

				err := json.Unmarshal(ul, f)
				if err != nil {
					log.Println(err)
					continue
				}

				if line[0] == '-' {
					md.delFriend(f.Addr)
				} else {
					md.addFriend(f.AliasName, f.Addr, f.Agree, f.AddTime)
				}

			}
		}
	}

	return nil
}

func safeWrite(f *os.File, data []byte) error {
	total := len(data)

	pos := 0

	for {
		if n, err := f.Write(data[pos:]); err != nil {
			return err
		} else {
			pos += n
			if pos >= total {
				return nil
			}
		}
	}

}

func safeAppend(filename string, data []byte) error {
	flag := os.O_RDWR | os.O_APPEND

	if !tools.FileExists(filename) {
		flag |= os.O_CREATE
	}

	f, err := os.OpenFile(filename, flag, 0755)
	if err != nil {
		return errors.New("Open friend file failed")
	}
	defer f.Close()

	err = safeWrite(f, data)
	if err != nil {
		return err
	}

	return safeWrite(f, []byte("\r\n"))
}

func (md *MetaDb) appendFriend(data []byte, delflag bool) error {
	fdbPath := path.Join(config.GetCCC().GetMetaPath(), friendDbFileName)

	var wdata []byte

	if delflag {
		wdata = []byte("-")
		wdata = append(wdata, data...)

	} else {
		wdata = data
	}

	return safeAppend(fdbPath, data)
}

func (md *MetaDb) loadGroup() error {
	gdbPath := path.Join(config.GetCCC().GetMetaPath(), groupDbFileName)

	flag := os.O_RDWR | os.O_APPEND

	if !tools.FileExists(gdbPath) {
		flag |= os.O_CREATE
	}

	f, err := os.OpenFile(gdbPath, flag, 0755)
	if err != nil {
		return errors.New("Open group file failed")
	}
	defer f.Close()

	bf := bufio.NewReader(f)

	for {
		if line, _, err := bf.ReadLine(); err != nil {
			if err == io.EOF {
				break
			}

			if err == bufio.ErrBufferFull {
				return err
			}

		} else {
			if len(line) > 0 {
				ul := line
				if line[0] == '-' {
					ul = line[1:]
				}
				g := &Group{}

				err := json.Unmarshal(ul, g)
				if err != nil {
					log.Println(err)
					continue
				}
				if line[0] == '-' {
					md.delGroup(g.GroupId)
				} else {
					md.addGroup(g.GroupName, g.GroupId, g.IsOwner, g.CreateTime)
				}

			}
		}
	}

	return nil
}

func (md *MetaDb) appendGroup(data []byte, delflag bool) error {
	gdbPath := path.Join(config.GetCCC().GetMetaPath(), groupDbFileName)

	var wdata []byte

	if delflag {
		wdata = []byte("-")
		wdata = append(wdata, data...)

	} else {
		wdata = data
	}

	return safeAppend(gdbPath, wdata)
}

func (md *MetaDb) loadGroupMember() error {
	gmdbPath := path.Join(config.GetCCC().GetMetaPath(), groupmemberDbFileName)

	flag := os.O_RDWR | os.O_APPEND

	if !tools.FileExists(gmdbPath) {
		flag |= os.O_CREATE
	}

	f, err := os.OpenFile(gmdbPath, flag, 0755)
	if err != nil {
		return errors.New("Open group file failed")
	}
	defer f.Close()

	bf := bufio.NewReader(f)

	for {
		if line, _, err := bf.ReadLine(); err != nil {
			if err == io.EOF {
				break
			}

			if err == bufio.ErrBufferFull {
				return err
			}

		} else {
			if len(line) > 0 {
				ul := line
				if line[0] == '-' {
					ul = line[1:]
				}

				gmo := &GroupMbrWithOwner{}

				err := json.Unmarshal(ul, gmo)
				if err != nil {
					log.Println(err)
					continue
				}

				gm := gmo.GrpMbr
				if line[0] == '-' {
					md.delGroupMember(gmo.GroupId, gm.Addr)
				} else {
					md.addGroupMember(gmo.GroupId, gm.AliasName, gm.Addr, gm.Agree, gm.JoinGroupTime)
				}

			}
		}
	}

	return nil
}

func (md *MetaDb) appendGroupMember(data []byte, delflag bool) error {
	gmdbPath := path.Join(config.GetCCC().GetMetaPath(), groupmemberDbFileName)

	var wdata []byte

	if delflag {
		wdata = []byte("-")
		wdata = append(wdata, data...)

	} else {
		wdata = data
	}

	return safeAppend(gmdbPath, data)
}

func (md *MetaDb) Load() MetaDbIntf {
	err := md.loadFriend()
	if err != nil {
		log.Println(err)
	}
	err = md.loadGroup()
	if err != nil {
		log.Println(err)
	}

	err = md.loadGroupMember()
	if err != nil {
		log.Println(err)
	}

	return md
}

func (md *MetaDb) writeFriend(data []byte) error {

	if md.friendFile == nil {
		flag := os.O_WRONLY | os.O_TRUNC

		fdbPath := path.Join(config.GetCCC().GetMetaPath(), friendDbFileName)

		if !tools.FileExists(fdbPath) {
			flag |= os.O_CREATE
		}
		if f, err := os.OpenFile(fdbPath, flag, 0755); err != nil {
			log.Println("Can't open friend file")
			return errors.New("Can't open friend file")
		} else {
			md.friendFile = f
		}
	}

	md.friendFile.Write(data)

	return nil

}

func (md *MetaDb) writeGroup(data []byte) error {
	if md.groupFile == nil {
		flag := os.O_WRONLY | os.O_TRUNC

		gdbPath := path.Join(config.GetCCC().GetMetaPath(), groupDbFileName)

		if !tools.FileExists(gdbPath) {
			flag |= os.O_CREATE
		}
		if f, err := os.OpenFile(gdbPath, flag, 0755); err != nil {
			log.Println("Can't open group file")
			return errors.New("Can't open group file")
		} else {
			md.groupFile = f
		}
	}

	md.groupFile.Write(data)

	return nil
}

func (md *MetaDb) writeGrpMbr(data []byte) error {
	if md.grpMbrFile == nil {
		flag := os.O_WRONLY | os.O_TRUNC

		gmdbPath := path.Join(config.GetCCC().GetMetaPath(), groupmemberDbFileName)

		if !tools.FileExists(gmdbPath) {
			flag |= os.O_CREATE
		}
		if f, err := os.OpenFile(gmdbPath, flag, 0755); err != nil {
			log.Println("Can't open group file")
			return errors.New("Can't open group file")
		} else {
			md.grpMbrFile = f
		}
	}

	md.grpMbrFile.Write(data)

	return nil
}

func (md *MetaDb) cleanup() {
	if md.friendFile != nil {
		md.friendFile.Close()
		md.friendFile = nil
	}

	if md.groupFile != nil {
		md.groupFile.Close()
		md.groupFile = nil
	}

	if md.grpMbrFile != nil {
		md.grpMbrFile.Close()
		md.grpMbrFile = nil
	}

}

func (md *MetaDb) Save() {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	md.LFriend.Traverse(nil, func(arg interface{}, v interface{}) (ret interface{}, err error) {
		f := v.(*Friend)

		data, _ := json.Marshal(*f)

		md.writeFriend(data)
		md.writeFriend([]byte("\r\n"))

		return nil, nil

	})

	md.LGroup.Traverse(nil, func(arg interface{}, v interface{}) (ret interface{}, err error) {
		g := v.(*Group)
		data, _ := json.Marshal(*g)
		md.writeGroup(data)
		md.writeGroup([]byte("\r\n"))

		return nil, nil
	})

	md.LGroup.Traverse(nil, func(arg interface{}, v interface{}) (ret interface{}, err error) {
		g := v.(*Group)
		g.LMember.Traverse(g.GroupId, func(arg interface{}, v interface{}) (ret interface{}, err error) {
			id := arg.(groupid.GrpID)
			gm := v.(*GroupMember)

			gmo := &GroupMbrWithOwner{}
			gmo.GroupId = id
			gmo.GrpMbr = gm

			data, _ := json.Marshal(*gmo)
			md.writeGrpMbr(data)
			md.writeGrpMbr([]byte("\r\n"))

			return nil, nil

		})

		return nil, nil
	})

	md.cleanup()

}

func (md *MetaDb) TraversFriends(arg interface{}, do func(arg, v interface{}) (ret interface{}, err error)) {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	md.LFriend.Traverse(arg, do)
}

func (md *MetaDb) TraversGroups(arg interface{}, do func(arg, v interface{}) (ret interface{}, err error)) {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	md.LGroup.Traverse(arg, do)
}

func (md *MetaDb) TraversGroupMembers(gid groupid.GrpID, arg interface{}, do func(arg, v interface{}) (ret interface{}, err error)) {
	md.Lock.Lock()
	defer md.Lock.Unlock()

	ag := &Group{}
	ag.GroupId = gid

	n := md.LGroup.Find(ag)

	if n == nil {
		return
	}

	g := n.Value.(*Group)

	g.LMember.Traverse(arg, do)
}
