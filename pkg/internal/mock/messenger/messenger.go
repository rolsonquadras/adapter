/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package messenger

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
)

// MockMessenger mock messenger.
type MockMessenger struct {
	ReplyToFunc func(msgID string, msg service.DIDCommMsgMap) error
	SendFunc    func(msg service.DIDCommMsgMap, myDID, theirDID string) error
}

// ReplyTo reply to a message.
func (m *MockMessenger) ReplyTo(msgID string, msg service.DIDCommMsgMap) error {
	if m.ReplyToFunc != nil {
		return m.ReplyToFunc(msgID, msg)
	}

	return nil
}

// Send send message.
func (m *MockMessenger) Send(msg service.DIDCommMsgMap, myDID, theirDID string) error {
	if m.SendFunc != nil {
		return m.SendFunc(msg, myDID, theirDID)
	}

	return nil
}

// SendToDestination send mesage to destination.
func (m *MockMessenger) SendToDestination(_ service.DIDCommMsgMap, _ string, _ *service.Destination) error {
	return nil
}

// ReplyToNested reply to nested message.
func (m *MockMessenger) ReplyToNested(_ service.DIDCommMsgMap, _ *service.NestedReplyOpts) error {
	return nil
}
