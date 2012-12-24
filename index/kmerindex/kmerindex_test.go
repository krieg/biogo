// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kmerindex

import (
	"code.google.com/p/biogo/exp/alphabet"
	"code.google.com/p/biogo/exp/seq/linear"
	"code.google.com/p/biogo/util"
	check "launchpad.net/gocheck"
	"math/rand"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	*linear.Seq
}

var _ = check.Suite(&S{})

var testLen = 1000

func (s *S) SetUpSuite(c *check.C) {
	MaxKmerLen = 14
	s.Seq = &linear.Seq{}
	s.Seq.Seq = make(alphabet.Letters, testLen)
	for i := range s.Seq.Seq {
		s.Seq.Seq[i] = [...]alphabet.Letter{'A', 'C', 'G', 'T'}[rand.Int()%4]
	}
}

func (s *S) TestKmerIndexCheck(c *check.C) {
	for k := MinKmerLen; k <= MaxKmerLen; k++ {
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			ok, _ := i.Check()
			c.Check(ok, check.Equals, false)
			i.Build()
			ok, f := i.Check()
			c.Check(f, check.Equals, s.Seq.Len()-k+1)
			c.Check(ok, check.Equals, true)
		}
	}
}

func (s *S) TestKmerFrequencies(c *check.C) {
	for k := MinKmerLen; k <= MaxKmerLen; k++ {
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			freqs, ok := i.KmerFrequencies()
			c.Check(ok, check.Equals, true)
			hashFreqs := make(map[string]int)
			for i := 0; i+k <= s.Seq.Len(); i++ {
				hashFreqs[string(alphabet.LettersToBytes(s.Seq.Seq[i:i+k]))]++
			}
			for key := range freqs {
				if freqs[key] != hashFreqs[i.Stringify(key)] {
					c.Logf("seq %s\n", s.Seq)
					c.Logf("key %x, string of %q\n", key, i.Stringify(key))
				}
				c.Check(freqs[key], check.Equals, hashFreqs[i.Stringify(key)])
			}
			for key := range hashFreqs {
				if keyKmer, err := i.KmerOf(key); err != nil {
					c.Fatal(err)
				} else {
					if freqs[keyKmer] != hashFreqs[key] {
						c.Logf("seq %s\n", s.Seq)
						c.Logf("keyKmer %x, string of %q, key %q\n", keyKmer, i.Stringify(keyKmer), key)
					}
					c.Check(freqs[keyKmer], check.Equals, hashFreqs[key])
				}
			}
		}
	}
}

func (s *S) TestKmerPositions(c *check.C) {
	for k := MinKmerLen; k < MaxKmerLen; k++ { // don't test full range to time's sake
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			i.Build()
			hashPos := make(map[string][]int)
			for i := 0; i+k <= s.Seq.Len(); i++ {
				hashPos[string(alphabet.LettersToBytes(s.Seq.Seq[i:i+k]))] = append(hashPos[string(alphabet.LettersToBytes(s.Seq.Seq[i:i+k]))], i)
			}
			pos, ok := i.KmerIndex()
			c.Check(ok, check.Equals, true)
			for p := range pos {
				c.Check(pos[p], check.DeepEquals, hashPos[i.Stringify(p)])
			}
		}
	}
}

func (s *S) TestKmerPositionsString(c *check.C) {
	for k := MinKmerLen; k < MaxKmerLen; k++ { // don't test full range to time's sake
		if i, err := New(k, s.Seq); err != nil {
			c.Fatalf("New KmerIndex failed: %v", err)
		} else {
			i.Build()
			hashPos := make(map[string][]int)
			for i := 0; i+k <= s.Seq.Len(); i++ {
				hashPos[string(alphabet.LettersToBytes(s.Seq.Seq[i:i+k]))] = append(hashPos[string(alphabet.LettersToBytes(s.Seq.Seq[i:i+k]))], i)
			}
			pos, ok := i.StringKmerIndex()
			c.Check(ok, check.Equals, true)
			for p := range pos {
				c.Check(pos[p], check.DeepEquals, hashPos[p])
			}
		}
	}
}

func (s *S) TestKmerKmerUtilities(c *check.C) {
	for k := MinKmerLen; k <= 8; k++ { // again not testing all exhaustively
		for kmer := Kmer(0); uint(kmer) <= util.Pow4(k)-1; kmer++ {
			// Interconversion between string and Kmer
			if rk, err := KmerOf(k, Stringify(k, kmer)); err != nil {
				c.Fatalf("Failed Kmer conversion: %v", err)
			} else {
				c.Check(rk, check.Equals, kmer)
			}

			// Complementation
			dc := ComplementOf(k, ComplementOf(k, kmer))
			if dc != kmer {
				c.Logf("kmer: %s\ndouble complement: %s\n", Stringify(k, kmer), Stringify(k, dc))
			}
			c.Check(dc, check.Equals, kmer)

			// GC content
			ks := Stringify(k, kmer)
			gc := 0
			for _, b := range ks {
				if b == 'G' || b == 'C' {
					gc++
				}
			}
			c.Check(GCof(k, kmer), check.Equals, float64(gc)/float64(k))
		}
	}
}
