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

		fmt.Println("Choose an option:")
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
			}

		case 1:

			fmt.Printf("Enter person id: ")
			var id string
			fmt.Scanf("%s", &id)

			result, err := contract.EvaluateTransaction("GetPerson", id)
			if err != nil {
				fmt.Printf("Failed to get person by id!")
			}

			fmt.Println(string(result))

		case 2:

			fmt.Printf("Enter car id: ")
			var id string
			fmt.Scanf("%s", &id)

			result, err := contract.EvaluateTransaction("GetCar", id)
			if err != nil {
				fmt.Printf("Failed to get car by id!")
			}

			fmt.Println(string(result))

		case 3:

			fmt.Printf("Enter color: ")
			var color string
			fmt.Scanf("%s", &color)

			result, err := contract.EvaluateTransaction("GetCarsByColor", color)
			if err != nil {
				fmt.Printf("Failed to get cars by color!")
			}

			resultJson := formatJson(result)
			if len(resultJson) <= 2 {
				fmt.Printf("There are no %s colored cars.", color)
			} else {
				fmt.Printf("%s\n", resultJson)
			}

		case 4:

			fmt.Printf("Enter color: ")
			var color string
			fmt.Scanf("%s", &color)

			fmt.Printf("Enter owner id: ")
			var ownerId string
			fmt.Scanf("%s", &ownerId)

			result, err := contract.EvaluateTransaction("GetCarsByOwnerAndColor", ownerId, color)
			if err != nil {
				fmt.Printf("Failed to get cars by color and owner!")
			}

			resultJson := formatJson(result)
			fmt.Printf("%s\n", resultJson)

		case 5:

			fmt.Printf("Enter car id: ")
			var carId string
			fmt.Scanf("%s", &carId)

			fmt.Printf("Enter new color: ")
			var newColor string
			fmt.Scanf("%s", &newColor)

			result, err := contract.SubmitTransaction("ChangeColor", carId, newColor)
			if err != nil {
				fmt.Printf("Failed to change car color!")
			}

			fmt.Println(string(result))

		case 6:

			fmt.Printf("Enter car id: ")
			var carId string
			fmt.Scanf("%s", &carId)

			result, err := contract.SubmitTransaction("RepairCar", carId)
			if err != nil {
				fmt.Printf("Failed to repair car!")
			}

			fmt.Println(string(result))

		case 7:

			fmt.Printf("Enter car id: ")
			var carId string
			fmt.Scanf("%s", &carId)

			fmt.Printf("Enter malfunction description: ")
			var description string
			fmt.Scanf("%s", &description)

			fmt.Printf("Enter malfunction price: ")
			var price string
			fmt.Scanf("%s", &price)

			result, err := contract.SubmitTransaction("AddNewMalfunction", carId, description, price)
			if err != nil {
				fmt.Printf("Failed to add malfunction!")
			}

			fmt.Println(string(result))

		case 8:

			fmt.Printf("Enter id of the car you want to buy: ")
			var carId string
			fmt.Scanf("%s", &carId)

			fmt.Printf("Enter buyer id: ")
			var buyerId string
			fmt.Scanf("%s", &buyerId)

			fmt.Printf("Do you want to buy the car despite its malfunctions? (yes/no)")
			var answer string
			fmt.Scanf("%s", &answer)

			result, err := contract.SubmitTransaction("BuyCar", carId, buyerId, answer)
			if err != nil {
				fmt.Printf("Failed to buy car!")
			}

			fmt.Println(string(result))

		case 9:

			fmt.Println("End program.")
			os.Exit(1)

		default:

			fmt.Println("Chosen option does not exist! Please try again.")
		}
	}
}

func formatJson(data []byte) string {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, " ", ""); err != nil {
		fmt.Errorf("failed to parse json: %w", err)
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
