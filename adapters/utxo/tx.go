package utxo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	btcec "github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/txsort"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	escrow "github.com/mobazha/go-p2p-escrow"
)

// TxInput describes one input to spend from the escrow.
type TxInput struct {
	TxID   string
	Vout   uint32
	Amount int64
}

// TxOutput describes one output in the spending transaction.
type TxOutput struct {
	Address string
	Amount  int64
}

// BuildUnsignedTx constructs an unsigned wire.MsgTx from the given inputs and outputs.
// Outputs are BIP 69 sorted for determinism.
func BuildUnsignedTx(inputs []TxInput, outputs []TxOutput, cfg *ChainConfig) (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(cfg.TxVersion)

	for _, in := range inputs {
		op, err := deserializeOutpoint(in.TxID, in.Vout)
		if err != nil {
			return nil, fmt.Errorf("input %s:%d: %w", in.TxID, in.Vout, err)
		}
		tx.TxIn = append(tx.TxIn, wire.NewTxIn(op, nil, nil))
	}

	for _, out := range outputs {
		scriptPubKey, err := GetPayToAddrScript(out.Address, cfg.Params)
		if err != nil {
			return nil, fmt.Errorf("output addr %s: %w", out.Address, err)
		}
		tx.TxOut = append(tx.TxOut, wire.NewTxOut(out.Amount, scriptPubKey))
	}

	txsort.InPlaceSort(tx)
	return tx, nil
}

// SignWitness produces a witness signature for each input of a multisig tx.
func SignWitness(tx *wire.MsgTx, key *btcec.PrivateKey, redeemScript []byte, inputAmounts []int64) ([]escrow.Signature, error) {
	if len(tx.TxIn) != len(inputAmounts) {
		return nil, errors.New("input count does not match amount count")
	}

	sigs := make([]escrow.Signature, 0, len(tx.TxIn))
	for i := range tx.TxIn {
		prevFetcher := txscript.NewCannedPrevOutputFetcher(redeemScript, inputAmounts[i])
		sigHashes := txscript.NewTxSigHashes(tx, prevFetcher)

		sig, err := txscript.RawTxInWitnessSignature(tx, sigHashes, i, inputAmounts[i], redeemScript, txscript.SigHashAll, key)
		if err != nil {
			return nil, fmt.Errorf("sign input %d: %w", i, err)
		}
		// Strip trailing SigHashAll byte — it's appended during assembly.
		sigs = append(sigs, escrow.Signature{
			InputIndex: i,
			Data:       sig[:len(sig)-1],
		})
	}
	return sigs, nil
}

// AssembleWitnessMultisig builds the witness stack for a multisig spend.
// The witness layout for P2WSH multisig is: OP_0 <sig1> <sig2> [OP_TRUE if timelock IF branch] <redeemScript>
func AssembleWitnessMultisig(tx *wire.MsgTx, allSignatures [][]escrow.Signature, redeemScript []byte, timeLocked bool) error {
	for i := range tx.TxIn {
		witness := wire.TxWitness{[]byte{}} // leading OP_0 for CHECKMULTISIG bug
		for _, sigSet := range allSignatures {
			for _, sig := range sigSet {
				if sig.InputIndex == i {
					witness = append(witness, append(sig.Data, byte(txscript.SigHashAll)))
					break
				}
			}
		}
		if timeLocked {
			witness = append(witness, []byte{0x01}) // OP_TRUE → take IF branch
		}
		witness = append(witness, redeemScript)
		tx.TxIn[i].Witness = witness
	}
	return nil
}

// AssembleWitnessTimeout builds the witness for a timeout-branch spend.
// Witness: <sig> OP_0 <redeemScript> with sequence lock set on inputs.
func AssembleWitnessTimeout(tx *wire.MsgTx, key *btcec.PrivateKey, redeemScript []byte, inputAmounts []int64, sequenceLock uint32) error {
	for i := range tx.TxIn {
		tx.TxIn[i].Sequence = sequenceLock
	}

	for i := range tx.TxIn {
		prevFetcher := txscript.NewCannedPrevOutputFetcher(redeemScript, inputAmounts[i])
		sigHashes := txscript.NewTxSigHashes(tx, prevFetcher)

		sig, err := txscript.RawTxInWitnessSignature(tx, sigHashes, i, inputAmounts[i], redeemScript, txscript.SigHashAll, key)
		if err != nil {
			return fmt.Errorf("sign input %d: %w", i, err)
		}
		tx.TxIn[i].Witness = wire.TxWitness{sig, {}, redeemScript}
	}
	return nil
}

// SerializeTx serializes a fully-assembled transaction to raw bytes.
func SerializeTx(tx *wire.MsgTx) ([]byte, error) {
	var buf bytes.Buffer
	if err := tx.BtcEncode(&buf, wire.ProtocolVersion, wire.WitnessEncoding); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeOutpoint(txHash string, vout uint32) (*wire.OutPoint, error) {
	h, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil, err
	}
	return wire.NewOutPoint(h, vout), nil
}

// SerializeOutpoint encodes an outpoint as txid(32) + vout(4LE) bytes.
func SerializeOutpoint(txHash string, vout uint32) ([]byte, error) {
	h, err := chainhash.NewHashFromStr(txHash)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 36)
	copy(buf[:32], h[:])
	binary.LittleEndian.PutUint32(buf[32:], vout)
	return buf, nil
}
