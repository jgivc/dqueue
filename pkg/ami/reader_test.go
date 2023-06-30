package ami_test

import (
	"io"
	"os"
	"testing"

	"github.com/jgivc/vapp/pkg/ami"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"
)

const (
	readerDataFile   = "testdata/reader_data.txt"
	readerEventsFile = "testdata/reader_events.yml"
)

type AmiReaderTestSuite struct {
	suite.Suite
	expected []ami.Event
}

func (s *AmiReaderTestSuite) SetupSuite() {
	ofile, err := os.Open(readerEventsFile)
	if err != nil {
		s.Require().NoError(err)
	}
	defer ofile.Close()

	dec := yaml.NewDecoder(ofile)
	if err2 := dec.Decode(&s.expected); err2 != nil {
		s.Require().NoError(err)
	}
}

func (s *AmiReaderTestSuite) TestAmiReader() {
	file, err := os.Open(readerDataFile)
	if err != nil {
		s.Require().NoError(err)
	}
	defer file.Close()

	ar := ami.NewAmiReader(file)
	defer ar.Close()

	out := make([]ami.Event, 0)
	for {
		e, err2 := ar.Read()
		if err2 != nil {
			s.Require().ErrorIs(err2, io.EOF)
			break
		}

		out = append(out, e)
	}

	s.Assert().ElementsMatch(s.expected, out)
}

func TestAmiReaderTestSuite(t *testing.T) {
	suite.Run(t, new(AmiReaderTestSuite))
}
