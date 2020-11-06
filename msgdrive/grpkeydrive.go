package msgdrive

import (
	"github.com/hyperorchidlab/chat-protocol/protocol"
)

var FMsgDrive func(hahsKey string) (*protocol.GroupKeys, error)

func RegMsgDriveFunc(f func(hahsKey string) (*protocol.GroupKeys, error)) {
	FMsgDrive = f
}

func DriveGroupKey(hashKey string) (gks *protocol.GroupKeys, err error) {
	return FMsgDrive(hashKey)
}
