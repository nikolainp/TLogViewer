package monitor

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type CancelFunc func() bool

type monitor struct {
	filesTotal    int64
	sizeTotal     int64
	sizeFinished  int64
	messageBuffer []string

	startTime time.Time
	ticker    *time.Ticker

	mu       sync.Mutex
	isCancel CancelFunc
}

func GetLogger(isCancelFunc CancelFunc) *monitor {
	obj := new(monitor)
	obj.startTime = time.Now()
	obj.ticker = time.NewTicker(100 * time.Millisecond)

	obj.isCancel = isCancelFunc

	return obj
	//return &monitor{startTime: time.Now()}
}

func (obj *monitor) WriteEvent(frmt string, args ...any) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.messageBuffer = append(obj.messageBuffer, fmt.Sprintf(frmt, args...))
}

func (obj *monitor) NewData(size int64) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.filesTotal++
	obj.sizeTotal += size
}

func (obj *monitor) FinishedData(count, size int64) {
	defer obj.mu.Unlock()
	obj.mu.Lock()

	obj.filesTotal += count
	obj.sizeTotal += size
}

func (obj *monitor) Start() {
	go obj.print()
}

func (obj *monitor) Stop() {
	obj.ticker.Stop()
}

///////////////////////////////////////////////////////////////////////////////

func (obj *monitor) print() {
	var prevFinishedSize int64
	var prevDuration time.Duration

	doPrint := func() {
		defer obj.mu.Unlock()
		obj.mu.Lock()

		var speed float64

		totalDuration := time.Since(obj.startTime)
		deltaDuration := totalDuration - prevDuration
		if deltaDuration.Seconds() > 0 {
			speed = float64(obj.sizeFinished-prevFinishedSize) / deltaDuration.Seconds()
			prevDuration = totalDuration
			prevFinishedSize = obj.sizeFinished
		}

		for i := range obj.messageBuffer {
			fmt.Fprint(os.Stderr, obj.messageBuffer[i])
		}
		obj.messageBuffer = obj.messageBuffer[:0]

		fmt.Fprintf(os.Stderr,
			"%s %s [%f]\r",
			byteCount(obj.sizeTotal), totalDuration,
			speed)

		// fmt.Fprintf(os.Stderr,
		// 	"%s %s [%f]: in work: %d finished: %d\r",
		// 	byteCount(obj.sizeTotal), totalDuration,
		// 	speed,
		// 	obj.streamsInWork, obj.streamsFinished)
	}

	go func() {
		for range obj.ticker.C {
			doPrint()

			if obj.isCancel() {
				break
			}
		}
		doPrint()
		fmt.Fprintf(os.Stderr, "\n")
	}()

	// TODO: + total bytesnv time spend [ speed ] [ in work %d - finished %d ]
}

///////////////////////////////////////////////////////////////////////////////

func byteCount(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
