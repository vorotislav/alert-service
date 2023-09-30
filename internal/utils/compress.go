package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
)

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	gw, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
	if err != nil {
		return nil, fmt.Errorf("failed init compress writer: %w", err)
	}

	_, err = gw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %w", err)
	}

	err = gw.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %w", err)
	}

	return b.Bytes(), nil
}
