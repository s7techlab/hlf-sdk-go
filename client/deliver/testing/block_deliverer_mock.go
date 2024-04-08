package testing

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	hlfproto "github.com/s7techlab/hlf-sdk-go/block"
)

type BlocksDelivererMock struct {
	//  <channel-name> => [<block1.pb>,...<blockN.pb>]
	data             map[string][]*common.Block
	parsedData       map[string][]*hlfproto.Block
	closeWhenAllRead bool
}

// NewBlocksDelivererMock - read all blocks in proto format from provided folder.
// returns sdkapi.BlocksDeliverer - interface
// closeWhenAllRead - close channel when all data have been written
// closerFunc - close func that will be returned from Blocks()
func NewBlocksDelivererMock(rootPath string, closeWhenAllRead bool) (*BlocksDelivererMock, error) {
	var err error

	dc := &BlocksDelivererMock{
		data:             make(map[string][]*common.Block),
		parsedData:       make(map[string][]*hlfproto.Block),
		closeWhenAllRead: closeWhenAllRead,
	}

	channels := make(map[string]map[int][]byte)

	if rootPath, err = filepath.EvalSymlinks(rootPath); err != nil {
		return nil, errors.Wrap(err, `failed to read real path`)
	}

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == rootPath {
			return nil
		}

		pathWithoutRoot := strings.TrimPrefix(path, rootPath+`/`)
		paths := strings.Split(pathWithoutRoot, `/`)

		if info.IsDir() {
			// is channel name
			if len(paths) == 1 {
				channels[paths[0]] = make(map[int][]byte)
				return nil
			} else if len(paths) > 1 {
				return filepath.SkipDir
			}
		}

		// is block of channel
		switch len(paths) {
		case 2:
			if filepath.Ext(paths[1]) != `.pb` {
				return nil
			}

			channel, ok := channels[paths[0]]
			if !ok {
				return nil
			}
			blockID, err := strconv.Atoi(strings.TrimSuffix(paths[1], `.pb`))
			if err != nil {
				return err
			}

			block, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			channel[blockID] = block
		default:
			println("IGNORED: ", path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	for channelID, data := range channels {
		channelBlocks := make([]*common.Block, len(data))
		parsedChannelBlocks := make([]*hlfproto.Block, len(data))
		for blockID, blockData := range data {
			block := &common.Block{}
			err = proto.Unmarshal(blockData, block)
			if err != nil {
				return nil, err
			}
			channelBlocks[blockID] = block

			parsedBlock, err := hlfproto.ParseBlock(block)
			if err != nil {
				return nil, err
			}
			parsedChannelBlocks[blockID] = parsedBlock
		}
		dc.data[channelID] = channelBlocks
		dc.parsedData[channelID] = parsedChannelBlocks
		println("fill channel '"+channelID+"' blocks from", 0, "...", len(channelBlocks)-1)
	}

	return dc, nil
}

func (m *BlocksDelivererMock) Blocks(
	_ context.Context,
	channelName string,
	_ msp.SigningIdentity,
	blockRange ...int64,
) (<-chan *common.Block, func() error, error) {

	return blocks[*common.Block](m.data, channelName, m.closeWhenAllRead, blockRange...)
}

func (m *BlocksDelivererMock) ParsedBlocks(
	_ context.Context,
	channelName string,
	_ msp.SigningIdentity,
	blockRange ...int64,
) (<-chan *hlfproto.Block, func() error, error) {

	return blocks[*hlfproto.Block](m.parsedData, channelName, m.closeWhenAllRead, blockRange...)
}

func blocks[T any](data map[string][]T, channelName string, closeWhenAllRead bool, blockRange ...int64) (<-chan T, func() error, error) {
	if _, ok := data[channelName]; !ok {
		return nil, nil, fmt.Errorf("have no mocked data for this channel")
	}
	closer := func() error { return nil }

	var (
		blockRangeFrom int64 = 0
		blockRangeTo   int64 = math.MaxInt64
	)

	if len(blockRange) > 0 {
		blockRangeFrom = blockRange[0]
	}
	if len(blockRange) > 1 {
		blockRangeTo = blockRange[1]
	}

	if blockRangeFrom < 0 {
		blockRangeFrom = int64(len(data[channelName])) + blockRangeFrom
	}

	if blockRangeTo < 0 {
		blockRangeTo = int64(len(data[channelName])) + blockRangeTo
	}

	if blockRangeFrom > int64(len(data[channelName])) {
		blockRangeFrom = int64(len(data[channelName])) - 1
	}

	if blockRangeTo > int64(len(data[channelName])) {
		blockRangeTo = int64(len(data[channelName])) - 1
	}

	ch := make(chan T, (blockRangeTo-blockRangeFrom)+1)
	for i := blockRangeFrom; i <= blockRangeTo; i++ {
		ch <- data[channelName][i]
	}

	if closeWhenAllRead {
		close(ch)
	}

	return ch, closer, nil
}
