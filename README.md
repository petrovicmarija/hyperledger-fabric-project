# Hyperledger Fabric project

## Start network

To start the network, create channel and deploy chaincode go to the folder client-application:
```bash
cd app/client-application
```
and run the script:

```bash
./startFabric.sh
```
<i> startFabric.sh </i> will: 
1. Bring down network (if it's not alreday down).
2. Start network - it will create 4 organizations with 4 peer nodes on each, one orderer node, using 4 CAs (Certificate Authorities) to generate network crypto material.
3. Create channel <i> mychannel. </i>
4. Deploy smart contract <i> carcc </i> to the previously created channel (<i> mychannel </i>).

Chaincode can be found on location:
```bash
cd app/chaincode/cars/go/
```

## Start application
Once the network is up and chaincode is installed on peers, go to:
```bash
cd app/client-application/go
```
and run the script which will run client application:
```bash
./runclient.sh
```
<i> runclient.sh </i> will run following command:
```go
go run fabcar.go
```
and chaincode is ready to be invoked using console.
