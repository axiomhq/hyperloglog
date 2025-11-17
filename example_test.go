package hyperloglog

import (
	"fmt"
	"log"
)

// Example demonstrates basic usage.
func Example() {
	sketch := New()

	sketch.Insert([]byte("alice"))
	sketch.Insert([]byte("bob"))
	sketch.Insert([]byte("charlie"))
	sketch.Insert([]byte("alice")) // Duplicate

	fmt.Printf("Estimated unique elements: %d\n", sketch.Estimate())

	// Output:
	// Estimated unique elements: 3
}

// Example_serialization demonstrates binary marshaling and error handling.
func Example_serialization() {
	original := New14()
	for i := range 100 {
		buf := fmt.Appendf(nil, "user-%d", i)
		original.Insert(buf)
	}

	data, err := original.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	restored := New14()
	if err := restored.UnmarshalBinary(data); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Original estimate: %d\n", original.Estimate())
	fmt.Printf("Restored estimate: %d\n", restored.Estimate())
	fmt.Println("Serialization successful")

	// Output:
	// Original estimate: 100
	// Restored estimate: 100
	// Serialization successful
}

// Example_merge shows merging sketches for distributed counting.
func Example_merge() {
	sketch1 := New14()
	sketch2 := New14()

	for i := range 500 {
		buf := fmt.Appendf(nil, "item-%d", i)
		sketch1.Insert(buf)
	}

	for i := range 500 {
		buf := fmt.Appendf(nil, "item-%d", i+250)
		sketch2.Insert(buf)
	}

	fmt.Printf("Sketch 1 estimate: %d\n", sketch1.Estimate())
	fmt.Printf("Sketch 2 estimate: %d\n", sketch2.Estimate())

	if err := sketch1.Merge(sketch2); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Merged estimate: %d\n", sketch1.Estimate())

	// Output:
	// Sketch 1 estimate: 500
	// Sketch 2 estimate: 500
	// Merged estimate: 750
}

// Example_precision shows different precision levels and their trade-offs.
func Example_precision() {
	sketch14 := New14() // 16 KB
	sketch16 := New16() // 64 KB

	for i := range 100 {
		buf := fmt.Appendf(nil, "element-%d", i)
		sketch14.Insert(buf)
		sketch16.Insert(buf)
	}

	fmt.Printf("Precision 14: %d unique elements\n", sketch14.Estimate())
	fmt.Printf("Precision 16: %d unique elements\n", sketch16.Estimate())

	// Output:
	// Precision 14: 100 unique elements
	// Precision 16: 100 unique elements
}

// Example_insertHash shows using InsertHash when you already have hash values.
func Example_insertHash() {
	sketch := New()

	sketch.InsertHash(0x1234567890abcdef)
	sketch.InsertHash(0xfedcba0987654321)
	sketch.InsertHash(0x1111111111111111)

	fmt.Printf("Estimated unique elements: %d\n", sketch.Estimate())

	// Output:
	// Estimated unique elements: 3
}
