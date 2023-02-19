/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"bytes"
	"encoding/json"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

func main() {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
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
		"..",
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

	contract := network.GetContract("carcc")
	
	var option int

	for {

		fmt.Println("Izaberite opciju:")
		fmt.Println("0 - Initialize ledger")
		fmt.Println("1 - Get person by id")
		fmt.Println("2 - Get car by id")
		fmt.Println("3 - Get cars by color")
		fmt.Println("4 - Get cars by color and owner")
		fmt.Println("5 - Change car color")
		fmt.Println("6 - Repari car")
		fmt.Println("7 - Add car malfunction")
		fmt.Println("8 - Buy car")
		fmt.Println("9 - Exit")

		fmt.Scanf("%d", &option)

		switch option {
		case 0:
			_, err := contract.SubmitTransaction("InitLedger")
			if err != nil {
				fmt.Printf("Failed to init ledger!")
				os.Exit(1)
			}
		case 1:
			result, err := contract.EvaluateTransaction("GetPerson", "2")
			if err != nil {
				fmt.Printf("Failed to get person by id!")
				os.Exit(1)
			}
			fmt.Println(string(result))
		case 2:
			result, err := contract.EvaluateTransaction("GetCar", "c2")
			if err != nil {
				fmt.Printf("Failed to get car by id!")
				os.Exit(1)
			}
			fmt.Println(string(result))
		case 3:
			result, err := contract.EvaluateTransaction("GetCarsByColor", "black")
			if err != nil {
				fmt.Printf("Failed to get cars by color!")
				os.Exit(1)
			}
			resultJson := formatJSON(result)
			fmt.Printf("%s\n", resultJson)
		case 4:
			result, err := contract.EvaluateTransaction("GetCarsByOwnerAndColor", "2", "white")
			if err != nil {
				fmt.Printf("Failed to get cars by color and owner!")
				os.Exit(1)
			}
			resultJson := formatJSON(result)
			fmt.Printf("%s\n", resultJson)
		case 5:
			result, err := contract.SubmitTransaction("ChangeColor", "c6", "yellow")
			if err != nil {
				fmt.Printf("Failed to change car color!")
				os.Exit(1)
			}
			fmt.Println(string(result))
			result, err = contract.EvaluateTransaction("GetCar", "c6")
			if err != nil {
				fmt.Printf("Failed to get car by id!")
				os.Exit(1)
			}
			fmt.Println(string(result))
		case 6:
			result, err := contract.SubmitTransaction("RepairCar", "c2")
			if err != nil {
				fmt.Printf("Failed to repair car!")
				os.Exit(1)
			}
			fmt.Println(string(result))
			result, err = contract.EvaluateTransaction("GetPerson", "1")
			if err != nil {
				fmt.Printf("Failed to get person!")
				os.Exit(1)
			}
			fmt.Println(string(result))
			result, err = contract.EvaluateTransaction("GetCar", "c2")
			if err != nil {
				fmt.Printf("Failed to get car by id!")
				os.Exit(1)
			}
			fmt.Println(string(result))
		case 7:
			result, err := contract.SubmitTransaction("AddNewMalfunction", "c2", "novi kvar", "10.0")
			if err != nil {
				fmt.Printf("Failed to add malfunction!")
				os.Exit(1)
			}
			fmt.Println(string(result))
		case 8:
			result, err := contract.SubmitTransaction("BuyCar", "c3", "3", "false")
			if err != nil {
				fmt.Printf("Failed to add malfunction!")
				os.Exit(1)
			}
			fmt.Println(string(result))
			fmt.Println(string(result))
			result, err = contract.EvaluateTransaction("GetCar", "c2")
			if err != nil {
				fmt.Printf("Failed to get car by id!")
				os.Exit(1)
			}
			fmt.Println(string(result))
		case 9:
			fmt.Println("Exited.")
			break
		default:
			fmt.Println("Izabrana je nepostojeca opcija. Pokusajte ponovo.")
		}
	}
}

func formatJSON(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, " ", ""); err != nil {
		panic(fmt.Errorf("failed to parse JSON: %w", err))
	}
	return prettyJSON.String()
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"..",
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
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
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
