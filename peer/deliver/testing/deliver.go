package testing

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func NewDeliverClient(rootPath string, closeWhenAllRead bool) (peer.DeliverClient, error) {

	var err error

	dc := &deliverClient{
		data:             make(map[string][]*common.Block, 0),
		closeWhenAllRead: closeWhenAllRead,
	}

	channels := make(map[string]map[int][]byte, 0)

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

		pathWithoutRool := strings.TrimPrefix(path, rootPath+`/`)
		paths := strings.Split(pathWithoutRool, `/`)

		if info.IsDir() {
			// is channel name
			if len(paths) == 1 {
				channels[paths[0]] = make(map[int][]byte, 0)
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

			block, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			channel[blockID] = block
		default:
			println("IGNORED", path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	for channelID, data := range channels {
		channelBlocks := make([]*common.Block, len(data))
		for blockID, blockData := range data {
			block := &common.Block{}
			err := proto.Unmarshal(blockData, block)
			if err != nil {
				return nil, err
			}
			channelBlocks[blockID] = block
		}
		dc.data[channelID] = channelBlocks
		println("fill channel '"+channelID+"' blocks from", 0, "...", len(channelBlocks)-1)
	}

	return dc, nil
}

type deliverClient struct {
	ctx context.Context
	//  <channel-name> => [<block1.pb>,...<blockN.pb>]
	data map[string][]*common.Block

	blockService     *blockService
	closeWhenAllRead bool
}

func (d *deliverClient) DeliverWithPrivateData(ctx context.Context, opts ...grpc.CallOption) (peer.Deliver_DeliverWithPrivateDataClient, error) {
	panic("implement me")
}

func (d *deliverClient) Send(env *common.Envelope) error {
	payload, err := protoutil.UnmarshalPayload(env.Payload)
	if err != nil {
		return err
	}

	ch, err := protoutil.UnmarshalChannelHeader(payload.Header.ChannelHeader)
	if err != nil {
		return err
	}

	if ch.Type != int32(common.HeaderType_DELIVER_SEEK_INFO) {
		return fmt.Errorf("unsupporter headerType %s", common.HeaderType(ch.Type).String())
	}

	blocks, ok := d.data[ch.ChannelId]
	if !ok {
		return fmt.Errorf("channel %s not exists", ch.ChannelId)
	}

	seekInfo := new(orderer.SeekInfo)
	err = proto.Unmarshal(payload.Data, seekInfo)
	if err != nil {
		return err
	}

	var (
		startNumber int
		endNumber   = len(blocks) - 1
	)

	if seekInfo.GetStart().GetOldest() != nil {
		startNumber = 0
	} else if seekInfo.GetStart().GetNewest() != nil {
		startNumber = endNumber
	} else if seekInfo.GetStart().GetSpecified() != nil {
		startNumber = int(seekInfo.GetStart().GetSpecified().Number)
	}

	if seekInfo.GetStop().GetSpecified() != nil {
		stopNumber := int(seekInfo.GetStop().GetSpecified().Number)
		if stopNumber > 0 && stopNumber < endNumber {
			endNumber = stopNumber
		}
	}

	// end start len of *chan
	// 1   1     = 1
	// 90  0     = 91
	// 90  1     = 90
	// 85  85    = 1
	d.blockService.blocks = make(chan *common.Block, (endNumber-startNumber)+1)
	for i := startNumber; i <= endNumber; i++ {
		d.blockService.blocks <- blocks[i]
	}

	if d.blockService.closeWhenAllRead {
		close(d.blockService.blocks)
	}

	ctx, cancel := context.WithCancel(d.ctx)

	d.blockService.cancel = cancel

	go d.blockService.watch(ctx)

	return nil
}

func (d *deliverClient) Recv() (*peer.DeliverResponse, error) {
	select {
	case <-d.ctx.Done():
		return nil, d.ctx.Err()

	case b, ok := <-d.blockService.blocks:
		if !ok {
			return nil, io.EOF
		}
		return &peer.DeliverResponse{
			Type: &peer.DeliverResponse_Block{
				Block: b,
			},
		}, nil
	}
}

func (d *deliverClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (d *deliverClient) Trailer() metadata.MD {
	return nil
}

func (d *deliverClient) CloseSend() error {
	return nil
}

func (d *deliverClient) Context() context.Context {
	return d.ctx
}

func (d *deliverClient) SendMsg(m interface{}) error {
	panic("implement me")
}

func (d *deliverClient) RecvMsg(m interface{}) error {
	panic("implement me")
}

func (d *deliverClient) Deliver(ctx context.Context, opts ...grpc.CallOption) (peer.Deliver_DeliverClient, error) {
	d.blockService = &blockService{
		once:             &sync.Once{},
		errC:             make(chan error),
		closeWhenAllRead: d.closeWhenAllRead,
	}
	d.ctx = ctx

	return d, nil
}

func (d *deliverClient) DeliverFiltered(ctx context.Context, opts ...grpc.CallOption) (peer.Deliver_DeliverFilteredClient, error) {
	panic("unimplemented")
}

type blockService struct {
	blocks           chan *common.Block
	errC             chan error
	once             *sync.Once
	cancel           context.CancelFunc
	closeWhenAllRead bool
}

func (b *blockService) watch(ctx context.Context) {
	<-ctx.Done()
	if b.errC == nil {
		return
	}

	select {
	case b.errC <- ctx.Err(): // if we can write to error channel try close
		_ = b.Close()
	default:
	}
}

func (b *blockService) Blocks() <-chan *common.Block {
	return b.blocks
}

func (b *blockService) Err() <-chan error {
	return b.errC
}

func (b *blockService) Close() error {
	if b.blocks != nil {
		b.once.Do(func() {
			select {
			case _, ok := <-b.blocks:
				if ok && !b.closeWhenAllRead {
					close(b.blocks)
				}
			default:
				if !b.closeWhenAllRead {
					close(b.blocks)
				}
			}

			close(b.errC)
			b.blocks = nil
			b.errC = nil
			b.cancel()
		})
	}

	return nil
}
