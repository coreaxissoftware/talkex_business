package observability

import (
	"fmt"
	"log"
	"runtime/debug"
)

// Safely runs fn and recovers any panic, logging the stack trace and
// capturing it to Sentry. Meant for background-worker loops where a
// panic in one tick must NOT bring the whole API process down.
//
// Typical use:
//
//	for range ticker.C {
//	    observability.Safely("messaging worker tick", func() {
//	        ProcessQueue(db, batchSize)
//	    })
//	}
//
// The label is what shows up in Sentry as the event message, so pick
// something a support engineer can grep for.
func Safely(label string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Printf("PANIC in %s: %v\n%s", label, r, stack)

			var err error
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
			CaptureError(err, map[string]interface{}{
				"context": "background_worker",
				"label":   label,
			})
		}
	}()
	fn()
}
