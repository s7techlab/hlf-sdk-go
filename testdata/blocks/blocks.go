package blocks

const (
	Path = "testdata/blocks/fixtures"

	SampleChannel              = "sample-channel"
	SampleChannelHeight uint64 = 10
	FabcarChannel              = "fabcar-channel"
	FabcarChannelHeight uint64 = 12

	SampleChaincode = "sample"
	FabcarChaincode = "fabcar"
)

var (
	Channels        = []string{SampleChannel, FabcarChannel}
	ChannelsHeights = map[string]uint64{SampleChannel: SampleChannelHeight, FabcarChannel: FabcarChannelHeight}

	Chaincodes = []string{SampleChaincode, FabcarChaincode}
)

//import (
//	"reflect"
//	"sort"
//
//	"github.com/golang/protobuf/ptypes/timestamp"
//	"github.com/hyperledger/fabric-protos-go/peer"
//	"github.com/hyperledger/fabric/core/chaincode/lifecycle"
//	"github.com/s7techlab/cckit/examples/fabcar"
//)
//
//const (
//	Path = "testdata/blocks/fixtures"
//
//	TotalCertificatesWithQueryIssuerLikeOrg1 = 8
//
//	SampleChannel = "sample-channel"
//	FabcarChannel = "fabcar-channel"
//
//	SampleChaincode = "sample"
//	FabcarChaincode = "fabcar"
//	CCVersion       = "1.0"
//
//	EntityNameCar             = "car"
//	EntityNameCarOwner        = "car_owner"
//	EntityNameCarDetail       = "car_detail"
//	TotalCompositeEntities    = 2
//	TotalCompositeStatesTypes = 1
//
//	initialized = "\x00\U0010ffffinitialized"
//)
//
//var (
//	valid   = peer.TxValidationCode_VALID.String()
//	invalid = "IN" + valid
//
//	CarCEType          = reflect.TypeOf(fabcar.CarView{}).Name()
//	CarCompositeEntity = &config.CompositeEntity{
//		Type: CarCEType,
//		MainEntity: &config.Entity{
//			Type: reflect.TypeOf(fabcar.Car{}).Name(),
//			Name: EntityNameCar,
//		},
//		Entities: []*config.Entity{
//			{
//				Type: reflect.TypeOf(fabcar.CarOwner{}).Name(),
//				Name: EntityNameCarOwner,
//			},
//			{
//				Type: reflect.TypeOf(fabcar.CarDetail{}).Name(),
//				Name: EntityNameCarDetail,
//			},
//		},
//	}
//)
//
//type (
//	testChannels map[string]struct {
//		Data map[int]struct {
//			States       []*entity.ChaincodeState
//			ReadStates   []*entity.ChaincodeReadSetState
//			Transactions []*entity.Transaction
//		}
//		Certificates    int
//		ChannelsHistory int
//	}
//
//	// GetTransactionsOpts - фильтры для GetTransactions
//	GetTransactionsOpts struct {
//		ChannelId   string
//		ChaincodeId string
//		Method      string
//		// Искать только по невалидным txs
//		InvalidTxs bool
//	}
//	GetTransactionsOpt func(*GetTransactionsOpts)
//)
//
//var TestChannels = testChannels{
//	SampleChannel: {
//		Data: map[int]struct {
//			States       []*entity.ChaincodeState
//			ReadStates   []*entity.ChaincodeReadSetState
//			Transactions []*entity.Transaction
//		}{
//			0: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			1: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			2: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			3: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			4: {
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs(lifecycle.ApproveChaincodeDefinitionForMyOrgFuncName)),
//				},
//			},
//			5: {
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs(lifecycle.ApproveChaincodeDefinitionForMyOrgFuncName)),
//				},
//			},
//			6: {
//				States: []*entity.ChaincodeState{
//					state(transform.Collection+"_"+SampleChaincode, transform.Collection, transform.LifecycleChaincodeName, `{"raw":"EgA="}`),
//					state(transform.EndorsementInfo+"_"+SampleChaincode, transform.EndorsementInfo, transform.LifecycleChaincodeName, `{"raw":"Eg0KAzEuMBABGgRlc2Nj"}`),
//					state(transform.Sequence+"_"+SampleChaincode, transform.Sequence, transform.LifecycleChaincodeName, `{"raw":"CAE="}`),
//					state(transform.ValidationInfo+"_"+SampleChaincode, transform.ValidationInfo, transform.LifecycleChaincodeName, `{"raw":"EioKBHZzY2MSIhIgL0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ="}`),
//					state(transform.MetadataPrefix+"_"+SampleChaincode, transform.MetadataPrefix, transform.LifecycleChaincodeName, `{"raw":"ChNDaGFpbmNvZGVEZWZpbml0aW9uEghTZXF1ZW5jZRIPRW5kb3JzZW1lbnRJbmZvEg5WYWxpZGF0aW9uSW5mbxILQ29sbGVjdGlvbnM="}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{ // 12
//					readState(transform.FieldsPrefix+"/"+SampleChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+SampleChaincode+"/"+transform.Sequence, SampleChannel, SampleChaincode, 6, 0),
//					readState(transform.FieldsPrefix+"/"+SampleChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+SampleChaincode+"/"+transform.Sequence, SampleChannel, SampleChaincode, 6, 0),
//					readState(transform.FieldsPrefix+"/"+SampleChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+SampleChaincode+"/"+transform.Sequence, SampleChannel, SampleChaincode, 6, 0),
//					readState(transform.Sequence+"_"+SampleChaincode, transform.Sequence, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.Sequence+"_"+SampleChaincode, transform.Sequence, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.Sequence+"_"+SampleChaincode, transform.Sequence, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.MetadataPrefix+"_"+SampleChaincode, transform.MetadataPrefix, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.MetadataPrefix+"_"+SampleChaincode, transform.MetadataPrefix, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.MetadataPrefix+"_"+SampleChaincode, transform.MetadataPrefix, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(SampleChaincode, SampleChaincode, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(SampleChaincode, SampleChaincode, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(SampleChaincode, SampleChaincode, SampleChannel, transform.LifecycleChaincodeName, 0, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs(lifecycle.CommitChaincodeDefinitionFuncName)),
//				},
//			},
//			7: {
//				States: []*entity.ChaincodeState{
//					state(initialized, initialized, SampleChaincode, `{"raw": "MC4w"}`),
//					state(`CAR0`, `CAR0`, SampleChaincode, `{"make": "Toyota", "model": "Prius", "owner": "Tomoko", "colour": "blue"}`),
//					state(`CAR1`, `CAR1`, SampleChaincode, `{"make": "Ford", "model": "Mustang", "owner": "Brad", "colour": "red"}`),
//					state(`CAR2`, `CAR2`, SampleChaincode, `{"make": "Hyundai", "model": "Tucson", "owner": "Jin Soo", "colour": "green"}`),
//					state(`CAR3`, `CAR3`, SampleChaincode, `{"make": "Volkswagen", "model": "Passat", "owner": "Max", "colour": "yellow"}`),
//					state(`CAR4`, `CAR4`, SampleChaincode, `{"make": "Tesla", "model": "S", "owner": "Adriana", "colour": "black"}`),
//					state(`CAR5`, `CAR5`, SampleChaincode, `{"make": "Peugeot", "model": "205", "owner": "Michel", "colour": "purple"}`),
//					state(`CAR6`, `CAR6`, SampleChaincode, `{"make": "Chery", "model": "S22L", "owner": "Aarav", "colour": "white"}`),
//					state(`CAR7`, `CAR7`, SampleChaincode, `{"make": "Fiat", "model": "Punto", "owner": "Pari", "colour": "violet"}`),
//					state(`CAR8`, `CAR8`, SampleChaincode, `{"make": "Tata", "model": "Nano", "owner": "Valeria", "colour": "indigo"}`),
//					state(`CAR9`, `CAR9`, SampleChaincode, `{"make": "Holden", "model": "Barina", "owner": "Shotaro", "colour": "brown"}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{ // 3
//					readState(initialized, initialized, SampleChannel, SampleChaincode, 0, 0),
//					readState(initialized, initialized, SampleChannel, SampleChaincode, 7, 0),
//					readState(initialized, initialized, SampleChannel, SampleChaincode, 7, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs("InitLedger"), withChaincode(SampleChaincode)),
//				},
//			},
//			8: {
//				States: []*entity.ChaincodeState{
//					state(`CAR10`, `CAR10`, SampleChaincode, `{"make": "Toyota", "model": "Prius", "owner": "Tomoko", "colour": "blue"}`),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs("CreateCar", "CAR10", "Toyota", "Prius", "blue", "Tomoko"), withChaincode(SampleChaincode)),
//				},
//			},
//			9: {
//				States: []*entity.ChaincodeState{
//					state(`CAR11`, `CAR11`, SampleChaincode, `{"make": "Ford", "model": "Mustang", "owner": "Brad", "colour": "red"}`),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs("CreateCar", "CAR11", "Ford", "Mustang", "red", "Brad"), withChaincode(SampleChaincode)),
//				},
//			},
//		},
//		Certificates:    17,
//		ChannelsHistory: 4, // [0;3] blocks has CONFIG type transactions, that's why ChannelsHistory is 4
//	},
//	FabcarChannel: {
//		Data: map[int]struct {
//			States       []*entity.ChaincodeState
//			ReadStates   []*entity.ChaincodeReadSetState
//			Transactions []*entity.Transaction
//		}{
//			0: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			1: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			2: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			3: {
//				Transactions: []*entity.Transaction{
//					transaction(valid),
//				},
//			},
//			4: {
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs(lifecycle.ApproveChaincodeDefinitionForMyOrgFuncName)),
//				},
//			},
//			5: {
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs(lifecycle.ApproveChaincodeDefinitionForMyOrgFuncName)),
//				},
//			},
//			6: {
//				States: []*entity.ChaincodeState{
//					state(transform.Collection+"_"+FabcarChaincode, transform.Collection, transform.LifecycleChaincodeName, `{"raw":"EgA="}`),
//					state(transform.EndorsementInfo+"_"+FabcarChaincode, transform.EndorsementInfo, transform.LifecycleChaincodeName, `{"raw":"Eg0KAzEuMBABGgRlc2Nj"}`),
//					state(transform.Sequence+"_"+FabcarChaincode, transform.Sequence, transform.LifecycleChaincodeName, `{"raw":"CAE="}`),
//					state(transform.ValidationInfo+"_"+FabcarChaincode, transform.ValidationInfo, transform.LifecycleChaincodeName, `{"raw":"EioKBHZzY2MSIhIgL0NoYW5uZWwvQXBwbGljYXRpb24vRW5kb3JzZW1lbnQ="}`),
//					state(transform.MetadataPrefix+"_"+FabcarChaincode, transform.MetadataPrefix, transform.LifecycleChaincodeName, `{"raw":"ChNDaGFpbmNvZGVEZWZpbml0aW9uEghTZXF1ZW5jZRIPRW5kb3JzZW1lbnRJbmZvEg5WYWxpZGF0aW9uSW5mbxILQ29sbGVjdGlvbnM="}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{ // 14
//					readState(transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, FabcarChannel, FabcarChaincode, 6, 0),
//					readState(transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, FabcarChannel, FabcarChaincode, 6, 0),
//					readState(transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, FabcarChannel, FabcarChaincode, 6, 0),
//					readState(transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, FabcarChannel, FabcarChaincode, 6, 0),
//					readState(transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, transform.FieldsPrefix+"/"+FabcarChaincode+"/"+transform.Sequence, FabcarChannel, FabcarChaincode, 6, 0),
//					readState(transform.Sequence+"_"+FabcarChaincode, transform.Sequence, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.Sequence+"_"+FabcarChaincode, transform.Sequence, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.Sequence+"_"+FabcarChaincode, transform.Sequence, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.MetadataPrefix+"_"+FabcarChaincode, transform.MetadataPrefix, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.MetadataPrefix+"_"+FabcarChaincode, transform.MetadataPrefix, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(transform.MetadataPrefix+"_"+FabcarChaincode, transform.MetadataPrefix, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(FabcarChaincode, FabcarChaincode, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(FabcarChaincode, FabcarChaincode, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//					readState(FabcarChaincode, FabcarChaincode, FabcarChannel, transform.LifecycleChaincodeName, 0, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs(lifecycle.CommitChaincodeDefinitionFuncName)),
//				},
//			},
//			7: {
//				States: []*entity.ChaincodeState{
//					state(initialized, initialized, FabcarChaincode, `{"raw": "MC4w"}`),
//					state(`OWNER`, `OWNER`, FabcarChaincode, `{"PEM": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUNLekNDQWRHZ0F3SUJBZ0lSQUxxUDdvdDQvMWdYSWJZU0NwKzlaZlV3Q2dZSUtvWkl6ajBFQXdJd2N6RUwKTUFrR0ExVUVCaE1DVlZNeEV6QVJCZ05WQkFnVENrTmhiR2xtYjNKdWFXRXhGakFVQmdOVkJBY1REVk5oYmlCRwpjbUZ1WTJselkyOHhHVEFYQmdOVkJBb1RFRzl5WnpFdVpYaGhiWEJzWlM1amIyMHhIREFhQmdOVkJBTVRFMk5oCkxtOXlaekV1WlhoaGJYQnNaUzVqYjIwd0hoY05Nakl3TkRFek1URXlNakF3V2hjTk16SXdOREV3TVRFeU1qQXcKV2pCc01Rc3dDUVlEVlFRR0V3SlZVekVUTUJFR0ExVUVDQk1LUTJGc2FXWnZjbTVwWVRFV01CUUdBMVVFQnhNTgpVMkZ1SUVaeVlXNWphWE5qYnpFUE1BMEdBMVVFQ3hNR1kyeHBaVzUwTVI4d0hRWURWUVFEREJaVmMyVnlNVUJ2CmNtY3hMbVY0WVcxd2JHVXVZMjl0TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowREFRY0RRZ0FFelRTdjIzV0YKdmk2VlBhbFM0ZUp2OVM1anJMMjlMdnY2RFBhQkFZK1E3NzZzUEYxOHpGSG1XTm9kRUJoSUdoTlhFWmptTkVoYgpoMVBqbkFjUzE3dk9DYU5OTUVzd0RnWURWUjBQQVFIL0JBUURBZ2VBTUF3R0ExVWRFd0VCL3dRQ01BQXdLd1lEClZSMGpCQ1F3SW9BZ1RNS1AzNFpvUUxZMy9pY21KekdtREV2UlZCNjdOT1FnK09TNzJobnorVzR3Q2dZSUtvWkkKemowRUF3SURTQUF3UlFJaEFPdk9FZWJHb0VtTjlPUk11Y09PZlJ5NjVLczJCZlNtdnZ4eWxEUi9oQXFOQWlBdgpwRWpxUVZSOXk0UFRWUWtEZi9qa1Z5OG42UDJFZjN3RFBDNFduMjhKSmc9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==", "MSPId": "Org1MSP", "Issuer": "CN=ca.org1.example.com,O=org1.example.com,L=San Francisco,ST=California,C=US", "Subject": "CN=User1@org1.example.com,OU=client,L=San Francisco,ST=California,C=US"}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{ // 6
//					readState(initialized, initialized, FabcarChannel, FabcarChaincode, 7, 0),
//					readState(initialized, initialized, FabcarChannel, FabcarChaincode, 7, 0),
//					readState(initialized, initialized, FabcarChannel, FabcarChaincode, 0, 0),
//					readState(initialized, initialized, FabcarChannel, FabcarChaincode, 7, 0),
//					readState(initialized, initialized, FabcarChannel, FabcarChaincode, 7, 0),
//					readState(`OWNER`, `OWNER`, FabcarChannel, FabcarChaincode, 0, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withArgs("init"), withChaincode(FabcarChaincode)),
//				},
//			},
//			8: {
//				States: []*entity.ChaincodeState{
//					state(`Maker_Toyota`, `Maker`, FabcarChaincode, `{"name": "Toyota", "country": "Japan", "foundation_year": "1937"}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{
//					readState(`Maker_Toyota`, `Maker`, FabcarChannel, FabcarChaincode, 0, 0),
//					readState(`Maker_Toyota`, `Maker`, FabcarChannel, FabcarChaincode, 8, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withEvents([]*entity.Event{{NameEvent: `MakerCreated`}}), withArgs("FabCarService.CreateMaker"), withChaincode(FabcarChaincode)),
//				},
//			},
//			9: {
//				States: []*entity.ChaincodeState{
//					state(`Maker_Ford`, `Maker`, FabcarChaincode, `{"name": "Ford", "country": "USA", "foundation_year": "1903"}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{
//					readState(`Maker_Ford`, `Maker`, FabcarChannel, FabcarChaincode, 0, 0),
//					readState(`Maker_Ford`, `Maker`, FabcarChannel, FabcarChaincode, 9, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withEvents([]*entity.Event{{NameEvent: `MakerCreated`}}), withArgs("FabCarService.CreateMaker"), withChaincode(FabcarChaincode)),
//				},
//			},
//			10: {
//				States: []*entity.ChaincodeState{
//					state(`Car_Toyota_Prius_111999`, `Car`, FabcarChaincode, `{"id": ["Toyota", "Prius", "111999"], "make": "Toyota", "model": "Prius", "colour": "blue", "number": "111999"}`),
//					state(`CarDetail_Toyota_Prius_111999_BATTERY`, `CarDetail`, FabcarChaincode, `{"car_id": ["Toyota", "Prius", "111999"], "type": "BATTERY", "make": "BYD"}`),
//					state(`CarDetail_Toyota_Prius_111999_WHEELS`, `CarDetail`, FabcarChaincode, `{"car_id": ["Toyota", "Prius", "111999"], "type": "WHEELS", "make": "Michelin"}`),
//					state(`CarOwner_Toyota_Prius_111999_Tomoko_Uemura`, `CarOwner`, FabcarChaincode, `{"car_id": ["Toyota", "Prius", "111999"], "first_name": "Tomoko", "second_name": "Uemura", "vehicle_passport": "111aaa"}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{
//					readState(`Car_Toyota_Prius_111999`, `Car`, FabcarChannel, FabcarChaincode, 0, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withEvents([]*entity.Event{{NameEvent: `CarCreated`}}), withArgs("FabCarService.CreateCar"), withChaincode(FabcarChaincode)),
//				},
//			},
//			11: {
//				States: []*entity.ChaincodeState{
//					state(`Car_Ford_Mustang_222888`, `Car`, FabcarChaincode, `{"id": ["Ford", "Mustang", "222888"], "make": "Ford", "model": "Mustang", "colour": "red", "number": "222888"}`),
//					state(`CarDetail_Ford_Mustang_222888_WHEELS`, `CarDetail`, FabcarChaincode, `{"car_id": ["Ford", "Mustang", "222888"], "make": "Continental", "type": "WHEELS"}`),
//					state(`CarOwner_Ford_Mustang_222888_Brad_McDonald`, `CarOwner`, FabcarChaincode, `{"car_id": ["Ford", "Mustang", "222888"], "first_name": "Brad", "second_name": "McDonald", "vehicle_passport": "222bbb"}`),
//					state(`CarOwner_Ford_Mustang_222888_Michel_Tailor`, `CarOwner`, FabcarChaincode, `{"car_id": ["Ford", "Mustang", "222888"], "first_name": "Michel", "second_name": "Tailor", "vehicle_passport": "333ccc"}`),
//				},
//				ReadStates: []*entity.ChaincodeReadSetState{
//					readState(`Car_Ford_Mustang_222888`, `Car`, FabcarChannel, FabcarChaincode, 0, 0),
//				},
//				Transactions: []*entity.Transaction{
//					transaction(valid, withEvents([]*entity.Event{{NameEvent: `CarCreated`}}), withArgs("FabCarService.CreateCar"), withChaincode(FabcarChaincode)),
//				},
//			},
//		},
//		Certificates:    4,
//		ChannelsHistory: 4, // [0;3] blocks has CONFIG type transactions, that's why ChannelsHistory is 4
//	},
//}
//
//func (t *testChannels) GetTotalChannels() int {
//	return len(channels)
//}
//
//func (t *testChannels) GetTotalStates() int {
//	var totalStates int
//	for _, channel := range *t {
//		for _, data := range channel.Data {
//			totalStates += len(data.States)
//		}
//	}
//	return totalStates
//}
//
//func (t *testChannels) GetTotalReadStates() int {
//	var totalStates int
//	for _, channel := range *t {
//		for _, data := range channel.Data {
//			totalStates += len(data.ReadStates)
//		}
//	}
//	return totalStates
//}
//
//func (t *testChannels) GetTotalStatesWithChaincode(chaincodeId string) int {
//	var totalStatesWithChaincode int
//	for _, channel := range *t {
//		for _, data := range channel.Data {
//			for _, s := range data.States {
//				if s.ChaincodeId == chaincodeId {
//					totalStatesWithChaincode++
//				}
//			}
//		}
//	}
//	return totalStatesWithChaincode
//}
//
//func (t *testChannels) GetTotalReadStatesWithChaincode(chaincodeId string) int {
//	var totalStatesWithChaincode int
//	for _, channel := range *t {
//		for _, data := range channel.Data {
//			for _, s := range data.ReadStates {
//				if s.ChaincodeId == chaincodeId {
//					totalStatesWithChaincode++
//				}
//			}
//		}
//	}
//	return totalStatesWithChaincode
//}
//
//func (t *testChannels) GetTotalStatesWithChannelAndChaincode(channel, chaincode string) int {
//	var totalStatesWithChannelAndChaincode int
//	for _, channelData := range (*t)[channel].Data {
//		for _, s := range channelData.States {
//			if s.ChaincodeId == chaincode {
//				totalStatesWithChannelAndChaincode++
//			}
//		}
//	}
//	return totalStatesWithChannelAndChaincode
//}
//
//func (t *testChannels) GetTotalReadStatesWithChannelAndChaincode(channel, chaincode string) int {
//	var totalStatesWithChannelAndChaincode int
//	for _, channelData := range (*t)[channel].Data {
//		for _, s := range channelData.ReadStates {
//			if s.ChaincodeId == chaincode {
//				totalStatesWithChannelAndChaincode++
//			}
//		}
//	}
//	return totalStatesWithChannelAndChaincode
//}
//
//func DefaultGetTransactionsOpts() *GetTransactionsOpts {
//	return &GetTransactionsOpts{
//		// по-умолчанию, искать только по валидным txs
//		InvalidTxs: false,
//	}
//}
//
//// WithChannelId - добавляет фильтр по каналу
//func WithChannelId(channelId string) GetTransactionsOpt {
//	return func(opts *GetTransactionsOpts) {
//		opts.ChannelId = channelId
//	}
//}
//
//// WithChaincodeId - добавляет фильтр по чейнкоду
//func WithChaincodeId(chaincodeId string) GetTransactionsOpt {
//	return func(opts *GetTransactionsOpts) {
//		opts.ChaincodeId = chaincodeId
//	}
//}
//
//// WithMethod - добавляет фильтр по вызываемому методу
//func WithMethod(method string) GetTransactionsOpt {
//	return func(opts *GetTransactionsOpts) {
//		opts.Method = method
//	}
//}
//
//// WithInvalidTxs - добавляет фильтр по флагу валидности транзакции
//func WithInvalidTxs(invalidTxs bool) GetTransactionsOpt {
//	return func(opts *GetTransactionsOpts) {
//		opts.InvalidTxs = invalidTxs
//	}
//}
//
//// GetCountTransactionsByChannel - возвращает количество транзакций, отфильтрованных по чейнкоду
//func GetCountTransactionsByChannel(channelId string) int {
//	return len(TestChannels.GetTransactions(WithChannelId(channelId)))
//}
//func GetCountInvalidTransactionsByChannel(channelId string) int {
//	return len(TestChannels.GetTransactions(WithChannelId(channelId), WithInvalidTxs(true)))
//}
//
//// GetCountTransactionsByMethod - возвращает количество транзакций, отфильтрованных по вызываемому методу
//func GetCountTransactionsByMethod(method string) int {
//	return len(TestChannels.GetTransactions(WithMethod(method)))
//}
//func GetCountInvalidTransactionsByMethod(method string) int {
//	return len(TestChannels.GetTransactions(WithMethod(method), WithInvalidTxs(true)))
//}
//
//// GetCountTransactionsMethodsDistinct - возвращает число уникальных вызываемых методов в указанных канале и чейнкоде
//func GetCountTransactionsMethodsDistinct(channelId, chaincodeId string, invalidTxs bool) int {
//	txs := TestChannels.GetTransactions(WithChannelId(channelId), WithChaincodeId(chaincodeId), WithInvalidTxs(invalidTxs))
//
//	transactionsMethods := make(map[string]struct{})
//	for _, tx := range txs {
//		if len(tx.Args) > 0 {
//			transactionsMethods[tx.Args[0]] = struct{}{}
//		}
//	}
//
//	return len(transactionsMethods)
//}
//
//// GetTransactions - возвращает список транзакций с учетом фильтров
//func (t *testChannels) GetTransactions(opts ...GetTransactionsOpt) []*entity.Transaction {
//	getTransactionsOpts := DefaultGetTransactionsOpts()
//
//	for _, opt := range opts {
//		opt(getTransactionsOpts)
//	}
//
//	var transactions []*entity.Transaction
//	for id, channel := range *t {
//
//		// filter by channel id
//		if getTransactionsOpts.ChannelId != "" && id != getTransactionsOpts.ChannelId {
//			continue
//		}
//
//		for _, data := range channel.Data {
//			for _, tx := range data.Transactions {
//
//				// filter by valid txs
//				if !getTransactionsOpts.InvalidTxs && tx.Valid == invalid {
//					continue
//				}
//
//				// filter by invalid txs
//				if getTransactionsOpts.InvalidTxs && tx.Valid == valid {
//					continue
//				}
//
//				// filter by chaincode id
//				if getTransactionsOpts.ChaincodeId != "" && tx.ChaincodeId != getTransactionsOpts.ChaincodeId {
//					continue
//				}
//
//				// filter by method name
//				if getTransactionsOpts.Method != "" {
//					if len(tx.Args) == 0 {
//						continue
//					}
//					if tx.Args[0] != getTransactionsOpts.Method {
//						continue
//					}
//				}
//
//				transactions = append(transactions, tx)
//			}
//		}
//	}
//
//	return transactions
//}
//
//func (t *testChannels) GetTotalBlocks() int {
//	var totalBlocks int
//	for _, channel := range *t {
//		totalBlocks += len(channel.Data)
//	}
//	return totalBlocks
//}
//
//func (t *testChannels) GetTotalBlocksWithChannel(channel string) int {
//	return len((*t)[channel].Data)
//}
//
//func (t *testChannels) GetTotalCertificates() int {
//	var totalCertificates int
//	for _, channel := range *t {
//		totalCertificates += channel.Certificates
//	}
//	return totalCertificates
//}
//
//func (t *testChannels) GetTotalStatesTypes() int {
//	totalStatesTypes := make(map[string]struct{})
//	for _, channel := range *t {
//		for _, data := range channel.Data {
//			for _, s := range data.States {
//				totalStatesTypes[s.Type] = struct{}{}
//			}
//		}
//	}
//	return len(totalStatesTypes)
//}
//
//func (t *testChannels) GetTotalReadStatesTypes() int {
//	totalStatesTypes := make(map[string]struct{})
//	for _, channel := range *t {
//		for _, data := range channel.Data {
//			for _, s := range data.ReadStates {
//				totalStatesTypes[s.Type] = struct{}{}
//			}
//		}
//	}
//	return len(totalStatesTypes)
//}
//
//func (t *testChannels) GetTotalStatesTypesFromChannel(channel string) int {
//	totalStatesTypesFromChannel := make(map[string]struct{})
//	for _, data := range (*t)[channel].Data {
//		for _, s := range data.States {
//			totalStatesTypesFromChannel[s.Type] = struct{}{}
//		}
//	}
//	return len(totalStatesTypesFromChannel)
//}
//
//func (t *testChannels) GetTotalReadStatesTypesFromChannel(channel string) int {
//	totalStatesTypesFromChannel := make(map[string]struct{})
//	for _, data := range (*t)[channel].Data {
//		for _, s := range data.ReadStates {
//			totalStatesTypesFromChannel[s.Type] = struct{}{}
//		}
//	}
//	return len(totalStatesTypesFromChannel)
//}
//
//func (t *testChannels) GetTotalChaincodes() int {
//	return len(chaincodes)
//}
//
//func (t *testChannels) GetTotalChaincodesOfVersion(version string) int {
//	var totalChaincodesOfVersion int
//	for _, chaincode := range chaincodes {
//		if chaincode[1].(string) == version {
//			totalChaincodesOfVersion++
//		}
//	}
//	return totalChaincodesOfVersion
//}
//
//func (t *testChannels) GetTotalChannelsHistory() int {
//	var totalChannelsHistory int
//	for _, channel := range *t {
//		totalChannelsHistory += channel.ChannelsHistory
//	}
//	return totalChannelsHistory
//}
//
//func (t *testChannels) GetChannelHistory(channel string) int {
//	return (*t)[channel].ChannelsHistory
//}
//
//func (t *testChannels) GetTransactionsEvents() int {
//	transactionsEvents := make(map[string]struct{})
//	for _, channel := range *t {
//		for _, data := range channel.Data {
//			for _, tx := range data.Transactions {
//				for _, e := range tx.Events {
//					transactionsEvents[e.NameEvent] = struct{}{}
//				}
//			}
//		}
//	}
//
//	return len(transactionsEvents)
//}
//
//func (t *testChannels) GetTransactionsEventsFromChannel(channel string) int {
//	transactionsEvents := make(map[string]struct{})
//	for _, data := range (*t)[channel].Data {
//		for _, tx := range data.Transactions {
//			for _, e := range tx.Events {
//				transactionsEvents[e.NameEvent] = struct{}{}
//			}
//		}
//	}
//	return len(transactionsEvents)
//}
//
//func state(key, stateType, chaincodeId, data string) *entity.ChaincodeState {
//	return &entity.ChaincodeState{
//		Key:         key,
//		Type:        stateType,
//		ChaincodeId: chaincodeId,
//		Data:        []byte(data),
//		DataRaw:     []byte(data),
//		CreatedAt:   &timestamp.Timestamp{},
//	}
//}
//
//func readState(key, stateType, channelId, chaincodeId string, versionBlockNum, versionTxNum uint64) *entity.ChaincodeReadSetState {
//	return &entity.ChaincodeReadSetState{
//		Key:             key,
//		Type:            stateType,
//		ChannelId:       channelId,
//		ChaincodeId:     chaincodeId,
//		VersionBlockNum: versionBlockNum,
//		VersionTxNum:    versionTxNum,
//	}
//}
//
//type opt func(*entity.Transaction)
//
//func withArgs(args ...string) opt {
//	return func(tx *entity.Transaction) {
//		tx.Args = args
//	}
//}
//
//func withEvents(events []*entity.Event) opt {
//	return func(tx *entity.Transaction) {
//		tx.Events = events
//	}
//}
//
//func withChaincode(chaincodeId string) opt {
//	return func(tx *entity.Transaction) {
//		tx.ChaincodeId = chaincodeId
//	}
//}
//
//func transaction(valid string, opts ...opt) *entity.Transaction {
//	tx := &entity.Transaction{
//		Args:  []string{},
//		Valid: valid,
//	}
//
//	for _, o := range opts {
//		o(tx)
//	}
//
//	return tx
//}
//
//type Block struct {
//	Number       int
//	States       []*entity.ChaincodeState
//	ReadStates   []*entity.ChaincodeReadSetState
//	Transactions []*entity.Transaction
//}
//
//type blocksByNumber []Block
//
//func (bs blocksByNumber) Len() int {
//	return len(bs)
//}
//
//func (bs blocksByNumber) Less(i, j int) bool {
//	return bs[i].Number < bs[j].Number
//}
//
//func (bs blocksByNumber) Swap(i, j int) {
//	bs[i], bs[j] = bs[j], bs[i]
//}
//
//var TestChannelsRearranged = map[string][]Block{}
//
//func init() {
//	for cID, c := range TestChannels {
//		for bID, b := range c.Data {
//			TestChannelsRearranged[cID] = append(TestChannelsRearranged[cID],
//				Block{
//					Number:       bID,
//					States:       b.States,
//					Transactions: b.Transactions,
//				})
//		}
//		sort.Sort(blocksByNumber(TestChannelsRearranged[cID]))
//	}
//}
