package logs

import (
	"net/http"
)

// writer is an [http.ResponseWriter] wrapper that records the status code
// passed to [http.ResponseWriter.WriteHeader].
type writer struct {
	http.ResponseWriter
	status int
}

func (w *writer) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *writer) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// Status returns the recorded status code. If neither WriteHeader nor Write
// was called (e.g. on a hijacked connection), it returns 0.
func (w *writer) Status() int {
	return w.status
}

// wrapResponseWriter wraps w so that the middleware can observe status and
// response headers. The returned [http.ResponseWriter] preserves whichever of
// [http.Hijacker], [http.Flusher], and [http.Pusher] the underlying writer
// implements, so that downstream code can still type-assert to those
// interfaces and have the assertions reflect real capability.
func wrapResponseWriter(w http.ResponseWriter) (*writer, http.ResponseWriter) {
	ww := &writer{ResponseWriter: w}

	h, isHijacker := w.(http.Hijacker)
	f, isFlusher := w.(http.Flusher)
	p, isPusher := w.(http.Pusher)

	switch {
	case isHijacker && isFlusher && isPusher:
		return ww, struct {
			*writer
			http.Hijacker
			http.Flusher
			http.Pusher
		}{ww, h, f, p}
	case isHijacker && isFlusher:
		return ww, struct {
			*writer
			http.Hijacker
			http.Flusher
		}{ww, h, f}
	case isHijacker && isPusher:
		return ww, struct {
			*writer
			http.Hijacker
			http.Pusher
		}{ww, h, p}
	case isFlusher && isPusher:
		return ww, struct {
			*writer
			http.Flusher
			http.Pusher
		}{ww, f, p}
	case isHijacker:
		return ww, struct {
			*writer
			http.Hijacker
		}{ww, h}
	case isFlusher:
		return ww, struct {
			*writer
			http.Flusher
		}{ww, f}
	case isPusher:
		return ww, struct {
			*writer
			http.Pusher
		}{ww, p}
	}

	return ww, ww
}
