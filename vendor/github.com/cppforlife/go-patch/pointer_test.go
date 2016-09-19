package patch_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/go-patch"
)

type PointerTestCase struct {
	String string
	Tokens []Token
}

var testCases = []PointerTestCase{
	{"", []Token{RootToken{}}},

	// Root level
	{"/", []Token{RootToken{}, KeyToken{Key: ""}}},
	{"/ ", []Token{RootToken{}, KeyToken{Key: " "}}},

	// Escaping (todo support ~2 for '?'; ~3 for '=')
	{"/m~0n", []Token{RootToken{}, KeyToken{Key: "m~n"}}},
	{"/a~01b", []Token{RootToken{}, KeyToken{Key: "a~1b"}}},
	{"/a~1b", []Token{RootToken{}, KeyToken{Key: "a/b"}}},

	// Speacial chars
	{"/c%d", []Token{RootToken{}, KeyToken{Key: "c%d"}}},
	{"/e^f", []Token{RootToken{}, KeyToken{Key: "e^f"}}},
	{"/g|h", []Token{RootToken{}, KeyToken{Key: "g|h"}}},
	{"/i\\j", []Token{RootToken{}, KeyToken{Key: "i\\j"}}},
	{"/k\"l", []Token{RootToken{}, KeyToken{Key: "k\"l"}}},

	// Maps
	{"/key", []Token{RootToken{}, KeyToken{Key: "key"}}},
	{"/key/", []Token{RootToken{}, KeyToken{Key: "key", Expected: true}, KeyToken{Key: ""}}},
	{"/key/key2", []Token{RootToken{}, KeyToken{Key: "key", Expected: true}, KeyToken{Key: "key2"}}},
	{"/key?/key2/key3", []Token{RootToken{}, KeyToken{Key: "key"}, KeyToken{Key: "key2"}, KeyToken{Key: "key3"}}},

	// Array indices
	{"/0", []Token{RootToken{}, IndexToken{0}}},
	{"/1000001", []Token{RootToken{}, IndexToken{1000001}}},
	{"/-2", []Token{RootToken{}, IndexToken{-2}}},

	{"/-", []Token{RootToken{}, AfterLastIndexToken{}}},
	{"/ary/-", []Token{RootToken{}, KeyToken{Key: "ary", Expected: true}, AfterLastIndexToken{}}},
	{"/-/key", []Token{RootToken{}, KeyToken{Key: "-", Expected: true}, KeyToken{Key: "key"}}},

	// Matching index token
	{"/name=val", []Token{RootToken{}, MatchingIndexToken{Key: "name", Value: "val"}}},
	{"/=", []Token{RootToken{}, MatchingIndexToken{Key: "", Value: ""}}},
	{"/name=", []Token{RootToken{}, MatchingIndexToken{Key: "name", Value: ""}}},
	{"/=val", []Token{RootToken{}, MatchingIndexToken{Key: "", Value: "val"}}},
	{"/==", []Token{RootToken{}, MatchingIndexToken{Key: "", Value: "="}}},
}

var _ = Describe("NewPointer", func() {
	It("panics if no tokens are given", func() {
		Expect(func() { NewPointer([]Token{}) }).To(Panic())
	})

	It("panics if first token is not root token", func() {
		Expect(func() { NewPointer([]Token{IndexToken{}}) }).To(Panic())
	})

	It("succeeds for basic case", func() {
		Expect(NewPointer([]Token{RootToken{}}).Tokens()).To(Equal([]Token{RootToken{}}))
	})
})

var _ = Describe("NewPointerFromString", func() {
	It("returns error if string doesn't start with /", func() {
		_, err := NewPointerFromString("abc")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected to start with '/'"))
	})
})

var _ = Describe("Pointer.String", func() {
	for _, tc := range testCases {
		tc := tc // copy
		It(fmt.Sprintf("'%#v' results in '%s'", tc.Tokens, tc.String), func() {
			Expect(NewPointer(tc.Tokens).String()).To(Equal(tc.String))
		})
	}
})

var _ = Describe("Pointer.Tokens", func() {
	parsingTestCases := []PointerTestCase{
		{"/key/key2?", []Token{RootToken{}, KeyToken{Key: "key", Expected: true}, KeyToken{Key: "key2"}}},
	}

	parsingTestCases = append(parsingTestCases, testCases...)

	for _, tc := range parsingTestCases {
		tc := tc // copy
		It(fmt.Sprintf("'%s' results in '%#v'", tc.String, tc.Tokens), func() {
			Expect(MustNewPointerFromString(tc.String).Tokens()).To(Equal(tc.Tokens))
		})
	}
})
