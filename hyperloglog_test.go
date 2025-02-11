package hyperloglog

import (
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"slices"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func estimateError(got, exp uint64) float64 {
	var delta uint64
	if got > exp {
		delta = got - exp
	} else {
		delta = exp - got
	}
	return float64(delta) / float64(exp)
}

func nopHash(buf []byte) uint64 {
	if len(buf) != 8 {
		panic(fmt.Sprintf("unexpected size buffer: %d", len(buf)))
	}
	return binary.BigEndian.Uint64(buf)
}

func TestHLL_CardinalityHashed(t *testing.T) {
	hlltc, err := NewSketch(14, true)
	require.NoError(t, err)

	step := 10
	unique := map[string]bool{}

	for i := 1; len(unique) <= 10000000; i++ {
		str := fmt.Sprintf("flow-%d", i)
		hlltc.Insert([]byte(str))
		unique[str] = true

		if len(unique)%step == 0 {
			step *= 5
			exact := uint64(len(unique))
			res := uint64(hlltc.Estimate())
			ratio := 100 * estimateError(res, exact)
			require.LessOrEqual(t, ratio, 2.0, "Exact %d, got %d which is %.2f%% error", exact, res, ratio)
		}
	}
	exact := uint64(len(unique))
	res := uint64(hlltc.Estimate())
	ratio := 100 * estimateError(res, exact)
	require.LessOrEqual(t, ratio, 2.0, "Exact %d, got %d which is %.2f%% error", exact, res, ratio)
}

func toByte(v uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	return buf[:]
}

func TestHLL_Add_NoSparse(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk := New16NoSparse()

	sk.Insert(toByte(0x00010fffffffffff))
	n := sk.regs[1]
	require.EqualValues(t, 5, n)

	sk.Insert(toByte(0x0002ffffffffffff))
	n = sk.regs[2]
	require.EqualValues(t, 1, n)

	sk.Insert(toByte(0x0003000000000000))
	n = sk.regs[3]
	require.EqualValues(t, 49, n)

	sk.Insert(toByte(0x0003000000000001))
	n = sk.regs[3]
	require.EqualValues(t, 49, n)

	sk.Insert(toByte(0xff03700000000000))
	n = sk.regs[0xff03]
	require.EqualValues(t, 2, n)

	sk.Insert(toByte(0xff03080000000000))
	n = sk.regs[0xff03]
	require.EqualValues(t, 5, n)
}

func TestHLL_Precision_NoSparse(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk, _ := NewSketch(4, false)

	sk.Insert(toByte(0x1fffffffffffffff))
	n := sk.regs[1]
	require.EqualValues(t, 1, n)

	sk.Insert(toByte(0xffffffffffffffff))
	n = sk.regs[0xf]
	require.EqualValues(t, 1, n)

	sk.Insert(toByte(0x00ffffffffffffff))
	n = sk.regs[0]
	require.EqualValues(t, 5, n)
}

func TestHLL_toNormal(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk := NewTestSketch(16)
	sk.Insert(toByte(0x00010fffffffffff))
	sk.toNormal()
	c := sk.Estimate()
	require.EqualValues(t, 1, c)

	require.False(t, sk.sparse(), "toNormal should convert to normal")

	sk = NewTestSketch(16)
	sk.Insert(toByte(0x00010fffffffffff))
	sk.Insert(toByte(0x0002ffffffffffff))
	sk.Insert(toByte(0x0003000000000000))
	sk.Insert(toByte(0x0003000000000001))
	sk.Insert(toByte(0xff03700000000000))
	sk.Insert(toByte(0xff03080000000000))
	sk.mergeSparse()
	sk.toNormal()

	n := sk.regs[1]
	require.EqualValues(t, 5, n)
	n = sk.regs[2]
	require.EqualValues(t, 1, n)
	n = sk.regs[3]
	require.EqualValues(t, 49, n)
	n = sk.regs[0xff03]
	require.EqualValues(t, 5, n)
}

func TestHLL_Cardinality(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk := NewTestSketch(16)

	n := sk.Estimate()
	require.EqualValues(t, 0, n)

	sk.Insert(toByte(0x00010fffffffffff))
	sk.Insert(toByte(0x00020fffffffffff))
	sk.Insert(toByte(0x00030fffffffffff))
	sk.Insert(toByte(0x00040fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))

	n = sk.Estimate()
	require.EqualValues(t, 5, n)

	// not mutated, still returns correct count
	n = sk.Estimate()
	require.EqualValues(t, 5, n)

	sk.Insert(toByte(0x00060fffffffffff))

	// mutated
	n = sk.Estimate()
	require.EqualValues(t, 6, n)
}

func TestHLL_Merge_Error(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk := NewTestSketch(16)
	sk2 := NewTestSketch(10)

	err := sk.Merge(sk2)
	require.Error(t, err, "different precision should return error")
}

func TestHLL_Merge_Sparse(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk := NewTestSketch(16)
	sk.Insert(toByte(0x00010fffffffffff))
	sk.Insert(toByte(0x00020fffffffffff))
	sk.Insert(toByte(0x00030fffffffffff))
	sk.Insert(toByte(0x00040fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))
	sk.Insert(toByte(0x00050fffffffffff))

	sk2 := NewTestSketch(16)
	require.NoError(t, sk2.Merge(sk))
	n := sk2.Estimate()
	require.EqualValues(t, 5, n)

	require.True(t, sk2.sparse(), "Merge should convert to normal")
	require.True(t, sk.sparse(), "Merge should not modify argument")

	require.NoError(t, sk2.Merge(sk))
	n = sk2.Estimate()
	require.EqualValues(t, 5, n)

	sk.Insert(toByte(0x00060fffffffffff))
	sk.Insert(toByte(0x00070fffffffffff))
	sk.Insert(toByte(0x00080fffffffffff))
	sk.Insert(toByte(0x00090fffffffffff))
	sk.Insert(toByte(0x000a0fffffffffff))
	sk.Insert(toByte(0x000a0fffffffffff))
	n = sk.Estimate()
	require.EqualValues(t, 10, n)

	require.NoError(t, sk2.Merge(sk))
	n = sk2.Estimate()
	require.EqualValues(t, 10, n)
}

func TestHLL_Merge_Complex(t *testing.T) {
	sk1, err := NewSketch(14, true)
	require.NoError(t, err)
	sk2, err := NewSketch(14, true)
	require.NoError(t, err)
	sk3, err := NewSketch(14, true)
	require.NoError(t, err)

	unique := map[string]bool{}

	for i := 1; len(unique) <= 10000000; i++ {
		str := fmt.Sprintf("flow-%d", i)
		sk1.Insert([]byte(str))
		if i%2 == 0 {
			sk2.Insert([]byte(str))
		}
		unique[str] = true
	}

	exact1 := uint64(len(unique))
	res1 := uint64(sk1.Estimate())
	ratio := 100 * estimateError(res1, exact1)
	require.LessOrEqual(t, ratio, 2.0, "Exact %d, got %d which is %.2f%% error", exact1, res1, ratio)

	exact2 := uint64(len(unique)) / 2
	res2 := uint64(sk1.Estimate())
	ratio = 100 * estimateError(res1, exact1)
	require.LessOrEqual(t, ratio, 2.0, "Exact %d, got %d which is %.2f%% error", exact2, res2, ratio)

	require.NoError(t, sk2.Merge(sk1))
	exact2 = uint64(len(unique))
	res2 = uint64(sk2.Estimate())
	ratio = 100 * estimateError(res1, exact1)
	require.LessOrEqual(t, ratio, 2.0, "Exact %d, got %d which is %.2f%% error", exact2, res2, ratio)

	for i := 1; i <= 500000; i++ {
		str := fmt.Sprintf("stream-%d", i)
		sk2.Insert([]byte(str))
		unique[str] = true
	}

	require.NoError(t, sk2.Merge(sk3))
	exact2 = uint64(len(unique))
	res2 = uint64(sk2.Estimate())
	ratio = 100 * estimateError(res1, exact1)
	require.LessOrEqual(t, ratio, 1.0, "Exact %d, got %d which is %.2f%% error", exact2, res2, ratio)
}

func TestHLL_EncodeDecode(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk := NewTestSketch(8)
	i, r := decodeHash(encodeHash(0xffffff8000000000, sk.p, pp), sk.p, pp)
	require.EqualValues(t, 0xff, i)
	require.EqualValues(t, 1, r)

	i, r = decodeHash(encodeHash(0xff00000000000000, sk.p, pp), sk.p, pp)
	require.EqualValues(t, 0xff, i)
	require.EqualValues(t, 57, r)

	i, r = decodeHash(encodeHash(0xff30000000000000, sk.p, pp), sk.p, pp)
	require.EqualValues(t, 0xff, i)
	require.EqualValues(t, 3, r)

	i, r = decodeHash(encodeHash(0xaa10000000000000, sk.p, pp), sk.p, pp)
	require.EqualValues(t, 0xaa, i)
	require.EqualValues(t, 4, r)

	i, r = decodeHash(encodeHash(0xaa0f000000000000, sk.p, pp), sk.p, pp)
	require.EqualValues(t, 0xaa, i)
	require.EqualValues(t, 5, r)
}

func TestHLL_Error(t *testing.T) {
	_, err := NewSketch(3, true)
	require.Error(t, err, "precision 3 should return error")

	_, err = NewSketch(18, true)
	require.NoError(t, err)

	_, err = NewSketch(19, true)
	require.Error(t, err, "precision 19 should return error")
}

func TestHLL_Marshal_Unmarshal_Sparse(t *testing.T) {
	sk, _ := NewSketch(4, true)
	sk.tmpSet = makeSet(2)
	sk.tmpSet.add(26)
	sk.tmpSet.add(40)

	// Add a bunch of values to the sparse representation.
	for i := 0; i < 10; i++ {
		sk.sparseList.Append(uint32(rand.Int()))
	}

	data, err := sk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	// Peeking at the first byte should reveal the version.
	if got, exp := data[0], byte(version); got != exp {
		t.Fatalf("got byte %v, expected %v", got, exp)
	}

	var res Sketch
	if err := res.UnmarshalBinary(data); err != nil {
		t.Fatal(err)
	}

	// reflect.DeepEqual will always return false when comparing non-nil
	// functions, so we'll set them to nil.
	if got, exp := &res, sk; !reflect.DeepEqual(got, exp) {
		t.Fatalf("got %v, wanted %v", spew.Sdump(got), spew.Sdump(exp))
	}
}

func TestHLL_Marshal_Unmarshal_Dense(t *testing.T) {
	sk, _ := NewSketch(4, false)

	// Add a bunch of values to the dense representation.
	for i := uint32(0); i < 10; i++ {
		sk.regs[i] = uint8(rand.Int())
	}

	data, err := sk.MarshalBinary()
	require.NoError(t, err)

	// Peeking at the first byte should reveal the version.
	require.EqualValues(t, byte(version), data[0], "got byte %v, expected %v", data[0], byte(version))

	var res Sketch
	require.NoError(t, res.UnmarshalBinary(data))

	// reflect.DeepEqual will always return false when comparing non-nil
	// functions, so we'll set them to nil.
	require.True(t, reflect.DeepEqual(&res, sk), "got %v, wanted %v", spew.Sdump(&res), spew.Sdump(sk))
}

// Tests that a sketch can be serialised / unserialised and keep an accurate
// cardinality estimate.
func TestHLL_Marshal_Unmarshal_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	count := make(map[string]struct{}, 1000000)
	sk, _ := NewSketch(16, true)

	buf := make([]byte, 8)
	for i := 0; i < 1000000; i++ {
		if _, err := crand.Read(buf); err != nil {
			panic(err)
		}

		count[string(buf)] = struct{}{}

		// Add to the sketch.
		sk.Insert(buf)
	}

	gotC := sk.Estimate()
	epsilon := 15000 // 1.5%
	require.LessOrEqual(t, math.Abs(float64(int(gotC)-len(count))), float64(epsilon), "error was %v for estimation %d and true cardinality %d", math.Abs(float64(int(gotC)-len(count))), gotC, len(count))

	// Serialise the sketch.
	sketch, err := sk.MarshalBinary()
	require.NoError(t, err)

	// Deserialise.
	sk = &Sketch{}
	require.NoError(t, sk.UnmarshalBinary(sketch))

	// The count should be the same
	oldC := gotC
	require.EqualValues(t, oldC, sk.Estimate())

	// Add some more values.
	for i := 0; i < 1000000; i++ {
		if _, err := crand.Read(buf); err != nil {
			panic(err)
		}

		count[string(buf)] = struct{}{}

		// Add to the sketch.
		sk.Insert(buf)
	}

	// The sketch should still be working correctly.
	gotC = sk.Estimate()
	epsilon = 30000 // 1.5%
	require.LessOrEqual(t, math.Abs(float64(int(gotC)-len(count))), float64(epsilon), "error was %v for estimation %d and true cardinality %d", math.Abs(float64(int(gotC)-len(count))), gotC, len(count))
}

// Tests that a sketch will be used in Unmarshal if it is unused
func TestHLL_Marshal_Unmarshal_Reuse(t *testing.T) {
	sk, _ := NewSketch(4, true)
	// Add a bunch of values to the sparse representation.
	for i := 0; i < 10; i++ {
		sk.sparseList.Append(uint32(rand.Int()))
	}
	data, err := sk.MarshalBinary()
	require.NoError(t, err)
	res, _ := NewSketch(4, true)
	// Change the "m" here because it's not adjusted so it'll allow us to
	// determine if newSketch was called
	res.m = 1
	require.NoError(t, res.UnmarshalBinary(data))

	// Compare the "m" to make sure it's the same
	require.EqualValues(t, 1, res.m, "UnmarshalBinary created a newSketch Sketch")

	// If we re-use the same sketch, newSketch should be called
	require.NoError(t, res.UnmarshalBinary(data))

	// Compare the "m" to make sure it was changed
	require.NotEqual(t, 1, res.m, "UnmarshalBinary did not create a newSketch Sketch")
}

func TestHLL_Unmarshal_ErrorTooShort(t *testing.T) {
	require.EqualValues(t, ErrorTooShort, (&Sketch{}).UnmarshalBinary(nil), "UnmarshalBinary(nil) should fail with ErrorTooShort")

	b := []byte{
		// precision:14, sparse:true, tmpSet:empty,
		// sparseList:{count:1, last:0, sz:1, ...}
		0x01, 0x0e, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x7f,
	}
	require.NoError(t, (&Sketch{}).UnmarshalBinary(b))
	for i := 0; i < len(b)-1; i++ {
		sk := &Sketch{}
		err := sk.UnmarshalBinary(b[0:i])
		require.EqualValues(t, ErrorTooShort, err, "should fail for incomplete bytes: i=%d", i)
	}
}

func TestHLL_AppendBinary(t *testing.T) {
	sk := NewTestSketch(16)
	for i := 0; i < 10; i++ {
		sk.InsertHash(uint64(rand.Int()))
	}
	data1, err := sk.MarshalBinary()
	require.NoError(t, err)

	bufSize := rand.Intn(100)
	buf := make([]byte, bufSize)
	for i := range buf {
		buf[i] = byte(rand.Intn(256))
	}
	bufCopy := slices.Clone(buf)
	data2, err := sk.AppendBinary(buf)
	require.NoError(t, err)

	require.Len(t, data2, len(data1)+len(bufCopy))
	require.Equal(t, bufCopy, data2[:len(bufCopy)])
	require.Equal(t, data1, data2[len(bufCopy):])
}

func Benchmark_HLL_Marshal(b *testing.B) {
	run := func(precision uint8, sparse bool) {
		name := fmt.Sprintf("precision%d_", precision)
		if sparse {
			name += "sparse"
		} else {
			name += "dense"
		}
		b.Run(name, func(b *testing.B) {
			sk, _ := NewSketch(precision, sparse)
			for i := 0; i < 1000; i++ {
				sk.InsertHash(uint64(rand.Int()))
			}
			b.Run("MarshalBinary", func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_, _ = sk.MarshalBinary()
				}
			})
			b.Run("AppendBinary", func(b *testing.B) {
				var buf []byte
				for i := 0; i < b.N; i++ {
					buf, _ = sk.AppendBinary(buf[:0])
				}
			})
		})
	}
	run(16, true)
	run(16, false)
}

func TestHLL_Clone(t *testing.T) {
	sk1 := NewTestSketch(16)

	sk1.Insert(toByte(0x00010fffffffffff))
	sk1.Insert(toByte(0x0002ffffffffffff))
	sk1.Insert(toByte(0x0003000000000000))
	sk1.Insert(toByte(0x000000))
	sk1.Insert(toByte(0x0003000000000001))
	sk1.Insert(toByte(0xff03700000000000))

	n := sk1.Estimate()
	require.EqualValues(t, 6, n)

	sk2 := sk1.Clone()
	require.EqualValues(t, sk1.Estimate(), sk2.Estimate())
	require.True(t, isSketchEqual(sk1, sk2))

	sk1.toNormal()
	sk2 = sk1.Clone()

	require.EqualValues(t, sk1.Estimate(), sk2.Estimate())
	require.True(t, isSketchEqual(sk1, sk2))
}

func TestHLL_Add_Hash(t *testing.T) {
	hash = nopHash
	defer func() {
		hash = hashFunc
	}()
	sk := NewTestSketch(16)

	n := sk.Estimate()
	require.EqualValues(t, 0, n)

	sk.InsertHash(0x00010fffffffffff)
	sk.InsertHash(0x00020fffffffffff)
	sk.InsertHash(0x00030fffffffffff)
	sk.InsertHash(0x00040fffffffffff)
	sk.InsertHash(0x00050fffffffffff)
	sk.InsertHash(0x00050fffffffffff)

	n = sk.Estimate()
	require.EqualValues(t, 5, n)

	sk.toNormal()
	sk.InsertHash(0x10010f00ffffffff)
	sk.InsertHash(0x20020f00ffffffff)
	sk.InsertHash(0x30030f00ffffffff)
	sk.InsertHash(0x40040f00ffffffff)
	sk.InsertHash(0x50050f00ffffffff)
	sk.InsertHash(0x60050f00ffffffff)

	// not mutated, still returns correct count
	n = sk.Estimate()
	require.EqualValues(t, 11, n)

	sk.InsertHash(0x00060fffffffffff)

	// mutated
	n = sk.Estimate()
	require.EqualValues(t, 12, n)
}

func isSketchEqual(sk1, sk2 *Sketch) bool {
	switch {
	case sk1.alpha != sk2.alpha:
		fmt.Printf("alpha mismatch: %f != %f", sk1.alpha, sk2.alpha)
		return false
	case sk1.p != sk2.p:
		fmt.Printf("p mismatch: %d != %d", sk1.p, sk2.p)
		return false
	case sk1.m != sk2.m:
		fmt.Printf("m mismatch: %d != %d", sk1.m, sk2.m)
		return false
	case !reflect.DeepEqual(sk1.sparseList, sk2.sparseList):
		fmt.Printf("sparseList mismatch: %v != %v", sk1.sparseList, sk2.sparseList)
		return false
	case len(sk1.regs) == 0 && len(sk2.regs) == 0:
		// Both are empty, consider them equal
		return true
	case !reflect.DeepEqual(sk1.regs, sk2.regs):
		fmt.Printf("regs mismatch: %v != %v", sk1.regs, sk2.regs)
		return false
	default:
		return true
	}
}

func NewTestSketch(p uint8) *Sketch {
	sk, _ := NewSketch(p, true)
	return sk
}

// Generate random data to add to the sketch.
func genData(num int) [][]byte {
	const dataLen = 8
	out := make([][]byte, num)
	numBytes := dataLen * num
	buf := make([]byte, numBytes)

	// generate random bytes
	n, err := crand.Read(buf)
	if err != nil {
		panic(err)
	} else if n != numBytes {
		panic(fmt.Errorf("only %d bytes generated, expected %d", n, numBytes))
	}

	for i := range out {
		out[i] = buf[:dataLen]
		buf = buf[dataLen:]
	}
	return out
}

// Memoises values to be added to a sketch during a benchmark.
var benchdata = map[int][][]byte{}

func benchmarkAdd(b *testing.B, sk *Sketch, n int) {
	blobs, ok := benchdata[n]
	if !ok {
		// Generate it.
		benchdata[n] = genData(n)
		blobs = benchdata[n]
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(blobs); j++ {
			sk.Insert(blobs[j])
		}
	}
	b.StopTimer()
}

// Report size and allocations of a new sparse HLL
func Benchmark_Size_New_Sparse(b *testing.B) {
	var sk *Sketch
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sk, _ = NewSketch(16, true)
	}
	_ = sk
}

func Benchmark_Add_100(b *testing.B) {
	sk, _ := NewSketch(16, true)
	benchmarkAdd(b, sk, 100)
}

func Benchmark_Add_1000(b *testing.B) {
	sk, _ := NewSketch(16, true)
	benchmarkAdd(b, sk, 1000)
}

func Benchmark_Add_10000(b *testing.B) {
	sk, _ := NewSketch(16, true)
	benchmarkAdd(b, sk, 10000)
}

func Benchmark_Add_100000(b *testing.B) {
	sk, _ := NewSketch(16, true)
	benchmarkAdd(b, sk, 100000)
}

func Benchmark_Add_1000000(b *testing.B) {
	sk, _ := NewSketch(16, true)
	benchmarkAdd(b, sk, 1000000)
}

func Benchmark_Add_10000000(b *testing.B) {
	sk, _ := NewSketch(16, true)
	benchmarkAdd(b, sk, 10000000)
}

func Benchmark_Add_100000000(b *testing.B) {
	sk, _ := NewSketch(16, true)
	benchmarkAdd(b, sk, 100000000)
}

func randStr(n int) string {
	i := rand.Uint32()
	return fmt.Sprintf("a%d %d", i, n)
}

func benchmark(precision uint8, n int) {
	hll, _ := NewSketch(precision, true)

	for i := 0; i < n; i++ {
		s := []byte(randStr(i))
		hll.Insert(s)
		hll.Insert(s)
	}

	e := hll.Estimate()
	var percentErr = func(est uint64) float64 {
		return 100 * math.Abs(float64(n)-float64(est)) / float64(n)
	}

	fmt.Printf("\nReal Cardinality: %8d\n", n)
	fmt.Printf("HyperLogLog     : %8d,   Error: %f%%\n", e, percentErr(e))
}

func BenchmarkHll4(b *testing.B) {
	fmt.Println("")
	benchmark(4, b.N)
}

func BenchmarkHll6(b *testing.B) {
	fmt.Println("")
	benchmark(6, b.N)
}

func BenchmarkHll8(b *testing.B) {
	fmt.Println("")
	benchmark(8, b.N)
}

func BenchmarkHll10(b *testing.B) {
	fmt.Println("")
	benchmark(10, b.N)
}

func BenchmarkHll14(b *testing.B) {
	fmt.Println("")
	benchmark(14, b.N)
}

func BenchmarkHll16(b *testing.B) {
	fmt.Println("")
	benchmark(16, b.N)
}

func BenchmarkZipf(b *testing.B) {
	cases := []struct {
		s    float64 // skew
		bits uint64  // log2 of the maximum zipf value
	}{
		{s: 1.1, bits: 3},
		{s: 1.1, bits: 10},
		{s: 1.1, bits: 64},
		{s: 1.5, bits: 3},
		{s: 1.5, bits: 10},
		{s: 1.5, bits: 64},
		{s: 2, bits: 3},
		{s: 2, bits: 10},
		{s: 2, bits: 64},
		{s: 5, bits: 3},
		{s: 5, bits: 10},
		{s: 5, bits: 64},
	}
	for _, tc := range cases {
		name := fmt.Sprintf("s%g/b%d", tc.s, tc.bits)
		b.Run(name, func(b *testing.B) {
			// Create a local rng using a seed from the global rand.
			rng := rand.New(rand.NewSource(rand.Int63()))
			zipf := rand.NewZipf(rng, tc.s, 1 /* v */, (uint64(1)<<tc.bits)-1)

			sk := New14()
			b.ResetTimer()
			const batchSize = 1000
			for i := 0; i < b.N/batchSize; i++ {
				b.StopTimer()
				// Generate a bunch of random values upfront; we don't want to
				// benchmark the RNG.
				var values [batchSize]uint64
				for j := range values {
					values[j] = zipf.Uint64()
				}
				b.StartTimer()
				var tmp [8]byte
				for _, v := range values {
					tmp[0] = byte(v)
					tmp[1] = byte(v >> 8)
					tmp[2] = byte(v >> 16)
					tmp[3] = byte(v >> 24)
					tmp[4] = byte(v >> 32)
					tmp[5] = byte(v >> 40)
					tmp[6] = byte(v >> 48)
					tmp[7] = byte(v >> 56)
					sk.Insert(tmp[:])
				}
			}
			b.Logf("Result: %d values, estimated cardinality %d", b.N/batchSize*batchSize, sk.Estimate())
		})
	}
}

func TestHLL_Merge_Order(t *testing.T) {
	for _, count1 := range []int{100, 1000, 10000, 100000, 1000000} {
		for _, count2 := range []int{100, 1000, 10000, 100000, 1000000} {
			t.Run(fmt.Sprintf("count1=%d, count2=%d", count1, count2), func(t *testing.T) {
				sk1 := New14()
				sk2 := New14()

				for i := 0; i < count1; i++ {
					sk1.Insert([]byte(fmt.Sprintf("a%d", i)))
				}
				for i := 0; i < count2; i++ {
					sk2.Insert([]byte(fmt.Sprintf("b%d", i)))
				}

				skX := New14()
				skX.Merge(sk1)
				skX.Merge(sk2)

				skY := New14()
				skY.Merge(sk2)
				skY.Merge(sk1)

				require.EqualValues(t, skX.Estimate(), skY.Estimate())
			})
		}
	}
}

func TestHLL_Add_Out_Of_Order(t *testing.T) {
	for _, sz := range []int{100, 1000, 10000, 100000, 1000000} {
		t.Run(fmt.Sprintf("size=%d", sz), func(t *testing.T) {
			data := make([]uint64, sz)
			for i := range data {
				data[i] = uint64(rand.Int63())
			}

			sk1 := New14()
			for _, v := range data {
				sk1.Insert([]byte(fmt.Sprintf("a%d", v)))
			}

			rand.Shuffle(len(data), func(i, j int) {
				data[i], data[j] = data[j], data[i]
			})

			sk2 := New14()
			for _, v := range data {
				sk2.Insert([]byte(fmt.Sprintf("a%d", v)))
			}

			require.EqualValues(t, sk1.Estimate(), sk2.Estimate())
		})
	}
}

func benchmarkMerge(b *testing.B, size1, size2 int) {
	// Generate data for first sketch
	sk1 := New14()
	for i := 0; i < size1; i++ {
		sk1.Insert([]byte(fmt.Sprintf("a%d", i)))
	}

	// Generate data for second sketch
	sk2 := New14()
	for i := 0; i < size2; i++ {
		sk2.Insert([]byte(fmt.Sprintf("b%d", i)))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		sk := New14()
		sk.Merge(sk1)
		sk.Merge(sk2)
	}
}

func Benchmark_Merge(b *testing.B) {
	sizes := []int{100, 10000, 1000000}

	for _, size1 := range sizes {
		for _, size2 := range sizes {
			name := fmt.Sprintf("size1=%d/size2=%d", size1, size2)
			b.Run(name, func(b *testing.B) {
				benchmarkMerge(b, size1, size2)
			})
		}
	}
}
