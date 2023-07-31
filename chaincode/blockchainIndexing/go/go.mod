module github.com/hyperledger/fabric-samples/chaincode/fabcar/go

go 1.13

require (
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200424173110-d7076418f212
	github.com/hyperledger/fabric-protos-go v0.0.0-20200424173316-dd554ba3746e
	github.com/kr/pretty v0.2.0 // indirect
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
)

replace github.com/hyperledger/fabric-protos-go => github.com/dgaron/fabric-protos/bindings/go-apiv1 v0.0.0-20230728170636-34077e8f7509

replace github.com/hyperledger/fabric-chaincode-go => github.com/dgaron/fabric-chaincode-go v0.0.0-20230728171125-c298c3f19d55
