package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tensor-programming/golang-blockchain/blockchain"
)

type ServerSetup struct {
	Status string
}

type BlockSuccess struct {
	Message string
}

func checkServer(w http.ResponseWriter, r *http.Request) {

	var newEmployee = ServerSetup{Status: "Server is in running state"}
	setupHeader(w)
	json.NewEncoder(w).Encode(newEmployee)
}

func createBlockChain(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if blockchain.DBexists() {

		json.NewEncoder(w).Encode(BlockSuccess{Message: "Blockchain already exists"})
	} else {

		data := r.FormValue("data")
		amount := r.FormValue("amount")
		amountInt, err := strconv.Atoi(amount)
		blockchain.Handle(err)
		chain := blockchain.InitBlockChain(data, amountInt)
		chain.Database.Close()

		json.NewEncoder(w).Encode(BlockSuccess{Message: "Blockchain Created"})
	}
}

func sendTransaction(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if !blockchain.DBexists() {
		json.NewEncoder(w).Encode(BlockSuccess{Message: "No existing blockchain found, create one!"})
	} else {

		from := r.FormValue("from")
		to := r.FormValue("to")
		amount := r.FormValue("amount")
		amountInt, err := strconv.Atoi(amount)
		blockchain.Handle(err)

		chain := blockchain.ContinueBlockChain(from)
		defer chain.Database.Close()

		tx := blockchain.NewTransaction(from, to, amountInt, chain)
		chain.AddBlock([]*blockchain.Transaction{tx})

		json.NewEncoder(w).Encode(BlockSuccess{Message: "Send Transaction successfully"})

	}
}

func printChain(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if !blockchain.DBexists() {
		json.NewEncoder(w).Encode(BlockSuccess{Message: "No existing blockchain found, create one!"})
	} else {

		var tmpRecords []blockchain.Block
		chain := blockchain.ContinueBlockChain("")
		iter := chain.Iterator()

		for {
			block := iter.Next()
			pow := blockchain.NewProof(block)
			fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
			tmpRecords = append(tmpRecords, *block)
			if len(block.PrevHash) == 0 {
				break
			}
		}
		json.NewEncoder(w).Encode(tmpRecords)
		defer chain.Database.Close()
	}
}

func getBalance(w http.ResponseWriter, r *http.Request) {
	setupHeader(w)
	if !blockchain.DBexists() {
		json.NewEncoder(w).Encode(BlockSuccess{Message: "No existing blockchain found, create one!"})
	} else {
		from := mux.Vars(r)["from"]
		chain := blockchain.ContinueBlockChain(from)

		balance := 0
		UTXOs := chain.FindUTXO(from)

		for _, out := range UTXOs {
			balance += out.Value
		}
		defer chain.Database.Close()
		bal := fmt.Sprintf("%s%d", "Balance is ", balance)
		json.NewEncoder(w).Encode(BlockSuccess{Message: bal})
	}
}
func setupHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
}

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", checkServer)
	router.HandleFunc("/createBlockChain", createBlockChain).Methods("POST")
	router.HandleFunc("/sendTransaction", sendTransaction).Methods("POST")
	router.HandleFunc("/printChain", printChain).Methods("GET")
	router.HandleFunc("/getBalance/{from}", getBalance).Methods("GET")

	log.Fatal(http.ListenAndServe(":8081", router))
}
