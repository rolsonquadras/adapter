/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package issuer

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite"
	"github.com/hyperledger/aries-framework-go/pkg/doc/signature/suite/ed25519signature2018"

	"github.com/hyperledger/aries-framework-go/pkg/doc/verifiable"
	mockdiddoc "github.com/hyperledger/aries-framework-go/pkg/mock/diddoc"
	mockvdri "github.com/hyperledger/aries-framework-go/pkg/mock/vdri"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/edge-adapter/pkg/internal/common/adapterutil"
	adaptervc "github.com/trustbloc/edge-adapter/pkg/vc"
)

const (
	id              = "e9e0f944-7b74-4298-9f3e-00ca609d6266"
	inviteeDID      = "did:example:7b744298e9e0f"
	inviterDID      = "agc"
	inviterLabel    = "user-agent"
	threadID        = "92d01dp5251sien42ad4dq4q2"
	connectionState = "completed"

	vc = `{
	   "@context":[
		  "https://www.w3.org/2018/credentials/v1",
		  "https://www.w3.org/2018/credentials/examples/v1"
	   ],
	   "id":"http://example.edu/credentials/1872",
	   "type":[
		  "VerifiableCredential",
		  "DIDConnection"
	   ],
	   "credentialSubject":{
		  "id":"` + id + `",
		  "inviteeDID":"` + inviteeDID + `",
		  "inviterDID":"` + inviterDID + `",
		  "inviterLabel":"` + inviterLabel + `",
		  "threadID":"` + threadID + `",
		  "connectionState": "` + connectionState + `"
	   },
	   "issuer":"did:example:76e12ec712ebc6f1c221ebfeb1f",
	   "issuanceDate":"2010-01-01T19:23:24Z"
	}`
)

func TestCreateManifestCredential(t *testing.T) {
	t.Run("test create manifest credential", func(t *testing.T) {
		issuerName := "TestIssuer"
		contexts := []string{"abc", "xyz"}

		vcBytes, err := CreateManifestCredential(issuerName, contexts)
		require.NoError(t, err)

		vc, err := verifiable.ParseCredential(vcBytes)
		require.NoError(t, err)
		require.True(t, adapterutil.StringsContains(ManifestCredentialType, vc.Types))

		manifestVC := &ManifestCredential{}

		err = adapterutil.DecodeJSONMarshaller(vc, manifestVC)
		require.NoError(t, err)
		require.Equal(t, contexts, manifestVC.Subject.Contexts)

		customFields := vc.CustomFields
		require.Equal(t, issuerName, customFields["name"])
	})
}

func TestParseWalletResponse(t *testing.T) {
	vdri := &mockvdri.MockVDRIRegistry{
		ResolveValue: mockdiddoc.GetMockDIDDoc(),
	}

	t.Run("test parse wallet - success", func(t *testing.T) {
		conn, err := ParseWalletResponse(getTestVP(t), vdri)
		require.NoError(t, err)
		require.NotNil(t, conn)

		require.Equal(t, conn.ID, id)
		require.Equal(t, conn.InviteeDID, inviteeDID)
		require.Equal(t, conn.InviterDID, inviterDID)
		require.Equal(t, conn.InviterLabel, inviterLabel)
		require.Equal(t, conn.ConnectionState, connectionState)
		require.Equal(t, conn.ThreadID, threadID)
	})

	t.Run("test parse wallet - invalid vp", func(t *testing.T) {
		conn, err := ParseWalletResponse([]byte("invalid json"), vdri)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid presentation")
		require.Nil(t, conn)
	})

	t.Run("test parse wallet - no credentials inside vp", func(t *testing.T) {
		vp := verifiable.Presentation{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Type:    []string{"VerifiablePresentation"},
		}
		vpJSON, err := vp.MarshalJSON()
		require.NoError(t, err)

		conn, err := ParseWalletResponse(vpJSON, vdri)
		require.Error(t, err)
		require.Contains(t, err.Error(), "there should be only one credential")
		require.Nil(t, conn)
	})

	t.Run("test parse wallet - invalid credential inside vp", func(t *testing.T) {
		vc := verifiable.Credential{
			Context: []string{"https://www.w3.org/2018/credentials/v1"},
			Types:   []string{"VerifiablePresentation"},
		}

		vp, err := vc.Presentation()
		require.NoError(t, err)

		vpJSON, err := vp.MarshalJSON()
		require.NoError(t, err)

		conn, err := ParseWalletResponse(vpJSON, vdri)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse credential")
		require.Nil(t, conn)
	})

	t.Run("test parse wallet - no credential of DIDConnectCredential type inside vp", func(t *testing.T) {
		vc, err := verifiable.ParseCredential([]byte(vc))
		require.NoError(t, err)

		vc.Types = []string{"VerifiableCredential"}

		vp, err := vc.Presentation()
		require.NoError(t, err)

		vpJSON, err := vp.MarshalJSON()
		require.NoError(t, err)

		conn, err := ParseWalletResponse(vpJSON, vdri)
		require.Error(t, err)
		require.Contains(t, err.Error(), "vc doesn't contain DIDConnection type")
		require.Nil(t, conn)
	})
}

func TestCreateConsentCredential(t *testing.T) {
	t.Run("test create didcomm init credential", func(t *testing.T) {
		didDocument := mockdiddoc.GetMockDIDDoc()

		didDocJSON, err := didDocument.JSONBytes()
		require.NoError(t, err)

		userDID := "did:example:abc789"

		rpDIDDoc := &adaptervc.DIDDoc{
			ID:  didDocument.ID,
			Doc: didDocJSON,
		}

		vc := CreateConsentCredential(didDocument.ID, didDocJSON, rpDIDDoc, userDID)
		require.True(t, adapterutil.StringsContains(adaptervc.ConsentCredentialType, vc.Types))

		consentVC := &adaptervc.ConsentCredential{}

		err = adapterutil.DecodeJSONMarshaller(vc, consentVC)
		require.NoError(t, err)
		require.Equal(t, didDocument.ID, consentVC.Subject.IssuerDIDDoc.ID)
		require.Equal(t, string(didDocJSON), string(consentVC.Subject.IssuerDIDDoc.Doc))
		require.Equal(t, rpDIDDoc.ID, consentVC.Subject.RPDIDDoc.ID)
		require.Equal(t, string(didDocJSON), string(consentVC.Subject.RPDIDDoc.Doc))
		require.Equal(t, userDID, consentVC.Subject.UserDID)
	})
}

func TestCreatePresentation(t *testing.T) {
	t.Run("test create presentation", func(t *testing.T) {
		vp, err := CreatePresentation(&verifiable.Credential{})
		require.NoError(t, err)
		require.NotNil(t, vp)
	})
}

func getTestVP(t *testing.T) []byte {
	vc, err := verifiable.ParseCredential([]byte(vc))
	require.NoError(t, err)

	vp, err := vc.Presentation()
	require.NoError(t, err)

	addLDProof(t, vp)

	vpJSON, err := vp.MarshalJSON()
	require.NoError(t, err)

	return vpJSON
}

func addLDProof(t *testing.T, vp *verifiable.Presentation) {
	t.Helper()

	_, secretKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	now := time.Now()
	err = vp.AddLinkedDataProof(&verifiable.LinkedDataProofContext{
		VerificationMethod:      "did:example:123",
		SignatureRepresentation: verifiable.SignatureJWS,
		SignatureType:           "Ed25519Signature2018",
		Suite:                   ed25519signature2018.New(suite.WithSigner(&testSigner{privKey: secretKey})),
		Created:                 &now,
		Domain:                  "user.example.com",
		Challenge:               uuid.New().String(),
		Purpose:                 "authentication",
	})
	require.NoError(t, err)
}

type testSigner struct {
	privKey []byte
}

func (t *testSigner) Sign(plaintext []byte) ([]byte, error) {
	return ed25519.Sign(t.privKey, plaintext), nil
}
