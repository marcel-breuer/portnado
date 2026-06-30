package protocol

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var ErrFrameTooLarge = errors.New("protocol frame too large")

type Codec struct {
	reader *bufio.Reader
	writer io.Writer
}

func NewCodec(readWriter io.ReadWriter) *Codec {
	return &Codec{
		reader: bufio.NewReaderSize(readWriter, 4096),
		writer: readWriter,
	}
}

func (c *Codec) ReadRequest() (Request, error) {
	var request Request
	if err := c.readJSON(&request); err != nil {
		return request, err
	}
	return request, nil
}

func (c *Codec) WriteRequest(request Request) error {
	return writeJSON(c.writer, request)
}

func (c *Codec) ReadResponse() (Response, error) {
	var response Response
	if err := c.readJSON(&response); err != nil {
		return response, err
	}
	return response, nil
}

func (c *Codec) WriteResponse(response Response) error {
	return writeJSON(c.writer, response)
}

func (c *Codec) readJSON(target any) error {
	line := make([]byte, 0, 4096)
	for {
		chunk, err := c.reader.ReadSlice('\n')
		line = append(line, chunk...)
		if len(line) > MaxFrameSize {
			return ErrFrameTooLarge
		}
		if err == nil {
			break
		}
		if !errors.Is(err, bufio.ErrBufferFull) {
			return err
		}
	}
	if err := json.Unmarshal(line, target); err != nil {
		return fmt.Errorf("decode protocol frame: %w", err)
	}
	return nil
}

func writeJSON(writer io.Writer, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode protocol frame: %w", err)
	}
	if len(data)+1 > MaxFrameSize {
		return ErrFrameTooLarge
	}
	data = append(data, '\n')
	_, err = writer.Write(data)
	return err
}
