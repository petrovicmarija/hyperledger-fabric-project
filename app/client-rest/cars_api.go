package main

import (
	"fmt"
	"errors"
	"path/filepath"
	"io/ioutil"
	"os"
	"net/http"
	"log"
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	// "github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/gorilla/mux"
)

func main() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")

	handleRequests()

}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!")
}

func initLedger(w http.ResponseWriter, r *http.Request) {
	contract := getContract()

	_, err := contract.SubmitTransaction("InitLedger")
	if err != nil {
		fmt.Fprintf(w, "Ledger is successfully initialized.")
	}
}

func getPersonById(w http.ResponseWriter, r *http.Request) {
	contract := getContract()

	vars := mux.Vars(r)
	personId := vars["id"]

	person, err := contract.EvaluateTransaction("GetPerson", personId)
	if err != nil {
		fmt.Fprintf(w, "Person with provided id does not exist!")
	}

	json.NewEncoder(w).Encode(person)
}

func getCarById(w http.ResponseWriter, r *http.Request) {
	contract := getContract()

	vars := mux.Vars(r)
	carId := vars["id"]

	car, err := contract.EvaluateTransaction("GetCar", carId)
	if err != nil {
		fmt.Fprintf(w, "Car with provided id does not exist!")
	}

	json.NewEncoder(w).Encode(car)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", helloWorld)
	myRouter.HandleFunc("/ledger", initLedger)
	myRouter.HandleFunc("/ledger/persons/{id}", getPersonById)
	myRouter.HandleFunc("/ledger/cars/{id}", getCarById)

	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func getContract() *gateway.Contract {
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("appUser") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	ccpPath := filepath.Join(
		// "..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org4.example.com",
		"connection-org4.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}

	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}

	// contract := network.GetContract("carcc")

	// return &contract

	return network.GetContract("carcc")
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		// "..",
		"..",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org4.example.com",
		"users",
		"User1@org4.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}

	if len(files) != 1 {
		return errors.New("Keystore folder should have contain one file!")
	}

	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org4MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}

	return nil
}