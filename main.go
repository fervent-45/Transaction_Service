package main

import (
	"crypto/sha256"
	//"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/syndtr/goleveldb/leveldb"
)

type BlockStatus string

const (
	Committed BlockStatus = "Committed"
	Pending   BlockStatus = "Pending"
)

type Transaction struct {
	blockNumber      int         `json:blockNumber`
	txnID            string      `json:txnID`
	valid            bool        `json:valid`
	txns             JsonPair    `json:txns`
	Timestamp        time.Time   `json:Timestamp`
	blockStatus      BlockStatus `json:blockStatus`
	prevBlockHash    string      `json:prevBlockHash`
	currentBlockHash string      `json:currentBlockHash`
}

type Block interface {
	pushValidTransaction() Transaction
	updateBlockStatus() BlockStatus
}

type BlockLedger struct {
	Txns   []Transaction
	Commit bool
}

type Pair struct {
	Val int     `json:val`
	Ver float32 `json:ver`
}

type JsonPair struct {
	Val int     `json:val`
	Ver float32 `json:ver`
}

func (c *JsonPair) Increment() {
	c.Val++
	c.Ver++
}
var blockMap = map[int]string{}

func main() {
	db, err := leveldb.OpenFile("path/to/db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	blockSize := 5

	InputTxns := make(map[string]Pair)
	//Taking userInput of Input Transaction:

	for i := 1; i <= 3; i++ {

		var key string
		var val int
		var ver float32

		fmt.Print("Enter the Key of Pair Struct ")
		fmt.Scanf("%s\n",&key)
		fmt.Print("Enter the values having value and version of Pair Struct ")
		fmt.Scanf("%d %f\n",&val,&ver)

		InputTxns[key]=Pair{Val: val,Ver: ver}
	}

	//Leval DB Connection

	//---> Inserting data inside leveldb
	for i := 1; i < 1001; i++ {
		key := fmt.Sprintf("SIM%d", i)
		dbValue := JsonPair{Val: i, Ver: 1.0}
		// fmt.Println(dbValue)
		data, err := json.Marshal(dbValue)
		if err != nil {
			log.Fatal("Error Marshalling struct jsonPair:", err)
		}
		err = db.Put([]byte(key), data, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	//---> Getting the from the leveldb to compute the operation
	//Time.Start
	for i := 1; i < 4; i++ {
		key := fmt.Sprintf("SIM%d", i)
		data, err := db.Get([]byte(key), nil)
		isValidTxn := false
		if err != nil {
			log.Fatal("Error Getting data from db:", err)
		}
		var jp JsonPair
		err = json.Unmarshal(data, &jp)
		if err != nil {
			log.Fatal("Error unmarshalling data from db:", err)
		}
		//fmt.Printf("jp: %v\n", jp)
		var inputpair Pair
		inputpair = InputTxns[fmt.Sprintf("SIM%d", i)]
		fmt.Printf("input val: %d,input ver: %f\n", inputpair.Val, inputpair.Ver)
		fmt.Printf("leveldb val: %d, leveldb ver: %f\n", jp.Val, jp.Ver)
		if jp.Ver == inputpair.Ver {
			println("Version is Matching")
			jp.Increment()
			isValidTxn = true
		}
		//fmt.Printf("val: %d, ver: %f\n", jp.Val, jp.Ver)
		//--> Uncomment to see Task 4
		//fmt.Printf("block details by blockNumber: \n",fetchBlockDetailsByBlockNumber(1))
		//fmt.Println()
        
		// ---> Uncomment to see Task 5
		//fetchAllBlockDetails()
		computedTxns := jp

		txnLedger := BlockLedger{
			Txns:   []Transaction{},
			Commit: true,
		}

		txn_id := uuid.New()

		computedTransaction := []Transaction{
			{blockNumber: i, txnID: txn_id.String(), valid: isValidTxn, txns: computedTxns, Timestamp: time.Now(), blockStatus: Committed, prevBlockHash: "", currentBlockHash: ""},
		}
		txnLedger.Txns = append(txnLedger.Txns, computedTransaction...)

		go blockCommitter()

		block := processBlock(txnLedger.Txns, blockSize)

		// Send block to commit channel
		blockCommitChannel <- block

		time.Sleep(1 * time.Second) // Allow blockCommitter to finish

		//Time.END

	}
	

	

}

func writeBlockToFile(block string) {
	file, err := os.OpenFile("blocks.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	if _, err := file.Write(append([]byte(block), '\n')); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	defer file.Close()
	fmt.Println("Block Inserted into the file.")
}

// Task1 Task2 and Task3
var blockCommitChannel = make(chan BlockLedger)

var prevHash string = "Genesis Block no Hash"

func calculateHash(txn *Transaction, wg *sync.WaitGroup) {
	defer wg.Done()
	h := sha256.New()
	combinedString := fmt.Sprintf("%d %s %t %s %s %s", txn.blockNumber, txn.txnID, txn.valid, txn.Timestamp, txn.blockStatus, txn.prevBlockHash)
	h.Write([]byte(combinedString))
	hash := h.Sum(nil)
	txn.currentBlockHash = prevHash
	prevHash = hex.EncodeToString(hash)
}

func processBlock(txns []Transaction, blockSize int) BlockLedger {
	var wg sync.WaitGroup
	for i := range txns {
		wg.Add(1)
		go calculateHash(&txns[i], &wg)
	}
	wg.Wait()
	return BlockLedger{Txns: txns, Commit: true}
}

func blockCommitter() {
	for block := range blockCommitChannel {
		if block.Commit {
			fmt.Println("Block committed with transactions:")
			for _, txn := range block.Txns {
				txnKey := fmt.Sprintf("SIM%d", txn.blockNumber)
				txnData := fmt.Sprintf("{%q, : {\"val\": %d, \"ver\": %f, \"valid\": %t}}", txnKey, txn.txns.Val, txn.txns.Ver, txn.valid)
				blockData := fmt.Sprintf("Block Number: %d, TransactionId: %s,Txns: %s, Timestamp: %s, BlockStatus: %s, PrevHashBlock: %s\n", txn.blockNumber, txn.txnID, txnData, txn.Timestamp, txn.blockStatus, txn.currentBlockHash)
				fmt.Printf("Block Number: %d, TransactionId: %s,Txns: %s, Timestamp: %s, BlockStatus: %s, PrevHashBlock: %s\n", txn.blockNumber, txn.txnID, txnData, txn.Timestamp, txn.blockStatus, txn.currentBlockHash)
                //TASK 6
				start := time.Now()
				writeBlockToFile(blockData)
				blockProcessingTime := time.Since(start)
				blockMap[txn.blockNumber]=blockData
				fmt.Printf("block processing time for Block%d: %s\n",txn.blockNumber,blockProcessingTime)
			}
		}
		fmt.Println()
	}

}

//Task 4 and 5

func fetchBlockDetailsByBlockNumber(blockNumber int) string{
      blockData := blockMap[blockNumber]
	  return blockData
}
func fetchAllBlockDetails() {
	fmt.Println("All committed block data: ")
    for _,data := range blockMap{
		fmt.Println(data)
	}
}



//Task 6
//Calculate the block processing time
// we can display it by using main fuction
