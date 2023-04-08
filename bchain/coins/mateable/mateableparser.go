package btc

import (
	"encoding/json"
	"math/big"

	"github.com/martinboehm/btcutil/chaincfg"
	"github.com/trezor/blockbook/bchain"
	"github.com/trezor/blockbook/common"
)

// network constants
const (
        MainnetMagic wire.MateableNet = 0x5d516455
        TestnetMagic wire.MateableNet = 0x0b110907
)

// parser parameters
var (
        MainNetParams chaincfg.Params
        TestNetParams chaincfg.Params
)

func init() {
        MainNetParams = chaincfg.MainNetParams
        MainNetParams.Net = MainnetMagic
        MainNetParams.PubKeyHashAddrID = []byte{51}
        MainNetParams.ScriptHashAddrID = []byte{128}
        MainNetParams.PrivateKeyID = []byte{170}
        MainNetParams.Bech32HRPSegwit = "MTBC"

        TestNetParams = chaincfg.TestNetParams
        TestNetParams.Net = TestnetMagic
        TestNetParams.PubKeyHashAddrID = []byte{51}
        TestNetParams.ScriptHashAddrID = []byte{128}
        TestNetParams.PrivateKeyID = []byte{170}
        TestNetParams.Bech32HRPSegwit = "tMTBC"
}

// MateableParser handle
type MateableParser struct {
	*BitcoinLikeParser
}

// NewMateableParser returns new MateableParser instance
func NewMateableParser(params *chaincfg.Params, c *Configuration) *MateableParser {
	p := &MateableParser{
		BitcoinLikeParser: NewBitcoinLikeParser(params, c),
	}
	p.VSizeSupport = true
	return p
}

// GetChainParams contains network parameters for the main Mateable network,
// the regression test Mateable network, the test Mateable network and
// the simulation test Mateable network, in this order
func GetChainParams(chain string) *chaincfg.Params {
	if !chaincfg.IsRegistered(&chaincfg.MainNetParams) {
		chaincfg.RegisterMateableParams()
	}
	switch chain {
	case "test":
		return &chaincfg.TestNetParams
	default:
		return &chaincfg.MainNetParams
	}
	return &chaincfg.MainNetParams
}

// ScriptPubKey contains data about output script
type ScriptPubKey struct {
	// Asm       string   `json:"asm"`
	Hex string `json:"hex,omitempty"`
	// Type      string   `json:"type"`
	Addresses []string `json:"addresses"` // removed from Mateabled 22.0.0
	Address   string   `json:"address"`   // used in Mateabled 22.0.0
}

// Vout contains data about tx output
type Vout struct {
	ValueSat     big.Int
	JsonValue    common.JSONNumber `json:"value"`
	N            uint32            `json:"n"`
	ScriptPubKey ScriptPubKey      `json:"scriptPubKey"`
}

// Tx is blockchain transaction
// unnecessary fields are commented out to avoid overhead
type Tx struct {
	Hex         string       `json:"hex"`
	Txid        string       `json:"txid"`
	Version     int32        `json:"version"`
	LockTime    uint32       `json:"locktime"`
	VSize       int64        `json:"vsize,omitempty"`
	Vin         []bchain.Vin `json:"vin"`
	Vout        []Vout       `json:"vout"`
	BlockHeight uint32       `json:"blockHeight,omitempty"`
	// BlockHash     string `json:"blockhash,omitempty"`
	Confirmations    uint32      `json:"confirmations,omitempty"`
	Time             int64       `json:"time,omitempty"`
	Blocktime        int64       `json:"blocktime,omitempty"`
	CoinSpecificData interface{} `json:"-"`
}

// ParseTxFromJson parses JSON message containing transaction and returns Tx struct
// Mateabled version 22.0.0 removed ScriptPubKey.Addresses from the API and replaced it by a single Address
func (p *MateableParser) ParseTxFromJson(msg json.RawMessage) (*bchain.Tx, error) {
	var bitcoinTx Tx
	var tx bchain.Tx
	err := json.Unmarshal(msg, &bitcoinTx)
	if err != nil {
		return nil, err
	}

	// it is necessary to copy bitcoinTx to Tx to make it compatible
	tx.Hex = bitcoinTx.Hex
	tx.Txid = bitcoinTx.Txid
	tx.Version = bitcoinTx.Version
	tx.LockTime = bitcoinTx.LockTime
	tx.VSize = bitcoinTx.VSize
	tx.Vin = bitcoinTx.Vin
	tx.BlockHeight = bitcoinTx.BlockHeight
	tx.Confirmations = bitcoinTx.Confirmations
	tx.Time = bitcoinTx.Time
	tx.Blocktime = bitcoinTx.Blocktime
	tx.CoinSpecificData = bitcoinTx.CoinSpecificData
	tx.Vout = make([]bchain.Vout, len(bitcoinTx.Vout))

	for i := range bitcoinTx.Vout {
		bitcoinVout := &bitcoinTx.Vout[i]
		vout := &tx.Vout[i]
		// convert vout.JsonValue to big.Int and clear it, it is only temporary value used for unmarshal
		vout.ValueSat, err = p.AmountToBigInt(bitcoinVout.JsonValue)
		if err != nil {
			return nil, err
		}
		vout.N = bitcoinVout.N
		vout.ScriptPubKey.Hex = bitcoinVout.ScriptPubKey.Hex
		// convert single Address to Addresses if Addresses are empty
		if len(bitcoinVout.ScriptPubKey.Addresses) == 0 {
			vout.ScriptPubKey.Addresses = []string{bitcoinVout.ScriptPubKey.Address}
		} else {
			vout.ScriptPubKey.Addresses = bitcoinVout.ScriptPubKey.Addresses
		}
	}

	return &tx, nil
}
