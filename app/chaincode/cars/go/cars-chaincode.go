package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}


type Person struct {
	ID		string
	Name	string
	Surname	string
	Email	string
	Money	float32
}

type Car struct {
	ID				string
	Brand			string
	Model			string
	Year			int
	Color			string
	Owner			string
	Malfunctions	[]Malfunction
	Price			float32
}

type Malfunction struct {
	Description		string
	Price			float32
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	persons := []Person {
		{ ID: "1", Name: "Petar", Surname: "Petrovic", Email: "petar@gmail.com", Money: 7700.0 },
		{ ID: "2", Name: "Marko", Surname: "Markovic", Email: "marko@gmail.com", Money: 2850.0 },
		{ ID: "3", Name: "Stefan", Surname: "Stefanovic", Email: "stefan@gmail.com", Money: 5100.0 },
	}

	cars := []Car {
		{ ID: "c1", Brand: "Jeep", Model: "Renegade", Year: 2015, Color: "black", Owner: "1", Malfunctions: []Malfunction{
			{ Description: "Popravak motora", Price: 32.3 },
			{ Description: "Popravak brave na vratima", Price: 12.5 },
		}, Price: 5200.0 },
		{ ID: "c2", Brand: "Dacia", Model: "Duster", Year: 2019, Color: "gray", Owner: "1", Malfunctions: []Malfunction{
			{ Description: "Curenje ulja", Price: 23.5 },
		}, Price: 3900.0},
		{ ID: "c3", Brand: "Toyota", Model: "RAV4", Year: 2018, Color: "black", Owner: "2", Malfunctions: []Malfunction{}, Price: 4150.0},
		{ ID: "c4", Brand: "Audi", Model: "A6", Year: 2010, Color: "red", Owner: "3", Malfunctions: []Malfunction{
			{ Description: "Popravak motora", Price: 28.0 },
			{ Description: "Zamena retrovizora", Price: 8.0 },
			{ Description: "Zamena stop svetla", Price: 5.0 },
		}, Price: 2700.0},
		{ ID: "c5", Brand: "Audi", Model: "R8", Year: 2015, Color: "white", Owner: "2", Malfunctions: []Malfunction{
			{ Description: "Popravak klime", Price: 20.0 },
		}, Price: 4300.0},
		{ ID: "c6", Brand: "BMW", Model: "IX3", Year: 2020, Color: "blue", Owner: "2", Malfunctions: []Malfunction{
			{ Description: "Popravak kocnice", Price: 24.7 },
		}, Price: 5000.0},
	}

	for _, person := range persons {
		personJson, err := json.Marshal(person)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(person.ID, personJson)
		if err != nil {
			return fmt.Errorf("Failed to put to world state! %v", err)
		}
	}

	for _, car := range cars {
		carJson, err := json.Marshal(car)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(car.ID, carJson)
		if err != nil {
			return fmt.Errorf("Failed to put to world state! %v", err)
		}

		index := "color~owner~ID"
		key, err := ctx.GetStub().CreateCompositeKey(index, []string{car.Color, car.Owner, car.ID})
		if err != nil {
			return err
		}

		value := []byte{0x00}
		err = ctx.GetStub().PutState(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SmartContract) GetPerson(ctx contractapi.TransactionContextInterface, id string) (*Person, error) {
	personJson, err := ctx.GetStub().GetState(id)

	if personJson == nil {
		return nil, fmt.Errorf("Person with id %s does not exist!", id)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to load person from world state: %v", err)
	}

	var person Person
	err = json.Unmarshal(personJson, &person)
	if err != nil {
		return nil, err
	}

	return &person, nil
}

func (s *SmartContract) GetCar(ctx contractapi.TransactionContextInterface, id string) (*Car, error) {
	carJson, err := ctx.GetStub().GetState(id)

	if carJson == nil {
		return nil, fmt.Errorf("Car with id %s does not exist!", id)
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to load person from world state: %v", err)
	}

	var car Car
	err = json.Unmarshal(carJson, &car)
	if err != nil {
		return nil, err
	}

	return &car, nil
}

func (s *SmartContract) GetCarsByColor(ctx contractapi.TransactionContextInterface, color string) ([]*Car, error) {
	carsIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color})
	if err != nil {
		return nil, err
	}
	defer carsIter.Close()

	cars := make([]*Car, 0)
	for i := 0; carsIter.HasNext(); i++ {
		responseRange, err := carsIter.Next()
		if err != nil {
			return nil, err
		}

		_, keyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		carId := keyParts[2]
		car, err := s.GetCar(ctx, carId)
		if err != nil {
			return nil, err
		}

		cars = append(cars, car)
	}

	return cars, nil
}

func (s *SmartContract) GetCarsByOwnerAndColor(ctx contractapi.TransactionContextInterface, ownerId string, color string) ([]*Car, error) {
	personExists, err := s.OwnerExists(ctx, ownerId)
	if err != nil {
		return nil, err
	}

	if !personExists {
		return nil, fmt.Errorf("Person with id %s does not exist!", ownerId)
	}

	carsIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color, ownerId})
	if err != nil {
		return nil, err
	}
	defer carsIter.Close()

	cars := make([]*Car, 0)
	for i := 0; carsIter.HasNext(); i++ {
		responseRange, err := carsIter.Next()
		if err != nil {
			return nil, err
		}

		_, keyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		carId := keyParts[2]
		car, err := s.GetCar(ctx, carId)
		if err != nil {
			return nil, err
		}

		cars = append(cars, car)
	}

	return cars, nil
}

func (s *SmartContract) OwnerExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	owner, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("Failed to get person: %v", err)
	}

	return owner != nil, nil
}

func (s *SmartContract) ChangeColor(ctx contractapi.TransactionContextInterface, carId string, color string) (bool, error) {
	car, err := s.GetCar(ctx, carId)
	if err != nil {
		return false, err
	}

	prevColor := car.Color

	car.Color = color
	carJson, err := json.Marshal(car)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(carId, carJson)
	if err != nil {
		return false, err
	}

	newIndexKey, err := ctx.GetStub().CreateCompositeKey("color~owner~ID", []string{color, car.Owner, car.ID})
	if err != nil {
		return false, err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(newIndexKey, value)
	if err != nil {
		return false, err
	}

	oldIndexKey, err := ctx.GetStub().CreateCompositeKey("color~owner~ID", []string{prevColor, car.Owner, car.ID})
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(oldIndexKey)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SmartContract) AddNewMalfunction(ctx contractapi.TransactionContextInterface, carId string, description string, price float32) error {
	malfunction := Malfunction {
		Description: description,
		Price: price,
	}

	car, err := s.GetCar(ctx, carId)
	if err != nil {
		return err
	}

	car.Malfunctions = append(car.Malfunctions, malfunction)

	repairPrice := float32(0)
	for _, malfunction := range car.Malfunctions {
		repairPrice += malfunction.Price
	}

	if repairPrice > car.Price {
		return ctx.GetStub().DelState(carId)
	} else {
		carJson, err := json.Marshal(car)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(carId, carJson)
		if err != nil {
			return err
		}

		return nil
	}
}

func (s *SmartContract) RepairCar(ctx contractapi.TransactionContextInterface, carId string) (bool, error) {
	car, err := s.GetCar(ctx, carId)
	if err != nil {
		return false, err
	}

	owner, err := s.GetPerson(ctx, car.Owner)
	if err != nil {
		return false, err
	}

	toPayForRepairement := float32(0)
	for _, malfunction := range car.Malfunctions {
		toPayForRepairement += malfunction.Price
		if owner.Money < toPayForRepairement {
			return false, fmt.Errorf("Owner does not have enough money to pay!")
		}
	}

	car.Malfunctions = []Malfunction{}
	owner.Money -= toPayForRepairement

	carJson, err := json.Marshal(car)
	if err != nil {
		return false, err
	}

	ownerJson, err := json.Marshal(owner)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(carId, carJson)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(owner.ID, ownerJson)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SmartContract) BuyCar(ctx contractapi.TransactionContextInterface, carId string, buyerId string, answer string) (bool, error) {
	car, err := s.GetCar(ctx, carId)
	if err != nil {
		return false, err
	}

	buyer, err := s.GetPerson(ctx, buyerId)
	if err != nil {
		return false, err
	}

	var okayWithMalfunctions bool
	if answer == "yes" {
		okayWithMalfunctions = true
	} else if answer == "no" {
		okayWithMalfunctions = false
	}

	currentOwner, err := s.GetPerson(ctx, car.Owner)
	if err != nil {
		return false, err
	}

	if car.Owner == buyer.ID {
		return false, fmt.Errorf("Buyer is already owner of the car!")
	}

	carPrice := float32(0)

	if car.Malfunctions == nil || len(car.Malfunctions) == 0 {
		carPrice = car.Price
	} else if okayWithMalfunctions {
		moneyForMalfunctions := float32(0)
		for _, malfunction := range car.Malfunctions {
			moneyForMalfunctions += malfunction.Price
		}

		carPrice = car.Price - moneyForMalfunctions
	} else {
		return false, fmt.Errorf("Buyer does not want to buy the car.")
	}

	car.Owner = buyerId

	if buyer.Money >= carPrice {
		currentOwner.Money += carPrice
		buyer.Money -= carPrice
	} else {
		return false, fmt.Errorf("Buyer does not have enough money!")
	}

	carJson, err := json.Marshal(car)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(carId, carJson)
	if err != nil {
		return false, err
	}

	buyerJson, err := json.Marshal(buyer)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(buyerId, buyerJson)
	if err != nil {
		return false, err
	}

	sellerJson, err := json.Marshal(currentOwner)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(currentOwner.ID, sellerJson) 
	if err != nil {
		return false, nil
	}

	colorBuyerIndexKey, err := ctx.GetStub().CreateCompositeKey("color~owner~ID", []string{car.Color, buyerId, car.ID})
	if err != nil {
		return false, err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(colorBuyerIndexKey, value)
	if err != nil {
		return false, err
	}

	colorSellerIndexKey, err := ctx.GetStub().CreateCompositeKey("color~owner~ID", []string{car.Color, currentOwner.ID, car.ID})
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(colorSellerIndexKey)
	if err != nil {
		return false, err
	}

	return true, nil

}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
