// This file is automatically generated. Do not edit - make changes to relevant got file.

// Copyright ©2011-2012 The bíogo Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package align

import (
	"code.google.com/p/biogo/alphabet"
	"code.google.com/p/biogo/feat"

	"fmt"
	"os"
	"text/tabwriter"
)

//line nw_type.got:17
func drawNWTableLetters(rSeq, qSeq alphabet.Letters, index alphabet.Index, table []int, a [][]int) {
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Printf("rSeq: %s\n", rSeq)
	fmt.Printf("qSeq: %s\n", qSeq)
	fmt.Fprint(tw, "\tqSeq\t")
	for _, l := range qSeq {
		fmt.Fprintf(tw, "%c\t", l)
	}
	fmt.Fprintln(tw)

	r, c := rSeq.Len()+1, qSeq.Len()+1
	fmt.Fprint(tw, "rSeq\t")
	for i := 0; i < r; i++ {
		if i != 0 {
			fmt.Fprintf(tw, "%c\t", rSeq[i-1])
		}

		for j := 0; j < c; j++ {
			pr := pointerNWRuneLetters(rSeq, qSeq, i, j, table, index, a, c)
			if pr != ' ' {
				fmt.Fprintf(tw, "%c % 3v\t", pr, table[i*c+j])
			} else {
				fmt.Fprintf(tw, "%v\t", table[i*c+j])
			}
		}
		fmt.Fprintln(tw)
	}
	tw.Flush()
}

func pointerNWRuneLetters(rSeq, qSeq alphabet.Letters, i, j int, table []int, index alphabet.Index, a [][]int, c int) rune {
	switch {
	case i == 0 && j == 0:
		return ' '
	case i == 0:
		return '⬅'
	case j == 0:
		return '⬆'
	}
	rVal := index[rSeq[i-1]]
	qVal := index[qSeq[j-1]]
	if rVal < 0 || qVal < 0 {
		return ' '
	} else {
		p := i*c + j
		gap := len(a) - 1
		switch table[p] {
		case table[p-c-1] + a[rVal][qVal]:
			return '⬉'
		case table[p-c] + a[rVal][gap]:
			return '⬆'
		case table[p-1] + a[gap][qVal]:
			return '⬅'
		default:
			return ' '
		}
	}
}

func (a NW) alignLetters(rSeq, qSeq alphabet.Letters, alpha alphabet.Alphabet) ([]feat.Pair, error) {
	gap := len(a) - 1
	for _, row := range a {
		if len(row) != gap+1 {
			return nil, ErrMatrixNotSquare
		}
	}

	index := alpha.LetterIndex()
	r, c := rSeq.Len()+1, qSeq.Len()+1
	table := make([]int, r*c)
	for j := range table[1:c] {
		table[j+1] = table[j] + a[gap][index[qSeq[j]]]
	}
	for i := 1; i < r; i++ {
		table[i*c] = table[(i-1)*c] + a[index[rSeq[i-1]]][gap]
	}

	var scores [3]int
	for i := 1; i < r; i++ {
		for j := 1; j < c; j++ {
			var (
				rVal = index[rSeq[i-1]]
				qVal = index[qSeq[j-1]]
			)
			if rVal < 0 || qVal < 0 {
				continue
			} else {
				p := i*c + j
				scores = [3]int{
					diag: table[p-c-1] + a[rVal][qVal],
					up:   table[p-c] + a[rVal][gap],
					left: table[p-1] + a[gap][qVal],
				}
				table[p] = max(&scores)
			}
		}
	}
	if debugNeedle {
		drawNWTableLetters(rSeq, qSeq, index, table, a)
	}

	var aln []feat.Pair
	score, last := 0, diag
	i, j := r-1, c-1
	maxI, maxJ := i, j
	for i > 0 && j > 0 {
		var (
			rVal = index[rSeq[i-1]]
			qVal = index[qSeq[j-1]]
		)
		if rVal < 0 || qVal < 0 {
			continue
		} else {
			p := i*c + j
			switch table[p] {
			case table[p-c-1] + a[rVal][qVal]:
				if last != diag {
					aln = append(aln, &featPair{
						a:     feature{start: i, end: maxI},
						b:     feature{start: j, end: maxJ},
						score: score,
					})
					maxI, maxJ = i, j
					score = 0
				}
				score += table[p] - table[p-c-1]
				i--
				j--
				last = diag
			case table[p-c] + a[rVal][gap]:
				if last != up {
					aln = append(aln, &featPair{
						a:     feature{start: i, end: maxI},
						b:     feature{start: j, end: maxJ},
						score: score,
					})
					maxI, maxJ = i, j
					score = 0
				}
				score += table[p] - table[p-c]
				i--
				last = up
			case table[p-1] + a[gap][qVal]:
				if last != left {
					aln = append(aln, &featPair{
						a:     feature{start: i, end: maxI},
						b:     feature{start: j, end: maxJ},
						score: score,
					})
					maxI, maxJ = i, j
					score = 0
				}
				score += table[p] - table[p-1]
				j--
				last = left
			default:
				panic(fmt.Sprintf("align: nw internal error: no path at row: %d col:%d\n", i, j))
			}
		}
	}

	aln = append(aln, &featPair{
		a:     feature{start: i, end: maxI},
		b:     feature{start: j, end: maxJ},
		score: score,
	})
	if i != j {
		aln = append(aln, &featPair{
			a:     feature{start: 0, end: i},
			b:     feature{start: 0, end: j},
			score: table[i*c+j],
		})
	}

	for i, j := 0, len(aln)-1; i < j; i, j = i+1, j-1 {
		aln[i], aln[j] = aln[j], aln[i]
	}

	return aln, nil
}
