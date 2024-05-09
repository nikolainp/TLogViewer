package filewalker

import (
	"bufio"
	"bytes"
	"io"
	"sync"
)

type streamBuffer struct {
	buf *[]byte
	len int
}

type lineChecker struct {
	poolBuf sync.Pool
	chBuf   chan streamBuffer

	isCancel CancelFunc

	bufSize          int
	prefixFirstLine  string
	prefixSecondLine []byte
}

func (obj *lineChecker) init(isCancelFunc CancelFunc) {
	obj.bufSize = 1024 * 1024 * 10

	obj.isCancel = isCancelFunc

	obj.poolBuf = sync.Pool{New: func() interface{} {
		lines := make([]byte, obj.bufSize)
		return &lines
	}}
}

func (obj *lineChecker) processStream(sName string, sIn io.Reader, fOut EventWalkFunc) {
	var wg sync.WaitGroup

	goFunc := func(work func()) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			work()
		}()
	}

	obj.prefixFirstLine = sName
	obj.prefixSecondLine = []byte("<line>")

	obj.chBuf = make(chan streamBuffer, 1)

	goFunc(func() { obj.doRead(sIn) })
	goFunc(func() { obj.doWrite(fOut) })

	wg.Wait()
}

func (obj *lineChecker) doRead(sIn io.Reader) {

	reader := bufio.NewReaderSize(sIn, obj.bufSize)

	readBuffer := func(buf []byte) int {
		n, err := reader.Read(buf)
		if n == 0 && err == io.EOF {
			return 0
		}

		if obj.isCancel() {
			return 0
		}

		return n
	}

	for {
		buf := obj.poolBuf.Get().(*[]byte)
		if n := readBuffer(*buf); n == 0 {
			break
		} else {
			obj.chBuf <- streamBuffer{buf, n}
		}
	}

	close(obj.chBuf)
}

func (obj *lineChecker) doWrite(fOut EventWalkFunc) {

	var bufOut bytes.Buffer
	lastLine := make([]byte, obj.bufSize*2)
	isExistsLastLine := false

	writeBuffer := func(buf []byte, n int) {
		isLastStringFull := bytes.Equal(buf[n-1:n], []byte("\n"))

		bufSlice := bytes.Split(buf[:n], []byte("\n"))
		if 3 <= len(bufSlice[0]) && bytes.Equal(bufSlice[0][:3], []byte("\ufeff")) {
			bufSlice[0] = bufSlice[0][3:]
		}

		for i := range bufSlice {
			if i == 0 && isExistsLastLine {
				lastLine = append(lastLine, bufSlice[i]...)
				if len(bufSlice) > 1 {
					obj.lineProcessor(&bufOut, lastLine, fOut)
					isExistsLastLine = false
				}
				continue
			}
			if i == len(bufSlice)-1 {
				if !isLastStringFull {
					lastLine = lastLine[0:len(bufSlice[i])]
					nc := copy(lastLine, bufSlice[i])
					if nc != len(bufSlice[i]) {
						panic(0)
					}
					isExistsLastLine = true
				}
				continue
			}

			if obj.isCancel() {
				return
			}

			obj.lineProcessor(&bufOut, bufSlice[i], fOut)
		}
	}

	for {
		if buffer, ok := <-obj.chBuf; ok {
			writeBuffer(*(buffer.buf), buffer.len)

			obj.poolBuf.Put(buffer.buf)
		} else {
			if isExistsLastLine {
				obj.lineProcessor(&bufOut, lastLine, fOut)
			}
			obj.writeEvent(&bufOut, fOut)
			break
		}

		if obj.isCancel() {
			break
		}
	}
}

func (obj *lineChecker) lineProcessor(buf *bytes.Buffer, data []byte, writer EventWalkFunc) {

	if obj.isFirstLine(data) {
		obj.writeEvent(buf, writer)
	} else {
		buf.Write(obj.prefixSecondLine)
	}
	buf.Write(data)
}

func (obj *lineChecker) writeEvent(buf *bytes.Buffer, writer EventWalkFunc) {
	if 0 < buf.Len() {
		writer(obj.prefixFirstLine, buf.String())
	}
	buf.Reset()
}

func (obj *lineChecker) isFirstLine(data []byte) bool {

	isNumber := func(data byte) bool {
		if data == '0' || data == '1' || data == '2' || data == '3' ||
			data == '4' || data == '5' || data == '6' || data == '7' ||
			data == '8' || data == '9' {
			return true
		}

		return false
	}

	// `^\d\d\:\d\d\.\d{6}\-\d+\,\w+\,`
	if len(data) < 14 {
		return false
	}
	if !isNumber(data[0]) || !isNumber(data[1]) || !isNumber(data[3]) || !isNumber(data[4]) {
		return false
	}
	if data[2] != ':' {
		return false
	}
	if data[5] != '.' {
		return false
	}
	if !isNumber(data[6]) || !isNumber(data[7]) || !isNumber(data[8]) ||
		!isNumber(data[9]) || !isNumber(data[10]) || !isNumber(data[11]) {
		return false
	}
	if data[12] != '-' {
		return false
	}
	if !isNumber(data[13]) {
		return false
	}

	return true
}
