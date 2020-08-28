package pusher

import (
	"encoding/json"
	"errors"
	"fmt"
)

type batchEvent struct {
	Channel  string  `json:"channel"`
	Name     string  `json:"name"`
	Data     string  `json:"data"`
	SocketID *string `json:"socket_id,omitempty"`
}
type batchPayload struct {
	Batch []batchEvent `json:"batch"`
}

type eventPayload struct {
	Name     string   `json:"name"`
	Channels []string `json:"channels"`
	Data     string   `json:"data"`
	SocketID *string  `json:"socket_id,omitempty"`
}

func encodeTriggerBody(channels []string, event string, data interface{}, socketID *string, encryptionKey []byte, maxEventPayloadSize int) ([]byte, error) {
	dataBytes, err := encodeEventData(data)
	if err != nil {
		return nil, err
	}
	var payloadData string
	if isEncryptedChannel(channels[0]) {
		payloadData = encrypt(channels[0], dataBytes, encryptionKey)
	} else {
		payloadData = string(dataBytes)
	}
	if len(payloadData) > maxEventPayloadSize {
		return nil, errors.New(fmt.Sprintf("Event payload exceeded maximum size (%d bytes is too much)", len(payloadData)))
	}
	return json.Marshal(&eventPayload{
		Name:     event,
		Channels: channels,
		Data:     payloadData,
		SocketID: socketID,
	})
}

func encodeTriggerBatchBody(batch []Event, encryptionKey []byte, maxEventPayloadSize int) ([]byte, error) {
	batchEvents := make([]batchEvent, len(batch))
	for idx, e := range batch {
		var stringifyedDataBytes string
		dataBytes, err := encodeEventData(e.Data)
		if err != nil {
			return nil, err
		}
		if isEncryptedChannel(e.Channel) {
			stringifyedDataBytes = encrypt(e.Channel, dataBytes, encryptionKey)
		} else {
			stringifyedDataBytes = string(dataBytes)
		}
		if len(stringifyedDataBytes) > maxEventPayloadSize {
			return nil, fmt.Errorf("Data of the event #%d in batch, exceeded maximum size (%d bytes is too much)", idx, len(stringifyedDataBytes))
		}
		newBatchEvent := batchEvent{
			Channel:  e.Channel,
			Name:     e.Name,
			Data:     stringifyedDataBytes,
			SocketID: e.SocketID,
		}
		batchEvents[idx] = newBatchEvent
	}
	return json.Marshal(&batchPayload{batchEvents})
}

func encodeEventData(data interface{}) ([]byte, error) {
	var dataBytes []byte
	var err error

	switch d := data.(type) {
	case []byte:
		dataBytes = d
	case string:
		dataBytes = []byte(d)
	default:
		dataBytes, err = json.Marshal(data)
		if err != nil {
			return nil, err
		}
	}
	return dataBytes, nil
}
