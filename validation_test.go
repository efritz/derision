package main

import (
	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type ValidationSuite struct{}

func (s *ValidationSuite) TestConvertExpectation(t sweet.T) {
	e, err := convertExpectation(jsonExpectation{
		Method:  `PUT|PATCH`,
		Path:    `^/admin(/.*)?$`,
		Headers: map[string]string{"Authorization": `^Basic .+$`},
	})

	Expect(err).To(BeNil())
	Expect(e.method).NotTo(BeNil())
	Expect(e.path).NotTo(BeNil())
	Expect(e.headers).To(HaveKey("Authorization"))
	Expect(e.headers["Authorization"]).NotTo(BeNil())
}

func (s *ValidationSuite) TestConvertExpectationEmptyRegex(t sweet.T) {
	e, err := convertExpectation(jsonExpectation{
		Method:  ``,
		Path:    ``,
		Headers: map[string]string{"Authorization": ``},
	})

	Expect(err).To(BeNil())
	Expect(e.method).To(BeNil())
	Expect(e.path).To(BeNil())
	Expect(e.headers).NotTo(HaveKey("Authorization"))
}

func (s *ValidationSuite) TestConvertExpectationBadMethodJSON(t sweet.T) {
	_, err := convertExpectation(jsonExpectation{
		Method:  `[`,
		Path:    ``,
		Headers: nil,
	})

	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("validation error"))
}

func (s *ValidationSuite) TestConvertExpectationBadPathJSON(t sweet.T) {
	_, err := convertExpectation(jsonExpectation{
		Method:  ``,
		Path:    `[`,
		Headers: nil,
	})

	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("validation error"))
}

func (s *ValidationSuite) TestConvertExpectationBadHeaderJSON(t sweet.T) {
	_, err := convertExpectation(jsonExpectation{
		Method:  ``,
		Path:    ``,
		Headers: map[string]string{"Authorization": `[`},
	})

	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("validation error"))
}

func (s *ValidationSuite) TestConvertExpectationBadBodyJSON(t sweet.T) {
	_, err := convertExpectation(jsonExpectation{
		Method:  ``,
		Body:    `[`,
		Headers: nil,
	})

	Expect(err).NotTo(BeNil())
	Expect(err.Error()).To(Equal("validation error"))
}
func (s *ValidationSuite) TestCompile(t sweet.T) {
	_, err := compile(`^\d+`)
	Expect(err).To(BeNil())
}

func (s *ValidationSuite) TestCompileEmpty(t sweet.T) {
	re, err := compile(``)
	Expect(re).To(BeNil())
	Expect(err).To(BeNil())
}
