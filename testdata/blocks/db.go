package blocks

import (
	"time"

	"github.com/s7techlab/hlf-sdk-go/observer/transform"
)

const (
	syscc = "syscc"
)

var chaincodes = [][]interface{}{
	{transform.LifecycleChaincodeName, syscc},
	//{FabcarChaincode, CCVersion},
	//{SampleChaincode, CCVersion},
}

var now = time.Now()

var channels = [][]interface{}{
	//{FabcarChannel, now, "{}"},
	//{SampleChannel, now, "{}"},
}

//func InitDBForAPITests(db *sql.DB) error {
//
//	for _, cc := range chaincodes {
//		_, err := db.Exec(`
//			insert into chaincode (id, version)
//			values ($1, $2)
//		`, cc...)
//		if err != nil {
//			return fmt.Errorf("add chaincode: %w", err)
//		}
//	}
//
//	for _, c := range channels {
//		_, err := db.Exec(`
//			insert into channel (id, created_at, config_parsed)
//			values ($1, $2, $3)
//		`, c...)
//		if err != nil {
//			return fmt.Errorf("add channel: %w", err)
//		}
//	}
//
//	for _, cc := range chaincodes {
//		for _, c := range channels {
//			_, err := db.Exec(`
//				insert into channel_chaincodes (channel_id, chaincode_id, chaincode_version, created_at)
//				values ($1, $2, $3, $4)
//			`, c[0], cc[0], cc[1], now)
//			if err != nil {
//				return fmt.Errorf("add channel_chaincode: %w", err)
//			}
//		}
//	}
//
//	return nil
//}
