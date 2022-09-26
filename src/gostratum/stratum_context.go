package gostratum

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type StratumContext struct {
	ctx           context.Context
	RemoteAddr    string
	WalletAddr    string
	WorkerName    string
	RemoteApp     string
	Logger        *zap.SugaredLogger
	connection    net.Conn
	disconnecting bool
	onDisconnect  chan *StratumContext
	State         any // gross, but go generics aren't mature enough this can be typed 😭
}

func NewMockContext(ctx context.Context, logger *zap.SugaredLogger, state any) (*StratumContext, *MockConnection) {
	mc := NewMockConnection()
	return &StratumContext{
		ctx:        ctx,
		State:      state,
		RemoteAddr: "127.0.0.1",
		WalletAddr: uuid.NewString(),
		WorkerName: uuid.NewString(),
		RemoteApp:  "mock.context",
		Logger:     logger,
		connection: mc,
	}, mc
}

var ErrorDisconnected = fmt.Errorf("disconnecting")

func (sc *StratumContext) Connected() bool {
	return !sc.disconnecting
}

func (sc *StratumContext) Reply(response JsonRpcResponse) error {
	if sc.disconnecting {
		return ErrorDisconnected
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		return errors.Wrap(err, "failed encoding jsonrpc response")
	}
	encoded = append(encoded, '\n')
	_, err = sc.connection.Write(encoded)
	sc.checkDisconnect(err)
	return err
}

func (sc *StratumContext) Send(event JsonRpcEvent) error {
	if sc.disconnecting {
		return ErrorDisconnected
	}
	encoded, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed encoding jsonrpc event")
	}
	encoded = append(encoded, '\n')
	_, err = sc.connection.Write(encoded)
	sc.checkDisconnect(err)
	return err
}

func (sc *StratumContext) ReplyStaleShare(id any) error {
	return sc.Reply(JsonRpcResponse{
		Id:     id,
		Result: nil,
		Error:  []any{21, "Job not found", nil},
	})
}
func (sc *StratumContext) ReplyDupeShare(id any) error {
	return sc.Reply(JsonRpcResponse{
		Id:     id,
		Result: nil,
		Error:  []any{22, "Duplicate share submitted", nil},
	})
}

func (sc *StratumContext) ReplyBadShare(id any) error {
	return sc.Reply(JsonRpcResponse{
		Id:     id,
		Result: nil,
		Error:  []any{20, "Unknown problem", nil},
	})
}

func (sc *StratumContext) ReplyLowDiffShare(id any) error {
	return sc.Reply(JsonRpcResponse{
		Id:     id,
		Result: nil,
		Error:  []any{23, "Invalid difficulty", nil},
	})
}

func (sc *StratumContext) checkDisconnect(err error) {
	if err != nil { // actual error
		sc.disconnecting = true
		sc.onDisconnect <- sc
	}
}

// Context interface impl

func (StratumContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (StratumContext) Done() <-chan struct{} {
	return nil
}

func (StratumContext) Err() error {
	return nil
}

func (d StratumContext) Value(key any) any {
	return d.ctx.Value(key)
}
