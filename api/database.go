package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tevino/abool"
	"github.com/tidwall/gjson"

	"github.com/safing/portbase/container"
	"github.com/safing/portbase/database"
	"github.com/safing/portbase/database/query"
	"github.com/safing/portbase/database/record"
	"github.com/safing/portbase/log"
)

const (
	dbMsgTypeOk      = "ok"
	dbMsgTypeError   = "error"
	dbMsgTypeDone    = "done"
	dbMsgTypeSuccess = "success"
	dbMsgTypeUpd     = "upd"
	dbMsgTypeNew     = "new"
	dbMsgTypeDel     = "del"
	dbMsgTypeWarning = "warning"

	dbAPISeperator = "|"
	emptyString    = ""
)

var (
	dbAPISeperatorBytes = []byte(dbAPISeperator)
)

func init() {
	RegisterHandleFunc("/api/database/v1", startDatabaseAPI) // net/http pattern matching only this exact path
}

// DatabaseAPI is a database API instance.
type DatabaseAPI struct {
	conn      *websocket.Conn
	sendQueue chan []byte
	subs      map[string]*database.Subscription

	shutdownSignal chan struct{}
	shuttingDown   *abool.AtomicBool
	db             *database.Interface
}

func allowAnyOrigin(r *http.Request) bool {
	return true
}

func startDatabaseAPI(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		CheckOrigin:     allowAnyOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 65536,
	}
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errMsg := fmt.Sprintf("could not upgrade: %s", err)
		log.Error(errMsg)
		http.Error(w, errMsg, 400)
		return
	}

	new := &DatabaseAPI{
		conn:           wsConn,
		sendQueue:      make(chan []byte, 100),
		subs:           make(map[string]*database.Subscription),
		shutdownSignal: make(chan struct{}),
		shuttingDown:   abool.NewBool(false),
		db:             database.NewInterface(nil),
	}

	go new.handler()
	go new.writer()

	log.Infof("api request: init websocket %s %s", r.RemoteAddr, r.RequestURI)
}

func (api *DatabaseAPI) handler() {

	// 123|get|<key>
	//    123|ok|<key>|<data>
	//    123|error|<message>
	// 124|query|<query>
	//    124|ok|<key>|<data>
	//    124|done
	//    124|error|<message>
	//    124|warning|<message> // error with single record, operation continues
	// 125|sub|<query>
	//    125|upd|<key>|<data>
	//    125|new|<key>|<data>
	//    127|del|<key>
	//    125|warning|<message> // error with single record, operation continues
	// 127|qsub|<query>
	//    127|ok|<key>|<data>
	//    127|done
	//    127|error|<message>
	//    127|upd|<key>|<data>
	//    127|new|<key>|<data>
	//    127|del|<key>
	//    127|warning|<message> // error with single record, operation continues

	// 128|create|<key>|<data>
	//    128|success
	//    128|error|<message>
	// 129|update|<key>|<data>
	//    129|success
	//    129|error|<message>
	// 130|insert|<key>|<data>
	//    130|success
	//    130|error|<message>
	// 131|delete|<key>
	//    131|success
	//    131|error|<message>

	for {

		_, msg, err := api.conn.ReadMessage()
		if err != nil {
			if !api.shuttingDown.IsSet() {
				api.shutdown()
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Warningf("api: websocket write error: %s", err)
				}
			}
			return
		}

		parts := bytes.SplitN(msg, []byte("|"), 3)
		if len(parts) != 3 {
			api.send(nil, dbMsgTypeError, "bad request: malformed message", nil)
			continue
		}

		switch string(parts[1]) {
		case "get":
			// 123|get|<key>
			go api.handleGet(parts[0], string(parts[2]))
		case "query":
			// 124|query|<query>
			go api.handleQuery(parts[0], string(parts[2]))
		case "sub":
			// 125|sub|<query>
			go api.handleSub(parts[0], string(parts[2]))
		case "qsub":
			// 127|qsub|<query>
			go api.handleQsub(parts[0], string(parts[2]))
		case "create", "update", "insert":
			// split key and payload
			dataParts := bytes.SplitN(parts[2], []byte("|"), 2)
			if len(dataParts) != 2 {
				api.send(nil, dbMsgTypeError, "bad request: malformed message", nil)
				continue
			}

			switch string(parts[1]) {
			case "create":
				// 128|create|<key>|<data>
				go api.handlePut(parts[0], string(dataParts[0]), dataParts[1], true)
			case "update":
				// 129|update|<key>|<data>
				go api.handlePut(parts[0], string(dataParts[0]), dataParts[1], false)
			case "insert":
				// 130|insert|<key>|<data>
				go api.handleInsert(parts[0], string(dataParts[0]), dataParts[1])
			}
		case "delete":
			// 131|delete|<key>
			go api.handleDelete(parts[0], string(parts[2]))
		default:
			api.send(parts[0], dbMsgTypeError, "bad request: unknown method", nil)
		}
	}
}

func (api *DatabaseAPI) writer() {
	var data []byte
	var err error

	for {
		data = nil

		select {
		// prioritize direct writes
		case data = <-api.sendQueue:
			if len(data) == 0 {
				api.shutdown()
				return
			}
		case <-api.shutdownSignal:
			return
		}

		// log.Tracef("api: sending %s", string(*msg))
		err = api.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			if !api.shuttingDown.IsSet() {
				api.shutdown()
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Warningf("api: websocket write error: %s", err)
				}
			}
			return
		}

	}
}

func (api *DatabaseAPI) send(opID []byte, msgType string, msgOrKey string, data []byte) {
	c := container.New(opID)
	c.Append(dbAPISeperatorBytes)
	c.Append([]byte(msgType))

	if msgOrKey != emptyString {
		c.Append(dbAPISeperatorBytes)
		c.Append([]byte(msgOrKey))
	}

	if len(data) > 0 {
		c.Append(dbAPISeperatorBytes)
		c.Append(data)
	}

	api.sendQueue <- c.CompileData()
}

func (api *DatabaseAPI) handleGet(opID []byte, key string) {
	// 123|get|<key>
	//    123|ok|<key>|<data>
	//    123|error|<message>

	var data []byte

	r, err := api.db.Get(key)
	if err == nil {
		data, err = r.Marshal(r, record.JSON)
	}
	if err == nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil) //nolint:nilness // FIXME: possibly false positive (golangci-lint govet/nilness)
		return
	}
	api.send(opID, dbMsgTypeOk, r.Key(), data)
}

func (api *DatabaseAPI) handleQuery(opID []byte, queryText string) {
	// 124|query|<query>
	//    124|ok|<key>|<data>
	//    124|done
	//    124|warning|<message>
	//    124|error|<message>
	//    124|warning|<message> // error with single record, operation continues

	var err error

	q, err := query.ParseQuery(queryText)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}

	api.processQuery(opID, q)
}

func (api *DatabaseAPI) processQuery(opID []byte, q *query.Query) (ok bool) {
	it, err := api.db.Query(q)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return false
	}

	for r := range it.Next {
		r.Lock()
		data, err := r.Marshal(r, record.JSON)
		r.Unlock()
		if err != nil {
			api.send(opID, dbMsgTypeWarning, err.Error(), nil)
		}
		api.send(opID, dbMsgTypeOk, r.Key(), data)
	}
	if it.Err() != nil {
		api.send(opID, dbMsgTypeError, it.Err().Error(), nil)
		return false
	}

	for {
		select {
		case <-api.shutdownSignal:
			// cancel query and return
			it.Cancel()
			return
		case r := <-it.Next:
			// process query feed
			if r != nil {
				// process record
				r.Lock()
				data, err := r.Marshal(r, record.JSON)
				r.Unlock()
				if err != nil {
					api.send(opID, dbMsgTypeWarning, err.Error(), nil)
				}
				api.send(opID, dbMsgTypeOk, r.Key(), data)
			} else {
				// sub feed ended
				if it.Err() != nil {
					api.send(opID, dbMsgTypeError, it.Err().Error(), nil)
					return false
				}
				api.send(opID, dbMsgTypeDone, emptyString, nil)
				return true
			}
		}
	}
}

// func (api *DatabaseAPI) runQuery()

func (api *DatabaseAPI) handleSub(opID []byte, queryText string) {
	// 125|sub|<query>
	//    125|upd|<key>|<data>
	//    125|new|<key>|<data>
	//    125|delete|<key>
	//    125|warning|<message> // error with single record, operation continues
	var err error

	q, err := query.ParseQuery(queryText)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}

	sub, ok := api.registerSub(opID, q)
	if !ok {
		return
	}
	api.processSub(opID, sub)
}

func (api *DatabaseAPI) registerSub(opID []byte, q *query.Query) (sub *database.Subscription, ok bool) {
	var err error
	sub, err = api.db.Subscribe(q)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return nil, false
	}
	return sub, true
}

func (api *DatabaseAPI) processSub(opID []byte, sub *database.Subscription) {
	for {
		select {
		case <-api.shutdownSignal:
			// cancel sub and return
			_ = sub.Cancel()
			return
		case r := <-sub.Feed:
			// process sub feed
			if r != nil {
				// process record
				r.Lock()
				data, err := r.Marshal(r, record.JSON)
				r.Unlock()
				if err != nil {
					api.send(opID, dbMsgTypeWarning, err.Error(), nil)
					continue
				}
				// TODO: use upd, new and delete msgTypes
				r.Lock()
				isDeleted := r.Meta().IsDeleted()
				new := r.Meta().Created == r.Meta().Modified
				r.Unlock()
				switch {
				case isDeleted:
					api.send(opID, dbMsgTypeDel, r.Key(), nil)
				case new:
					api.send(opID, dbMsgTypeNew, r.Key(), data)
				default:
					api.send(opID, dbMsgTypeUpd, r.Key(), data)
				}
			} else if sub.Err != nil {
				// sub feed ended
				api.send(opID, dbMsgTypeError, sub.Err.Error(), nil)
			}
		}
	}
}

func (api *DatabaseAPI) handleQsub(opID []byte, queryText string) {
	// 127|qsub|<query>
	//    127|ok|<key>|<data>
	//    127|done
	//    127|error|<message>
	//    127|upd|<key>|<data>
	//    127|new|<key>|<data>
	//    127|delete|<key>
	//    127|warning|<message> // error with single record, operation continues

	var err error

	q, err := query.ParseQuery(queryText)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}

	sub, ok := api.registerSub(opID, q)
	if !ok {
		return
	}
	ok = api.processQuery(opID, q)
	if !ok {
		return
	}
	api.processSub(opID, sub)
}

func (api *DatabaseAPI) handlePut(opID []byte, key string, data []byte, create bool) {
	// 128|create|<key>|<data>
	//    128|success
	//    128|error|<message>

	// 129|update|<key>|<data>
	//    129|success
	//    129|error|<message>

	if len(data) < 2 {
		api.send(opID, dbMsgTypeError, "bad request: malformed message", nil)
		return
	}

	// FIXME: remove transition code
	if data[0] != record.JSON {
		typedData := make([]byte, len(data)+1)
		typedData[0] = record.JSON
		copy(typedData[1:], data)
		data = typedData
	}

	r, err := record.NewWrapper(key, nil, data[0], data[1:])
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}

	if create {
		err = api.db.PutNew(r)
	} else {
		err = api.db.Put(r)
	}
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}
	api.send(opID, dbMsgTypeSuccess, emptyString, nil)
}

func (api *DatabaseAPI) handleInsert(opID []byte, key string, data []byte) {
	// 130|insert|<key>|<data>
	//    130|success
	//    130|error|<message>

	r, err := api.db.Get(key)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}

	acc := r.GetAccessor(r)

	result := gjson.ParseBytes(data)
	anythingPresent := false
	var insertError error
	result.ForEach(func(key gjson.Result, value gjson.Result) bool {
		anythingPresent = true
		if !key.Exists() {
			insertError = errors.New("values must be in a map")
			return false
		}
		if key.Type != gjson.String {
			insertError = errors.New("keys must be strings")
			return false
		}
		if !value.Exists() {
			insertError = errors.New("non-existent value")
			return false
		}
		insertError = acc.Set(key.String(), value.Value())
		return insertError == nil
	})

	if insertError != nil {
		api.send(opID, dbMsgTypeError, insertError.Error(), nil)
		return
	}
	if !anythingPresent {
		api.send(opID, dbMsgTypeError, "could not find any valid values", nil)
		return
	}

	err = api.db.Put(r)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}

	api.send(opID, dbMsgTypeSuccess, emptyString, nil)
}

func (api *DatabaseAPI) handleDelete(opID []byte, key string) {
	// 131|delete|<key>
	//    131|success
	//    131|error|<message>

	err := api.db.Delete(key)
	if err != nil {
		api.send(opID, dbMsgTypeError, err.Error(), nil)
		return
	}
	api.send(opID, dbMsgTypeSuccess, emptyString, nil)
}

func (api *DatabaseAPI) shutdown() {
	if api.shuttingDown.SetToIf(false, true) {
		close(api.shutdownSignal)
		api.conn.Close()
	}
}
