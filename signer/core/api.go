// Copyright 2018 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/accounts/usbwallet"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/internal/ethapi"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"mime"
)

// ExternalAPI defines the external API through which signing requests are made.
type ExternalAPI interface {
	// List available accounts
	List(ctx context.Context) ([]common.Address, error)
	// New request to create a new account
	New(ctx context.Context) (accounts.Account, error)
	// SignTransaction request to sign the specified transaction
	SignTransaction(ctx context.Context, args SendTxArgs, methodSelector *string) (*ethapi.SignTransactionResult, error)
	// Sign - request to sign the given data (plus prefix)
	SignData(ctx context.Context, contentType string, addr common.MixedcaseAddress, data hexutil.Bytes) (hexutil.Bytes, error)
	// EcRecover - request to perform ecrecover
	EcRecover(ctx context.Context, data, sig hexutil.Bytes) (common.Address, error)
	// Export - request to export an account
	Export(ctx context.Context, addr common.Address) (json.RawMessage, error)
	// Import - request to import an account
	Import(ctx context.Context, keyJSON json.RawMessage) (Account, error)
}

// SignerUI specifies what method a UI needs to implement to be able to be used as a UI for the signer
type SignerUI interface {
	// ApproveTx prompt the user for confirmation to request to sign Transaction
	ApproveTx(request *SignTxRequest) (SignTxResponse, error)
	// ApproveSignData prompt the user for confirmation to request to sign data
	ApproveSignData(request *SignDataRequest) (SignDataResponse, error)
	// ApproveExport prompt the user for confirmation to export encrypted Account json
	ApproveExport(request *ExportRequest) (ExportResponse, error)
	// ApproveImport prompt the user for confirmation to import Account json
	ApproveImport(request *ImportRequest) (ImportResponse, error)
	// ApproveListing prompt the user for confirmation to list accounts
	// the list of accounts to list can be modified by the UI
	ApproveListing(request *ListAccountsRequest) (ListAccountsResponse, error)
	// ApproveNewAccount prompt the user for confirmation to create new Account, and reveal to caller
	ApproveNewAccount(request *NewAccountRequest) (NewAccountResponse, error)
	// ShowError displays error message to user
	ShowError(message string)
	// ShowInfo displays info message to user
	ShowInfo(message string)
	// OnApprovedTx notifies the UI about a transaction having been successfully signed.
	// This method can be used by a UI to keep track of e.g. how much has been sent to a particular recipient.
	OnApprovedTx(tx ethapi.SignTransactionResult)
	// OnSignerStartup is invoked when the signer boots, and tells the UI info about external API location and version
	// information
	OnSignerStartup(info StartupInfo)
}

// SignerAPI defines the actual implementation of ExternalAPI
type SignerAPI struct {
	chainID   *big.Int
	am        *accounts.Manager
	UI        SignerUI
	validator *Validator
}

// Metadata about a request
type Metadata struct {
	Remote    string `json:"remote"`
	Local     string `json:"local"`
	Scheme    string `json:"scheme"`
	UserAgent string `json:"User-Agent"`
	Origin    string `json:"Origin"`
}

// MetadataFromContext extracts Metadata from a given context.Context
func MetadataFromContext(ctx context.Context) Metadata {
	m := Metadata{"NA", "NA", "NA", "", ""} // batman

	if v := ctx.Value("remote"); v != nil {
		m.Remote = v.(string)
	}
	if v := ctx.Value("scheme"); v != nil {
		m.Scheme = v.(string)
	}
	if v := ctx.Value("local"); v != nil {
		m.Local = v.(string)
	}
	if v := ctx.Value("Origin"); v != nil {
		m.Origin = v.(string)
	}
	if v := ctx.Value("User-Agent"); v != nil {
		m.UserAgent = v.(string)
	}
	return m
}

// String implements Stringer interface
func (m Metadata) String() string {
	s, err := json.Marshal(m)
	if err == nil {
		return string(s)
	}
	return err.Error()
}

// types for the requests/response types between signer and UI
type (
	// SignTxRequest contains info about a Transaction to sign
	SignTxRequest struct {
		Transaction SendTxArgs       `json:"transaction"`
		Callinfo    []ValidationInfo `json:"call_info"`
		Meta        Metadata         `json:"meta"`
	}
	// SignTxResponse result from SignTxRequest
	SignTxResponse struct {
		//The UI may make changes to the TX
		Transaction SendTxArgs `json:"transaction"`
		Approved    bool       `json:"approved"`
		Password    string     `json:"password"`
	}
	// ExportRequest info about query to export accounts
	ExportRequest struct {
		Address common.Address `json:"address"`
		Meta    Metadata       `json:"meta"`
	}
	// ExportResponse response to export-request
	ExportResponse struct {
		Approved bool `json:"approved"`
	}
	// ImportRequest info about request to import an Account
	ImportRequest struct {
		Meta Metadata `json:"meta"`
	}
	ImportResponse struct {
		Approved    bool   `json:"approved"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	SignDataRequest struct {
		Address     common.MixedcaseAddress `json:"address"`
		Rawdata     hexutil.Bytes           `json:"raw_data"`
		Message     string                  `json:"message"`
		Hash        hexutil.Bytes           `json:"hash"`
		Meta        Metadata                `json:"meta"`
		ContentType string                  `json:"content_type"`
	}
	SignDataResponse struct {
		Approved bool `json:"approved"`
		Password string
	}
	NewAccountRequest struct {
		Meta Metadata `json:"meta"`
	}
	NewAccountResponse struct {
		Approved bool   `json:"approved"`
		Password string `json:"password"`
	}
	ListAccountsRequest struct {
		Accounts Accounts `json:"accounts"`
		Meta     Metadata `json:"meta"`
	}
	ListAccountsResponse struct {
		Accounts Accounts `json:"accounts"`
	}
	Message struct {
		Text string `json:"text"`
	}
	StartupInfo struct {
		Info map[string]interface{} `json:"info"`
	}
)

var ErrRequestDenied = errors.New("Request denied")

// NewSignerAPI creates a new API that can be used for Account management.
// ksLocation specifies the directory where to store the password protected private
// key that is generated when a new Account is created.
// noUSB disables USB support that is required to support hardware devices such as
// ledger and trezor.
func NewSignerAPI(chainID int64, ksLocation string, noUSB bool, ui SignerUI, abidb *AbiDb, lightKDF bool) *SignerAPI {
	var (
		backends []accounts.Backend
		n, p     = keystore.StandardScryptN, keystore.StandardScryptP
	)
	if lightKDF {
		n, p = keystore.LightScryptN, keystore.LightScryptP
	}
	// support password based accounts
	if len(ksLocation) > 0 {
		backends = append(backends, keystore.NewKeyStore(ksLocation, n, p))
	}
	if !noUSB {
		// Start a USB hub for Ledger hardware wallets
		if ledgerhub, err := usbwallet.NewLedgerHub(); err != nil {
			log.Warn(fmt.Sprintf("Failed to start Ledger hub, disabling: %v", err))
		} else {
			backends = append(backends, ledgerhub)
			log.Debug("Ledger support enabled")
		}
		// Start a USB hub for Trezor hardware wallets
		if trezorhub, err := usbwallet.NewTrezorHub(); err != nil {
			log.Warn(fmt.Sprintf("Failed to start Trezor hub, disabling: %v", err))
		} else {
			backends = append(backends, trezorhub)
			log.Debug("Trezor support enabled")
		}
	}
	return &SignerAPI{big.NewInt(chainID), accounts.NewManager(backends...), ui, NewValidator(abidb)}
}

// List returns the set of addresses this signer manages.
func (api *SignerAPI) List(ctx context.Context) ([]common.Address, error) {
	var accs []Account
	for _, wallet := range api.am.Wallets() {
		for _, acc := range wallet.Accounts() {
			acc := Account{Typ: "Account", URL: wallet.URL(), Address: acc.Address}
			accs = append(accs, acc)
		}
	}
	result, err := api.UI.ApproveListing(&ListAccountsRequest{Accounts: accs, Meta: MetadataFromContext(ctx)})
	if err != nil {
		return nil, err
	}
	if result.Accounts == nil {
		return nil, ErrRequestDenied

	}

	addresses := make([]common.Address, 0)
	for _, acc := range result.Accounts {
		addresses = append(addresses, acc.Address)
	}

	return addresses, nil
}

// New creates a new password protected Account. The private key is protected with
// the given password. Users are responsible to backup the private key that is stored
// in the keystore location thas was specified when this API was created.
func (api *SignerAPI) New(ctx context.Context) (accounts.Account, error) {
	be := api.am.Backends(keystore.KeyStoreType)
	if len(be) == 0 {
		return accounts.Account{}, errors.New("password based accounts not supported")
	}
	resp, err := api.UI.ApproveNewAccount(&NewAccountRequest{MetadataFromContext(ctx)})

	if err != nil {
		return accounts.Account{}, err
	}
	if !resp.Approved {
		return accounts.Account{}, ErrRequestDenied
	}
	return be[0].(*keystore.KeyStore).NewAccount(resp.Password)
}

// logDiff logs the difference between the incoming (original) transaction and the one returned from the signer.
// it also returns 'true' if the transaction was modified, to make it possible to configure the signer not to allow
// UI-modifications to requests
func logDiff(original *SignTxRequest, new *SignTxResponse) bool {
	modified := false
	if f0, f1 := original.Transaction.From, new.Transaction.From; !reflect.DeepEqual(f0, f1) {
		log.Info("Sender-account changed by UI", "was", f0, "is", f1)
		modified = true
	}
	if t0, t1 := original.Transaction.To, new.Transaction.To; !reflect.DeepEqual(t0, t1) {
		log.Info("Recipient-account changed by UI", "was", t0, "is", t1)
		modified = true
	}
	if g0, g1 := original.Transaction.Gas, new.Transaction.Gas; g0 != g1 {
		modified = true
		log.Info("Gas changed by UI", "was", g0, "is", g1)
	}
	if g0, g1 := big.Int(original.Transaction.GasPrice), big.Int(new.Transaction.GasPrice); g0.Cmp(&g1) != 0 {
		modified = true
		log.Info("GasPrice changed by UI", "was", g0, "is", g1)
	}
	if v0, v1 := big.Int(original.Transaction.Value), big.Int(new.Transaction.Value); v0.Cmp(&v1) != 0 {
		modified = true
		log.Info("Value changed by UI", "was", v0, "is", v1)
	}
	if d0, d1 := original.Transaction.Data, new.Transaction.Data; d0 != d1 {
		d0s := ""
		d1s := ""
		if d0 != nil {
			d0s = common.ToHex(*d0)
		}
		if d1 != nil {
			d1s = common.ToHex(*d1)
		}
		if d1s != d0s {
			modified = true
			log.Info("Data changed by UI", "was", d0s, "is", d1s)
		}
	}
	if n0, n1 := original.Transaction.Nonce, new.Transaction.Nonce; n0 != n1 {
		modified = true
		log.Info("Nonce changed by UI", "was", n0, "is", n1)
	}
	return modified
}

// SignTransaction signs the given Transaction and returns it both as json and rlp-encoded form
func (api *SignerAPI) SignTransaction(ctx context.Context, args SendTxArgs, methodSelector *string) (*ethapi.SignTransactionResult, error) {
	var (
		err    error
		result SignTxResponse
	)
	msgs, err := api.validator.ValidateTransaction(&args, methodSelector)
	if err != nil {
		return nil, err
	}

	req := SignTxRequest{
		Transaction: args,
		Meta:        MetadataFromContext(ctx),
		Callinfo:    msgs.Messages,
	}
	// Process approval
	result, err = api.UI.ApproveTx(&req)
	if err != nil {
		return nil, err
	}
	if !result.Approved {
		return nil, ErrRequestDenied
	}
	// Log changes made by the UI to the signing-request
	logDiff(&req, &result)
	var (
		acc    accounts.Account
		wallet accounts.Wallet
	)
	acc = accounts.Account{Address: result.Transaction.From.Address()}
	wallet, err = api.am.Find(acc)
	if err != nil {
		return nil, err
	}
	// Convert fields into a real transaction
	var unsignedTx = result.Transaction.toTransaction()

	// The one to sign is the one that was returned from the UI
	signedTx, err := wallet.SignTxWithPassphrase(acc, result.Password, unsignedTx, api.chainID)
	if err != nil {
		api.UI.ShowError(err.Error())
		return nil, err
	}

	rlpdata, err := rlp.EncodeToBytes(signedTx)
	response := ethapi.SignTransactionResult{Raw: rlpdata, Tx: signedTx}

	// Finally, send the signed tx to the UI
	api.UI.OnApprovedTx(response)
	// ...and to the external caller
	return &response, nil

}

// cliqueSigHash returns the hash which is used as input for the proof-of-authority
// signing. It is the hash of the entire header apart from the 65 byte signature
// contained at the end of the extra data.
//
// The method requires the extra data to be at least 65 bytes -- the original implementation
// in clique.go panics if this is the case, thus it's been reimplemented here to avoid the panic
// and simply return an error instead
func cliqueSigHash(header *types.Header) (hexutil.Bytes, error) {
	hash := common.Hash{}
	if len(header.Extra) < 65 {
		return hash.Bytes(), fmt.Errorf("clique header extradata too short, %d < 65", len(header.Extra))
	}
	hasher := sha3.NewKeccak256()
	rlp.Encode(hasher, []interface{}{
		header.ParentHash,
		header.UncleHash,
		header.Coinbase,
		header.Root,
		header.TxHash,
		header.ReceiptHash,
		header.Bloom,
		header.Difficulty,
		header.Number,
		header.GasLimit,
		header.GasUsed,
		header.Time,
		header.Extra[:len(header.Extra)-65],
		header.MixDigest,
		header.Nonce,
	})
	hasher.Sum(hash[:0])
	return hash.Bytes(), nil
}

func (api *SignerAPI) determineSignatureFormat(contentType string, data hexutil.Bytes) (*SignDataRequest, error) {
	var req *SignDataRequest
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, err
	}
	switch mediaType {
	case "application/clique":
		header := &types.Header{}
		if err := rlp.DecodeBytes(data, header); err != nil {
			return nil, err
		}
		sighash, err := cliqueSigHash(header)
		if err != nil {
			return nil, err
		}
		msg := fmt.Sprintf("Clique block %d [0x%x]", header.Number, header.Hash())
		req = &SignDataRequest{Rawdata: data, Message: msg, Hash: sighash, ContentType: mediaType}
	case "text/plain":
		// Sign calculates an Ethereum ECDSA signature for:
		// keccack256("\x19Ethereum Signed Message:\n" + len(message) + message))

		// In the cases where it matter ensure that the charset is handled. The charset
		// resides in the 'params' returned as the second returnvalue from mime.ParseMediaType
		// charset, ok := params["charset"]
		// As it is now, we accept any charset and just treat it as 'raw'.

		sighash, msg := SignHash(data)
		req = &SignDataRequest{Rawdata: data, Message: msg, Hash: sighash, ContentType: mediaType}
	default:
		//TODO! Add a content-type for EIP712 typed data.
		return nil, fmt.Errorf("content type '%s' not implemented for signing")
	}
	return req, nil

}

// SignData signs the hash of the provided data, but does so differently
// depending on the content-type specified.
//
// Depending on the content-type, different types of validations will occur.
//
// Note, the produced signature conforms to the secp256k1 curve R, S and V values,
// where the V value will be 27 or 28 for legacy reasons.
func (api *SignerAPI) SignData(ctx context.Context, contentType string, addr common.MixedcaseAddress, data hexutil.Bytes) (hexutil.Bytes, error) {

	var req, err = api.determineSignatureFormat(contentType, data)
	if err != nil {
		return nil, err
	}
	req.Address = addr
	req.Meta = MetadataFromContext(ctx)

	// We make the request prior to looking up if we actually have the account, to prevent
	// account-enumeration via the API
	res, err := api.UI.ApproveSignData(req)
	if err != nil {
		return nil, err
	}
	if !res.Approved {
		return nil, ErrRequestDenied
	}
	// Look up the wallet containing the requested signer
	account := accounts.Account{Address: addr.Address()}
	wallet, err := api.am.Find(account)
	if err != nil {
		return nil, err
	}
	// Sign the data with the wallet
	signature, err := wallet.SignHashWithPassphrase(account, res.Password, req.Hash)
	if err != nil {
		api.UI.ShowError(err.Error())
		return nil, err
	}
	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	return signature, nil
}

// EcRecover returns the address for the Account that was used to create the signature.
// Note, this function is compatible with eth_sign and personal_sign. As such it recovers
// the address of:
// hash = keccak256("\x19Ethereum Signed Message:\n"${message length}${message})
// addr = ecrecover(hash, signature)
//
// Note, the signature must conform to the secp256k1 curve R, S and V values, where
// the V value must be be 27 or 28 for legacy reasons.
//
// https://github.com/ethereum/go-ethereum/wiki/Management-APIs#personal_ecRecover
func (api *SignerAPI) EcRecover(ctx context.Context, data, sig hexutil.Bytes) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes long")
	}
	if sig[64] != 27 && sig[64] != 28 {
		return common.Address{}, fmt.Errorf("invalid Ethereum signature (V is not 27 or 28)")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1
	hash, _ := SignHash(data)
	rpk, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return common.Address{}, err
	}
	return crypto.PubkeyToAddress(*rpk), nil
}

// SignHash is a helper function that calculates a hash for the given message that can be
// safely used to calculate a signature from.
//
// The hash is calculated as
//   keccak256("\x19Ethereum Signed Message:\n"${message length}${message}).
//
// This gives context to the signed message and prevents signing of transactions.
func SignHash(data []byte) ([]byte, string) {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg)), msg
}

// Export returns encrypted private key associated with the given address in web3 keystore format.
func (api *SignerAPI) Export(ctx context.Context, addr common.Address) (json.RawMessage, error) {
	res, err := api.UI.ApproveExport(&ExportRequest{Address: addr, Meta: MetadataFromContext(ctx)})

	if err != nil {
		return nil, err
	}
	if !res.Approved {
		return nil, ErrRequestDenied
	}
	// Look up the wallet containing the requested signer
	wallet, err := api.am.Find(accounts.Account{Address: addr})
	if err != nil {
		return nil, err
	}
	if wallet.URL().Scheme != keystore.KeyStoreScheme {
		return nil, fmt.Errorf("Account is not a keystore-account")
	}
	return ioutil.ReadFile(wallet.URL().Path)
}

// Import tries to import the given keyJSON in the local keystore. The keyJSON data is expected to be
// in web3 keystore format. It will decrypt the keyJSON with the given passphrase and on successful
// decryption it will encrypt the key with the given newPassphrase and store it in the keystore.
func (api *SignerAPI) Import(ctx context.Context, keyJSON json.RawMessage) (Account, error) {
	be := api.am.Backends(keystore.KeyStoreType)

	if len(be) == 0 {
		return Account{}, errors.New("password based accounts not supported")
	}
	res, err := api.UI.ApproveImport(&ImportRequest{Meta: MetadataFromContext(ctx)})

	if err != nil {
		return Account{}, err
	}
	if !res.Approved {
		return Account{}, ErrRequestDenied
	}
	acc, err := be[0].(*keystore.KeyStore).Import(keyJSON, res.OldPassword, res.NewPassword)
	if err != nil {
		api.UI.ShowError(err.Error())
		return Account{}, err
	}
	return Account{Typ: "Account", URL: acc.URL, Address: acc.Address}, nil
}
