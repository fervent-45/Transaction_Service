package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
)

type BlockStatus int

const (
	committed BlockStatus = iota
	pending
)

type Transaction struct {
	blockNumber   int         `json:blockNumber`
	txnID         []string    `json:txnID`
	valid         bool        `json:valid`
	Timestamp     time.Time   `json:Timestamp`
	blockStatus   BlockStatus `json:blockStatus`
	prevBlockHash string      `json:prevBlockHash`
}

type blockStruct interface {
	pushValidTransaction() map[string]Pair //here KEY="SIM1" and Value = {val:1,version:1.0}
	updateBlockStatus() BlockStatus
}

type Pair struct {
	Value   int64   `json:"val"`
	Version float64 `json:"ver"`
}

func main() {
	InputTxns := make(map[string]json.RawMessage)
	data1 := `{"val": 2, "ver": 1.0}`
	data2 := `{"val": 3, "ver": 1.0}`
	data3 := `{"val": 4, "ver": 2.0}`

	// var pair1 Pair
	// var pair2 Pair
	// var pair3 Pair
	// json.Unmarshal([]byte(data1), &pair1)
	// json.Unmarshal([]byte(data2), &pair2)
	// json.Unmarshal([]byte(data3), &pair3)
	InputTxns["SIM1"] = json.RawMessage(data1)
	InputTxns["SIM2"] = json.RawMessage(data2)
	InputTxns["SIM3"] = json.RawMessage(data3)

	db, err := leveldb.OpenFile("path/to/db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for i := 1; i < 1001; i++ {
		key := fmt.Sprintf("SIM%d", i)
		data := fmt.Sprintf(`{"val": %d, "ver": 1.0}`, i)
		err = db.Put([]byte(key), []byte(data), nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	loop := db.NewIterator(nil, nil)

	for loop.Next() {
		key := string(loop.Key())
		val := loop.Value()
		fmt.Printf("%s - %s\n", key, val)
	}
	loop.Release()

	for i := 1; i <= 3; i++ {
		key := fmt.Sprintf("SIM%d", i)
		fmt.Printf("%s - %s\n", key, InputTxns[key])
	}

	txnlist := [3]string{"ed3462rfsg556", "gyghdgc565hgdg", "igdfgb656gvfg"}

	BlockStatus := pending
	BlockLedger := Transaction{blockNumber: 1, txnID: txnlist[:], valid: true, Timestamp: time.Now(), blockStatus: BlockStatus, prevBlockHash: "oxcfdg44"}
	fmt.Printf("%+v\n", BlockLedger)

}

// Task1
var blockCommitChannel = make(chan Block)

func calculateHash(txn *Transaction, wg *sync.WaitGroup) {
	defer wg.Done()
	// Simulate hash calculation
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d", txn.Amount)))
	txn.ID = hex.EncodeToString(h.Sum(nil))
	// Simulate transaction validity check
	txn.Valid = true
}

func processBlock(txns []Transaction, blockSize int) Block {
	var wg sync.WaitGroup
	for i := range txns {
		wg.Add(1)
		go calculateHash(&txns[i], &wg)
	}
	wg.Wait()

	return Block{Txns: txns, Commit: true}
}

func blockCommitter() {
	for block := range blockCommitChannel {
		if block.Commit {
			fmt.Println("Block committed with transactions:")
			for _, txn := range block.Txns {
				fmt.Printf("Txn ID: %s, Amount: %d, Valid: %t\n", txn.ID, txn.Amount, txn.Valid)
			}
		}
	}
}
