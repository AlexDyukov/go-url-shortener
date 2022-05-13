package webhandler

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

type compressWriter struct {
	HTTPResponseWriter http.ResponseWriter
	CompressedWriter   io.Writer
}

func (w compressWriter) Write(b []byte) (int, error) {
	return w.CompressedWriter.Write(b)
}

func (w compressWriter) WriteHeader(statusCode int) {
	w.HTTPResponseWriter.WriteHeader(statusCode)
}

func (w compressWriter) Header() http.Header {
	return w.HTTPResponseWriter.Header()
}

func newCompressHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//response
		acceptEncoding := strings.ToLower(r.Header.Get("Accept-Encoding"))
		writer := io.Writer(w)
		switch {
		case strings.Contains(acceptEncoding, "gzip"):
			gzipWriter, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil { // failure to initialize gzip -> ignore
				break
			}
			defer gzipWriter.Close()
			w.Header().Set("Content-Encoding", "gzip")
			writer = gzipWriter
		case strings.Contains(acceptEncoding, "deflate"):
			zlibWriter, err := zlib.NewWriterLevel(w, zlib.BestSpeed)
			if err != nil { // failure to initialize zlib -> ignore
				break
			}
			defer zlibWriter.Close()
			w.Header().Set("Content-Encoding", "deflate")
			writer = zlibWriter
		}

		//request
		contentEncoding := strings.ToLower(r.Header.Get("Content-Encoding"))
		switch {
		case contentEncoding == "gzip":
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil { // failure to initialize gzip
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			defer gzipReader.Close()
			r.Body = gzipReader
		case contentEncoding == "deflate":
			zlibReader, err := zlib.NewReader(r.Body)
			if err != nil { // failure to initialize zlib
				http.Error(w, err.Error(), http.StatusBadRequest)
			}
			defer zlibReader.Close()
			r.Body = zlibReader
		}

		next.ServeHTTP(compressWriter{HTTPResponseWriter: w, CompressedWriter: writer}, r)
	})
}
