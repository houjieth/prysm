package ssz

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// TODOs for this PR:
// - Review all error handling
// - Use customized error types to avoid error string everywhere!
// - Add check for input sizes!

func Encode(w io.Writer, val interface{}) error {
	eb := &encbuf{}
	if err := eb.encode(val); err != nil {
		return err
	}
	return eb.toWriter(w)
}

type encbuf struct {
	str []byte
}

func (w *encbuf) encode(val interface{}) error {
	rval := reflect.ValueOf(val)
	encDec, err := getEncoderDecoderForType(rval.Type())
	if err != nil {
		return err
	}
	return encDec.encoder(rval, w)
}

func (w *encbuf) toWriter(out io.Writer) error {
	if _, err := out.Write(w.str); err != nil {
		return err
	}
	return nil
}

func makeEncoder(typ reflect.Type) (encoder, error) {
	kind := typ.Kind()
	switch {
	case kind == reflect.Uint8:
		return encodeUint8, nil
	case kind == reflect.Uint16:
		return encodeUint16, nil
	case kind == reflect.Slice && typ.Elem().Kind() == reflect.Uint8:
		return encodeBytes, nil
	case kind == reflect.Slice:
		return makeSliceEncoder(typ)
	default:
		return nil, fmt.Errorf("ssz: type %v is not serializable", typ)
	}
}

func encodeUint8(val reflect.Value, w *encbuf) error {
	v := val.Uint()
	w.str = append(w.str, uint8(v))
	return nil
}

func encodeUint16(val reflect.Value, w *encbuf) error {
	v := val.Uint()
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	w.str = append(w.str, b...)
	return nil
}

func encodeBytes(val reflect.Value, w *encbuf) error {
	b := val.Bytes()
	sizeEnc := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeEnc, uint32(len(b)))
	w.str = append(w.str, sizeEnc...)
	w.str = append(w.str, val.Bytes()...)
	return nil
}

func makeSliceEncoder(typ reflect.Type) (encoder, error) {
	elemEncoderDecoder, err := getEncoderDecoderForType(typ.Elem())
	if err != nil {
		return nil, fmt.Errorf("failed to get encoder/decoder: %v", err)
	}
	encoder := func(val reflect.Value, w *encbuf) error {
		// TODO: totalSize should've been already known in the parsing pass. You need to add that feature to your parsing code
		origBufSize := len(w.str)
		totalSizeEnc := make([]byte, 4)
		w.str = append(w.str, totalSizeEnc...)
		for i := 0; i < val.Len(); i++ {
			if err := elemEncoderDecoder.encoder(val.Index(i), w); err != nil {
				return err
			}
		}
		totalSize := len(w.str) - 4 - origBufSize
		binary.BigEndian.PutUint32(totalSizeEnc, uint32(totalSize))
		copy(w.str[origBufSize: origBufSize+4], totalSizeEnc)
		return nil
	}
	return encoder, nil
}
