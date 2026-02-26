package payload

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

// ---------------------------------------------------------------------------
// Payload layout (136 bits total, well within the 450-bit limit)
//
//  [0:15]   START_FLAG  – 16 bits  : 1111 0000 1111 0000
//  [16:19]  VERSION     –  4 bits  : protocol version (0-15)
//  [20]     IS_AI_FLAG  –  1 bit   : 1 = AI-generated content
//  [21:23]  RESERVED    –  3 bits  : future flags (set to 0)
//  [24:87]  METADATA_ID – 64 bits  : uint64 row ID from PostgreSQL
//  [88:119] CRC32       – 32 bits  : CRC-32/IEEE over bytes [2:11]
//                                    (VERSION+FLAGS+METADATA_ID region)
//  [120:135] END_FLAG   – 16 bits  : 0000 1111 0000 1111
//
// Total: 136 bits
// ---------------------------------------------------------------------------

const (
	PayloadVersion   = 1
	PayloadTotalBits = 136

	startFlagVal uint16 = 0xF0F0 // 1111 0000 1111 0000
	endFlagVal   uint16 = 0x0F0F // 0000 1111 0000 1111
)

// PayloadFields holds the structured data embedded in the watermark.
type PayloadFields struct {
	Version    uint8  // 4 bits (0-15)
	IsAI       bool   // 1 bit
	Reserved   uint8  // 3 bits (must be 0)
	MetadataID uint64 // 64 bits — PostgreSQL row/UUID numeric ID
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func uint16ToBits(v uint16) []int {
	bits := make([]int, 16)
	for i := 15; i >= 0; i-- {
		bits[15-i] = int((v >> i) & 1)
	}
	return bits
}

func uint32ToBits(v uint32) []int {
	bits := make([]int, 32)
	for i := 31; i >= 0; i-- {
		bits[31-i] = int((v >> i) & 1)
	}
	return bits
}

func uint64ToBits(v uint64) []int {
	bits := make([]int, 64)
	for i := 63; i >= 0; i-- {
		bits[63-i] = int((v >> i) & 1)
	}
	return bits
}

func bitsToUint16(bits []int) uint16 {
	var v uint16
	for _, b := range bits {
		v = (v << 1) | uint16(b)
	}
	return v
}

func bitsToUint32(bits []int) uint32 {
	var v uint32
	for _, b := range bits {
		v = (v << 1) | uint32(b)
	}
	return v
}

func bitsToUint64(bits []int) uint64 {
	var v uint64
	for _, b := range bits {
		v = (v << 1) | uint64(b)
	}
	return v
}

// buildProtectedBytes serialises the region covered by the CRC (10 bytes):
//
//	byte 0   : VERSION (high nibble) | FLAGS (low nibble)
//	bytes 1-8: MetadataID big-endian
func buildProtectedBytes(fields PayloadFields) []byte {
	buf := make([]byte, 9)
	flagsByte := (fields.Version & 0x0F) << 4
	if fields.IsAI {
		flagsByte |= 0x08
	}
	flagsByte |= fields.Reserved & 0x07
	buf[0] = flagsByte
	binary.BigEndian.PutUint64(buf[1:], fields.MetadataID)
	return buf
}

func computeCRC(protected []byte) uint32 {
	return crc32.ChecksumIEEE(protected)
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// PayloadGenerate builds the 136-bit watermark payload as a slice of ints (0/1).
//
// Layout:
//
//	START_FLAG(16) | VERSION(4) | IS_AI(1) | RESERVED(3) | METADATA_ID(64) | CRC32(32) | END_FLAG(16)
func PayloadGenerate(fields PayloadFields) ([]int, error) {
	if fields.Version > 15 {
		return nil, errors.New("version must fit in 4 bits (0-15)")
	}
	if fields.Reserved > 7 {
		return nil, errors.New("reserved must fit in 3 bits (0-7)")
	}

	protected := buildProtectedBytes(fields)
	crc := computeCRC(protected)

	payload := make([]int, 0, PayloadTotalBits)

	// START FLAG (16 bits)
	payload = append(payload, uint16ToBits(startFlagVal)...)

	// VERSION (4 bits)
	for i := 3; i >= 0; i-- {
		payload = append(payload, int((fields.Version>>i)&1))
	}

	// IS_AI_FLAG (1 bit)
	if fields.IsAI {
		payload = append(payload, 1)
	} else {
		payload = append(payload, 0)
	}

	// RESERVED (3 bits)
	for i := 2; i >= 0; i-- {
		payload = append(payload, int((fields.Reserved>>i)&1))
	}

	// METADATA_ID (64 bits)
	payload = append(payload, uint64ToBits(fields.MetadataID)...)

	// CRC32 (32 bits)
	payload = append(payload, uint32ToBits(crc)...)

	// END FLAG (16 bits)
	payload = append(payload, uint16ToBits(endFlagVal)...)

	if len(payload) != PayloadTotalBits {
		return nil, errors.New("internal error: unexpected payload length")
	}

	return payload, nil
}

// ---------------------------------------------------------------------------

// ParseResult holds the decoded payload or an error description.
type ParseResult struct {
	Fields PayloadFields
	Valid  bool   // false if CRC mismatch or flags invalid
	Err    string // human-readable reason when Valid == false
}

// parsePayload decodes a single 136-bit payload slice.
func parsePayload(bits []int) ParseResult {
	if len(bits) < PayloadTotalBits {
		fmt.Println("error in PayloadTotalBits , length = ", len(bits), " PAYload total bits", PayloadTotalBits)
		return ParseResult{Err: "wrong payload length"}
	}

	// Validate start flag
	if bitsToUint16(bits[0:16]) != startFlagVal {
		fmt.Println(" error in startFlagVal")
		return ParseResult{Err: "start flag mismatch"}
	}

	// Validate end flag
	if bitsToUint16(bits[120:136]) != endFlagVal {
		fmt.Println("Error in endfalag ")
		return ParseResult{Err: "end flag mismatch"}
	}

	version := uint8(bitsToUint32(bits[16:20]))
	isAI := bits[20] == 1
	reserved := uint8(bitsToUint32(bits[21:24]))
	metadataID := bitsToUint64(bits[24:88])
	embeddedCRC := bitsToUint32(bits[88:120])

	fields := PayloadFields{
		Version:    version,
		IsAI:       isAI,
		Reserved:   reserved,
		MetadataID: metadataID,
	}

	//Verify CRC
	protected := buildProtectedBytes(fields)
	expectedCRC := computeCRC(protected)
	if embeddedCRC != expectedCRC {
		return ParseResult{
			Fields: fields,
			Err:    "CRC mismatch — payload corrupted",
		}
	}

	return ParseResult{Fields: fields, Valid: true}
}

// ---------------------------------------------------------------------------

// PayloadVerify receives multiple copies of a 136-bit payload recovered from
// different image blocks, determines the most frequently occurring valid payload
// (majority vote), and returns the decoded fields.
//
// It returns an error only when no valid, agreeing majority can be found.
func PayloadVerify(payloads [][]int) (PayloadFields, error) {
	if len(payloads) == 0 {
		return PayloadFields{}, errors.New("no payloads provided")
	}

	type candidate struct {
		fields PayloadFields
		count  int
	}

	// Use MetadataID+flags as the grouping key (represented as uint64+byte).
	type key struct {
		metadataID uint64
		flagByte   byte // version|isAI|reserved packed
	}

	tally := make(map[key]*candidate)
	validCount := 0

	for _, bits := range payloads {
		r := parsePayload(bits)
		if !r.Valid {
			continue
		}
		validCount++

		fb := (r.Fields.Version & 0x0F) << 4
		if r.Fields.IsAI {
			fb |= 0x08
		}
		fb |= r.Fields.Reserved & 0x07

		k := key{metadataID: r.Fields.MetadataID, flagByte: fb}
		if _, ok := tally[k]; !ok {
			tally[k] = &candidate{fields: r.Fields}
		}
		tally[k].count++
	}

	if validCount == 0 {
		return PayloadFields{}, errors.New("all payloads failed CRC / flag validation")
	}

	// Pick the candidate with the highest vote count.
	var best *candidate
	for _, c := range tally {
		if best == nil || c.count > best.count {
			best = c
		}
	}
	println("Payload verify success ")
	return best.fields, nil
}
