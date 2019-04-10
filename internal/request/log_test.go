package request

import (
	"fmt"

	"github.com/aphistic/sweet"
	. "github.com/onsi/gomega"
)

type LogSuite struct{}

func (s *LogSuite) TestLog(t sweet.T) {
	log := NewLog(0)
	go readAll(log)

	requests := []*Request{
		&Request{Path: "/foo"},
		&Request{Path: "/bar"},
		&Request{Path: "/baz"},
	}

	for _, r := range requests {
		log.Add(r)
	}

	Expect(log.Copy(false)).To(Equal(requests))
	Expect(log.Copy(true)).To(Equal(requests))
	Expect(log.Copy(false)).To(BeEmpty())
}

func (s *LogSuite) TestCapacity(t sweet.T) {
	log := NewLog(5)
	go readAll(log)

	for i := 0; i < 20; i++ {
		log.Add(&Request{
			Path: fmt.Sprintf("%d", i+1),
		})
	}

	Expect(log.Copy(true)).To(Equal([]*Request{
		&Request{Path: "16"},
		&Request{Path: "17"},
		&Request{Path: "18"},
		&Request{Path: "19"},
		&Request{Path: "20"},
	}))
}

func (s *LogSuite) TestClear(t sweet.T) {
	log := NewLog(5)
	go readAll(log)

	log.Add(&Request{Path: "1"})
	log.Add(&Request{Path: "2"})
	log.Add(&Request{Path: "3"})
	log.Clear()
	log.Add(&Request{Path: "4"})
	log.Add(&Request{Path: "5"})

	Expect(log.Copy(true)).To(Equal([]*Request{
		&Request{Path: "4"},
		&Request{Path: "5"},
	}))
}

func (s *LogSuite) TestChan(t sweet.T) {
	log := NewLog(5)

	chanRequests := []*Request{}
	go func() {
		for r := range log.Chan() {
			chanRequests = append(chanRequests, r)
		}
	}()

	requests := []*Request{
		&Request{Path: "/foo"},
		&Request{Path: "/bar"},
		&Request{Path: "/baz"},
	}

	for _, r := range requests {
		log.Add(r)
	}

	Expect(chanRequests).To(Equal(requests))
}

func readAll(log Log) {
	for range log.Chan() {
	}
}
