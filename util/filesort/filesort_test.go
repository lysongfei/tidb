package filesort

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/tidb/sessionctx/variable"
	"github.com/pingcap/tidb/util/testleak"
	"github.com/pingcap/tidb/util/types"
)

func TestT(t *testing.T) {
	CustomVerboseFlag = true
	TestingT(t)
}

var _ = Suite(&testFileSortSuite{})

type testFileSortSuite struct {
}

func nextRow(r *rand.Rand, keySize int, valSize int) (key []types.Datum, val []types.Datum, handle int64) {
	key = make([]types.Datum, keySize)
	for i, _ := range key {
		key[i] = types.NewDatum(r.Int())
	}

	val = make([]types.Datum, valSize)
	for j, _ := range val {
		val[j] = types.NewDatum(r.Int())
	}

	handle = r.Int63()
	return
}

func (s *testFileSortSuite) TestSingleFile(c *C) {
	defer testleak.AfterTest(c)()

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	sc := new(variable.StatementContext)
	keySize := r.Intn(10) + 1 // random int in range [1, 10]
	valSize := r.Intn(20) + 1 // random int in range [1, 20]
	bufSize := 40             // hold up to 40 items per file
	byDesc := make([]bool, keySize)
	for i, _ := range byDesc {
		byDesc[i] = (r.Intn(2) == 0)
	}

	var (
		err  error
		fs   *FileSorter
		pkey []types.Datum
		key  []types.Datum
	)

	fs, err = NewFileSorter(sc, keySize, valSize, bufSize, byDesc)
	c.Assert(err, IsNil)

	nRows := r.Intn(bufSize-1) + 1 // random int in range [1, bufSize - 1]
	for i := 1; i <= nRows; i++ {
		err = fs.Input(nextRow(r, keySize, valSize))
		c.Assert(err, IsNil)
	}

	pkey, _, _, err = fs.Output()
	c.Assert(err, IsNil)
	for i := 1; i < nRows; i++ {
		key, _, _, err = fs.Output()
		c.Assert(err, IsNil)
		c.Assert(lessThan(sc, key, pkey, byDesc), IsFalse)
		pkey = key
	}
}

func (s *testFileSortSuite) TestMultipleFiles(c *C) {
	defer testleak.AfterTest(c)()

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	sc := new(variable.StatementContext)
	keySize := r.Intn(10) + 1 // random int in range [1, 10]
	valSize := r.Intn(20) + 1 // random int in range [1, 20]
	bufSize := 40             // hold up to 40 items per file
	byDesc := make([]bool, keySize)
	for i, _ := range byDesc {
		byDesc[i] = (r.Intn(2) == 0)
	}

	var (
		err  error
		fs   *FileSorter
		pkey []types.Datum
		key  []types.Datum
	)

	fs, err = NewFileSorter(sc, keySize, valSize, bufSize, byDesc)
	c.Assert(err, IsNil)

	nRows := (r.Intn(bufSize) + 1) * (r.Intn(10) + 2)
	for i := 1; i <= nRows; i++ {
		err = fs.Input(nextRow(r, keySize, valSize))
		c.Assert(err, IsNil)
	}

	pkey, _, _, err = fs.Output()
	c.Assert(err, IsNil)
	for i := 1; i < nRows; i++ {
		key, _, _, err = fs.Output()
		c.Assert(err, IsNil)
		c.Assert(lessThan(sc, key, pkey, byDesc), IsFalse)
		pkey = key
	}
}