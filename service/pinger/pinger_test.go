package pinger

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/davidharrigan/pinger/grpc/protos"
)

func server(ctx context.Context) (pb.PingerClient, func()) {
	buffer := 1024 * 1024
	listener := bufconn.Listen(buffer)

	s := grpc.NewServer()
	pb.RegisterPingerServer(s, &Pinger{})
	go func() {
		if err := s.Serve(listener); err != nil {
			panic(err)
		}
	}()

	conn, _ := grpc.DialContext(ctx, "", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure(), grpc.WithBlock())

	closer := func() {
		s.Stop()
	}

	client := pb.NewPingerClient(conn)

	return client, closer
}

func TestPinger(t *testing.T) {
	type expectation struct {
		out *pb.PingResponse
		err error
	}

	tcs := map[string]struct {
		in       *pb.PingRequest
		expected expectation
	}{
		"ok": {
			in: &pb.PingRequest{},
			expected: expectation{
				out: &pb.PingResponse{
					Payload: []byte(`pong`),
				},
			},
		},
	}

	for scenario, tc := range tcs {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			assert := assert.New(t)

			client, closer := server(ctx)
			defer closer()

			out, err := client.Ping(ctx, tc.in)
			assert.Nil(err)

			if tc.expected.err == nil {
				assert.Nil(err)
				assert.Equal(tc.expected.out, out)
			} else {
				assert.Nil(out)
				assert.Equal(tc.expected.err, err)
			}
		})
	}
}

func TestPingerStream(t *testing.T) {

	type expectation struct {
		count int
		out   *pb.PingResponse
		err   error
	}

	tcs := map[string]struct {
		in       *pb.PingRequest
		expected expectation
	}{
		"ok": {
			in: &pb.PingRequest{Count: 5},
			expected: expectation{
				count: 5,
				out: &pb.PingResponse{
					Payload: []byte(`pong`),
				},
			},
		},
	}

	for scenario, tc := range tcs {
		t.Run(scenario, func(t *testing.T) {
			assert := assert.New(t)
			ctx := context.Background()

			client, closer := server(ctx)
			defer closer()

			stream, err := client.PingStream(ctx, tc.in)
			if !assert.Nil(err) {
				return
			}

			for i := 0; i < tc.expected.count; i++ {
				out, err := stream.Recv()
				if tc.expected.err == nil {
					assert.Nil(err)
					assert.Equal(tc.expected.out, out)
				} else {
					assert.Nil(out)
					assert.Equal(tc.expected.err, err)
				}
			}

			_, err = stream.Recv()
			assert.Equal(io.EOF, err)
		})
	}
}

func TestNoRPCs(t *testing.T) {
	_, closer := server(context.Background())
	defer closer()
}
