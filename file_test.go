package main

import (
	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type FileSuite struct{}

func (s *FileSuite) TestMakeHandlersFromFileMissing(t sweet.T) {
	_, err := makeHandlersFromFile("test/missing.json")
	Expect(err).NotTo(BeNil())
}

func (s *FileSuite) TestMakeHandlersFromFileMalformed(t sweet.T) {
	_, err := makeHandlersFromFile("test/malformed.json")
	Expect(err).NotTo(BeNil())
}

func (s *FileSuite) TestMakeHandlersFromFileUnprocessable(t sweet.T) {
	_, err := makeHandlersFromFile("test/unprocessable.json")
	Expect(err).NotTo(BeNil())
}

func (s *FileSuite) TestMakeHandlersFromFileInvalid(t sweet.T) {
	_, err := makeHandlersFromFile("test/invalid.json")
	Expect(err).NotTo(BeNil())
	Expect(err).To(BeAssignableToTypeOf(&validationError{}))
}

func (s *FileSuite) TestMakeHandlersFromFileEmpty(t sweet.T) {
	handlers, err := makeHandlersFromFile("test/empty.json")
	Expect(err).To(BeNil())
	Expect(handlers).To(BeEmpty())
}

func (s *FileSuite) TestMakeHandlersFromFileValid(t sweet.T) {
	handlers, err := makeHandlersFromFile("test/valid.json")
	Expect(err).To(BeNil())
	Expect(handlers).To(HaveLen(1))
	testHandlerBehavior(handlers[0])
}
