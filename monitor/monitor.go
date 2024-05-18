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
	filesFinished int64
	sizeTotal     int64
	sizeFinished  int64
	messageBuffer []string

	startTime time.Time
	ticker    *time.Ticker
	done      chan bool

	mu         sync.Mutex
	wg         sync.WaitGroup
	cancelChan chan bool
}

func New(isCancelChan chan bool) *monitor {
	obj := new(monitor)
	obj.startTime = time.Now()
	obj.ticker = time.NewTicker(500 * time.Millisecond)
	obj.done = make(chan bool)

	obj.cancelChan = isCancelChan

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

	obj.filesFinished += count
	obj.sizeFinished += size
}

func (obj *monitor) Start() {
	obj.print()
}

func (obj *monitor) Stop() {
	defer obj.ticker.Stop()

	obj.done <- true
	obj.wg.Wait()
}

func (obj *monitor) IsCancel() bool {
	select {
	case _, ok := <-obj.cancelChan:
		return !ok
	default:
		return false
	}
}

func (obj *monitor) Cancel() chan bool {
	return obj.cancelChan
}

///////////////////////////////////////////////////////////////////////////////

func (obj *monitor) print() {
	var prevFinishedSize int64
	var prevDuration time.Duration

	doPrint := func() {
		defer obj.mu.Unlock()
		obj.mu.Lock()

		var speed int64
		var totalSpeed int64

		totalDuration := time.Since(obj.startTime)
		totalSpeed = 1000 * obj.sizeFinished / totalDuration.Milliseconds()

		deltaDuration := totalDuration - prevDuration
		if deltaDuration.Seconds() > 0 {
			speed = 1000 * (obj.sizeFinished - prevFinishedSize) / deltaDuration.Milliseconds()
			if deltaDuration.Seconds() < 1 {
				speed = 1000 * speed / deltaDuration.Milliseconds()
			}
		}
		if deltaDuration.Seconds() > 1 {
			prevDuration = totalDuration
			prevFinishedSize = obj.sizeFinished
		}

		for i := range obj.messageBuffer {
			fmt.Fprint(os.Stderr, obj.messageBuffer[i])
		}
		obj.messageBuffer = obj.messageBuffer[:0]

		fmt.Fprintf(os.Stderr,
			"files: %d/%d size: %s/%s time: %s [speed %s/s/%s/s ]                           \r",
			obj.filesFinished, obj.filesTotal,
			byteCount(obj.sizeFinished), byteCount(obj.sizeTotal), totalDuration,
			byteCount(speed), byteCount(totalSpeed))

		// fmt.Fprintf(os.Stderr,
		// 	"%s %s [%f]: in work: %d finished: %d\r",
		// 	byteCount(obj.sizeTotal), totalDuration,
		// 	speed,
		// 	obj.streamsInWork, obj.streamsFinished)
	}

	obj.wg.Add(1)
	go func() {
		defer obj.wg.Done()

		for {
			var done, cancel bool

			select {
			case done = <-obj.done:

			case _, ok := <-obj.cancelChan:
				cancel = !ok

			case <-obj.ticker.C:
				doPrint()
			}

			if done || cancel {
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
		return fmt.Sprintf("%db", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cb",
		float64(b)/float64(div), "kMGTPE"[exp])
}
