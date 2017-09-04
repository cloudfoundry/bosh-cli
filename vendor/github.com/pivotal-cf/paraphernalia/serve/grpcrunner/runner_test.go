package grpcrunner_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/grpc_testing"

	"github.com/pivotal-cf/paraphernalia/serve/grpcrunner"
)

var _ = Describe("GRPC Server", func() {
	var (
		logger      lager.Logger
		dummyServer *DummyServer

		listenAddr string
		runner     ifrit.Runner
		process    ifrit.Process
	)

	BeforeEach(func() {
		listenAddr = fmt.Sprintf("localhost:%d", GinkgoParallelNode()+9000)
		dummyServer = &DummyServer{}

		logger = lagertest.NewTestLogger("grpc-server")
		runner = grpcrunner.New(logger, listenAddr, func(server *grpc.Server) {
			grpc_testing.RegisterTestServiceServer(server, dummyServer)
		})
		process = ginkgomon.Invoke(runner)
	})

	AfterEach(func() {
		ginkgomon.Interrupt(process)
	})

	It("exits when signaled", func() {
		process.Signal(os.Interrupt)
		Eventually(process.Wait()).Should(Receive())
	})

	Context("when given a request", func() {
		var (
			conn   *grpc.ClientConn
			client grpc_testing.TestServiceClient
		)

		BeforeEach(func() {
			var err error
			conn, err = grpc.Dial(
				listenAddr,
				grpc.WithInsecure(),
				grpc.WithBlock(),
			)
			Expect(err).NotTo(HaveOccurred())

			client = grpc_testing.NewTestServiceClient(conn)
		})

		AfterEach(func() {
			conn.Close()
		})

		It("is a real GRPC server", func() {
			_, err := client.EmptyCall(context.Background(), &grpc_testing.Empty{})
			Expect(err).NotTo(HaveOccurred())

			Expect(dummyServer.CallCount()).To(Equal(1))
		})
	})
})

type DummyServer struct {
	callCount int
}

func (d *DummyServer) CallCount() int {
	return d.callCount
}

func (d *DummyServer) EmptyCall(ctx context.Context, e *grpc_testing.Empty) (*grpc_testing.Empty, error) {
	d.callCount++

	return e, nil
}

func (d *DummyServer) UnaryCall(context.Context, *grpc_testing.SimpleRequest) (*grpc_testing.SimpleResponse, error) {
	return nil, nil
}

func (d *DummyServer) StreamingOutputCall(*grpc_testing.StreamingOutputCallRequest, grpc_testing.TestService_StreamingOutputCallServer) error {
	return nil
}

func (d *DummyServer) StreamingInputCall(grpc_testing.TestService_StreamingInputCallServer) error {
	return nil
}

func (d *DummyServer) FullDuplexCall(grpc_testing.TestService_FullDuplexCallServer) error {
	return nil
}

func (d *DummyServer) HalfDuplexCall(grpc_testing.TestService_HalfDuplexCallServer) error {
	return nil
}
