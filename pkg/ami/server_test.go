package ami

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AmiServerTestSuite struct {
	suite.Suite
	server amiServer
}

func (s *AmiServerTestSuite) TestOne() {

	if a, err := net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			fmt.Println(l.Addr().(*net.TCPAddr).Port)
		}
	}
}

func TestAmiServerTestSuite(t *testing.T) {
	suite.Run(t, new(AmiServerTestSuite))
}
