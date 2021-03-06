package timeflake

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/gioni06/go-timeflake/internal/alphabets"
	"github.com/gioni06/go-timeflake/internal/customerr"
	"github.com/gioni06/go-timeflake/internal/utils"
)

const (
	maxTimestamp = "281474976710655"
	maxRandom    = "1208925819614629174706175"
	maxTimeflake = "340282366920938463463374607431768211455"
)

type Timeflake struct {
	Base62 string
	Hex    string
	Bytes  []byte
	Int    big.Int
	UUID   string
	rand   big.Int
}

func (f *Timeflake) Log() {
	fmt.Printf("ts=%d\trand=%s\tint=%s\thex=%s\tbase62=%s\t\n",
		f.Timestamp(),
		f.BigRand().String(),
		f.Int.String(),
		f.Hex,
		f.Base62,
	)
}

// calculate and return the internal timestamp from big.Int
func (f *Timeflake) Timestamp() int64 {
	t := new(big.Int)
	t.Rsh(&f.Int, 80)
	return t.Int64() / 1000
}

// return the random part of the Timeflake as a string
func (f *Timeflake) Rand() string {
	return f.rand.String()
}

// return the random part of the Timeflake as big.Int
func (f *Timeflake) BigRand() *big.Int {
	return &f.rand
}

func MaxRandom() *big.Int {
	bufMaxRandom := new(bytes.Buffer)
	var MaxRandom = big.NewInt(0)
	MaxRandom.SetString(maxRandom, 62)
	errMaxRandom := binary.Write(bufMaxRandom, binary.LittleEndian, MaxRandom.Bytes())

	if errMaxRandom != nil {
		fmt.Println("binary.Write failed:", errMaxRandom)
	}
	return MaxRandom
}

func MaxTimestamp() *big.Int {
	bufMaxTimestamp := new(bytes.Buffer)
	var MaxTimestamp = big.NewInt(0)
	MaxTimestamp.SetString(maxTimestamp, 10)
	errMaxTimestamp := binary.Write(bufMaxTimestamp, binary.LittleEndian, MaxTimestamp.Bytes())

	if errMaxTimestamp != nil {
		fmt.Println("binary.Write failed:", errMaxTimestamp)
	}
	return MaxTimestamp
}

func MaxTimeflake() *big.Int {
	bufMaxTimeflake := new(bytes.Buffer)
	var MaxTimeflake = big.NewInt(0)
	MaxTimeflake.SetString(maxTimeflake, 10)
	errMaxTimeflake := binary.Write(bufMaxTimeflake, binary.LittleEndian, MaxTimeflake.Bytes())

	if errMaxTimeflake != nil {
		fmt.Println("binary.Write failed:", errMaxTimeflake)
	}

	return MaxTimeflake
}

func Random() (*Timeflake, error) {
	now := time.Now()
	timestamp := now.Unix()

	bigTimestamp := big.NewInt(timestamp * 1000)

	//Generate cryptographically strong pseudo-random between 0 - max
	p := make([]byte, 10) // 80bits
	if _, err := rand.Read(p); err != nil {
		return nil, err
	}

	randomPart := new(big.Int)
	randomPart.SetBytes(p)

	timestampPart := new(big.Int)
	timestampPart.Lsh(bigTimestamp, 80)

	//Mix with random number
	randomAndTimestampCombined := timestampPart.Or(timestampPart, randomPart)

	v62 := big.NewInt(0)
	v62.SetBytes(randomAndTimestampCombined.Bytes())

	vHex := big.NewInt(0)
	vHex.SetBytes(randomAndTimestampCombined.Bytes())

	b62, b62Err := utils.BigIntToASCII(v62, alphabets.BASE62, 22)
	hex, hexErr := utils.BigIntToASCII(vHex, alphabets.HEX, 32)

	if b62Err != nil {
		return nil, b62Err
	}
	if hexErr != nil {
		return nil, hexErr
	}

	u, UUIDErr := uuid.FromBytes(randomAndTimestampCombined.Bytes())

	if UUIDErr != nil {
		return nil, UUIDErr
	}

	f := Timeflake{
		Base62: b62,
		Hex:    hex,
		Bytes:  randomAndTimestampCombined.Bytes(),
		UUID:   u.String(),
		Int:    *randomAndTimestampCombined,
		rand:   *randomPart,
	}
	return &f, nil
}

func FromBytes(fromBytes []byte) (*Timeflake, error) {
	const op = "timeflake:FromBytes"
	if len(fromBytes) != 16 {
		return nil, &customerr.OutOfBoundsError{
			Err: errors.New("fromBytes must be 16 Bytes"),
			Op:  op,
		}
	}

	bigTimestamp := new(big.Int)
	bigTimestamp.SetBytes(fromBytes[0:6])

	randomBytes := fromBytes[6:16]

	randomPart := new(big.Int)
	randomPart.SetBytes(randomBytes)

	timestampPart := big.NewInt(0)
	timestampPart.SetBytes(bigTimestamp.Bytes())
	timestampPart.Lsh(bigTimestamp, 80)

	//Mix with random number
	randomAndTimestampCombined := timestampPart.Or(timestampPart, randomPart)

	v62 := big.NewInt(0)
	v62.SetBytes(randomAndTimestampCombined.Bytes())

	vHex := big.NewInt(0)
	vHex.SetBytes(randomAndTimestampCombined.Bytes())

	b62, b62Err := utils.BigIntToASCII(v62, alphabets.BASE62, 22)
	hex, hexErr := utils.BigIntToASCII(vHex, alphabets.HEX, 32)

	if b62Err != nil {
		return nil, &customerr.ConversionError{
			Err: errors.New("conversion to BASE62 failed"),
			Op:  op,
		}
	}
	if hexErr != nil {
		return nil, &customerr.ConversionError{
			Err: errors.New("conversion to HEX failed"),
			Op:  op,
		}
	}

	u, UUIDErr := uuid.FromBytes(randomAndTimestampCombined.Bytes())

	if UUIDErr != nil {
		return nil, &customerr.UUIDError{
			Err: errors.New("UUID generation failed"),
			Op:  op,
		}
	}

	f := Timeflake{
		Base62: b62,
		Hex:    hex,
		Bytes:  fromBytes,
		UUID:   u.String(),
		Int:    *randomAndTimestampCombined,
		rand:   *randomPart,
	}

	return &f, nil
}

func FromHex(hexValue string) (*Timeflake, error) {
	bigInt := utils.ASCIIToBigInt(hexValue, alphabets.HEX)
	return FromBytes(bigInt.Bytes())
}

func FromBase62(b62 string) (*Timeflake, error) {
	bigInt := utils.ASCIIToBigInt(b62, alphabets.BASE62)
	return FromBytes(bigInt.Bytes())
}

type Values interface {
	Timestamp() int64
	Random() *big.Int
}

// Struct is not exported
type valuesParam struct {
	ts int64
	r  *big.Int
}

func (v *valuesParam) Timestamp() int64 {
	return v.ts
}

func (v *valuesParam) Random() *big.Int {
	return v.r
}

func NewValues(timestamp int64, random *big.Int) (Values, error) {
	if random == nil {
		//Generate cryptographically strong pseudo-random between 0 - max
		p := make([]byte, 10)
		if _, err := rand.Read(p); err != nil {
			return nil, err
		}

		random = new(big.Int)
		random.SetBytes(p)
	}
	return &valuesParam{timestamp, random}, nil // enforce the default value here
}

func FromValues(v Values) (*Timeflake, error) {

	timestamp := v.Timestamp()

	bigTimestamp := big.NewInt(timestamp * 1000)

	timestampPart := new(big.Int)
	timestampPart.Lsh(bigTimestamp, 80)

	//Mix with random number
	randomAndTimestampCombined := timestampPart.Or(timestampPart, v.Random())
	return FromBytes(randomAndTimestampCombined.Bytes())
}
