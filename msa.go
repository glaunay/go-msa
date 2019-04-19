package msago

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"sync"
)

// Single record container
type MsaRecord struct {
	name     string
	Sequence []byte
}

// Top container
type Msa struct {
	// r io.Reader
	source  string
	nRow    int
	nCol    int
	asMap   map[string]*MsaRecord
	Records []*MsaRecord
}

// Readin/Factory functions
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func ParseFile(filePath []byte) Msa {
	file, err := os.Open(string(filePath))
	check(err)
	defer file.Close()

	newMsa := Msa{source: string(filePath), asMap: make(map[string]*MsaRecord)}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		cName, cSeq, empty := lineParser(text)
		if empty {
			continue
		}
		newMsa.glob(cName, cSeq)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return newMsa
}

/* MSA container level */
func (msa *Msa) glob(name string, subSeq string) {
	var newRecord *MsaRecord

	newRecord, ok := msa.asMap[name]
	if !ok {
		/*
			fmt.Println(`globing new ::`, name)
			fmt.Println(`globing new ::`, subSeq)
		*/
		newRecord = &(MsaRecord{name: name, Sequence: []byte{}})
		msa.asMap[name] = newRecord
		//fmt.Println("Inital record content:::\n", msa.asMap[name], "\n#########\n")
		msa.Records = append(msa.Records, newRecord)
		msa.nRow++
	}

	newRecord.Sequence = append(newRecord.Sequence, []byte(subSeq)...)
	//fmt.Println("current record content:::\n", msa.asMap[name], "\n#########\n")

	//append(msa.asMap[name].sequence, subSeq...)
}

func (msa *Msa) Len() int {
	return msa.nRow
}

func (msa *Msa) Iterator() (func() (*MsaRecord, bool), bool) {
	n := -1
	// closure captures variable n
	return func() (*MsaRecord, bool) {
		n += 1
		return msa.Records[n], n < msa.nRow-1
	}, msa.nRow > 0
}

func (msa *Msa) MapSearch(predicate func(string, string) bool) Msa {

	newMsa := Msa{source: msa.source, asMap: make(map[string]*MsaRecord)}

	var waitgroup sync.WaitGroup
	status := make([]bool, msa.nRow, msa.nRow)
	for i := 0; i < msa.nRow; i++ {
		waitgroup.Add(1)
		go func(w *sync.WaitGroup, _i *bool, r *MsaRecord) {
			// Filter gap out of seq
			seq := make([]byte, 0, len(r.Sequence))
			for _, s := range r.Sequence {
				if s != '-' {
					seq = append(seq, s)
				}
			}

			*_i = predicate(r.name, string(seq))
			w.Done()
		}(&waitgroup, &status[i], msa.Records[i])
	}
	waitgroup.Wait()

	for i := 0; i < msa.nRow; i++ {
		if !status[i] {
			continue
		}
		newMsa.nRow++
		n := msa.Records[i].copy()
		newMsa.Records = append(newMsa.Records, n)
		newMsa.asMap[n.name] = n
	}

	return newMsa
}

/*
func (msa *Msa) seekName(motif string) {
	newMsa := Msa{source: msa.source, asMap: make(map[string]*MsaRecord)}
	for v := range (0, msa.nRow) {
        fmt.Println(v)
    }

}*/

func lineParser(lineString string) (string, string, bool) {
	regExp := regexp.MustCompile("(^[\\S]+)[\\s]+([\\S]+)$")
	match := regExp.FindStringSubmatch(lineString)
	if match == nil {
		return "", "", true
	} else {
		return match[1], match[2], false
	}
}

/* record level */
func (x *MsaRecord) String() string {
	return fmt.Sprintf(">%s\n%s", x.name, x.Sequence)
}

func (x *MsaRecord) copy() *MsaRecord {
	target := MsaRecord{name: x.name, Sequence: append([]byte(nil), x.Sequence...)}
	return &target
}
