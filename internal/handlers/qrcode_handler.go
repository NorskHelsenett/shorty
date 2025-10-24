package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/NorskHelsenett/shorty/internal/media"
	redisdb "github.com/NorskHelsenett/shorty/internal/redis"

	"github.com/NorskHelsenett/ror/pkg/rlog"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

type QrWriteCloser struct {
	io.Writer
}

func (mwc *QrWriteCloser) Close() error {
	return nil
}

// getBaseURL returns the base URL from environment variable or default
func getBaseURL() string {
	baseURL := viper.GetString("BASE_URL")
	if baseURL == "" {
		baseURL = "https://k.nhn.no"
	}
	return baseURL
}

// @Summary	Get qr-code by id
// @Schemes
// @Description	gets qrcode by id
// @Tags			v1
// @Accept			application/json
// @Produce		application/json
// @Produce		image/png
// @Param			id	path		string	true	"Id"
// @Success		200	{file }		imge/png
// @Failure		403	{string}	Forbidden
// @Failure		401	{string}	Unauthorized
// @Failure		500	{string}	Failure	message
// @Router			/v1/qr/{id} [get]
// @Security		AccessToken
func GenerateQRCode(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		rlog.Debug("GenerateQRCode", rlog.Any("id", id))

		path, err := redisdb.GetURL(rdb, id)
		shorturl := path
		if len(path) != 0 {
			shorturl = fmt.Sprintf("%s/%s", getBaseURL(), id)
		}
		if err != nil {
			rlog.Info("GenerateQRCode - Error in GetURL", rlog.Any("client", r.Host), rlog.Any("path", id), rlog.Any("to", path))
			http.Error(w, "Could not fetch URL from database", http.StatusInternalServerError)
			return
		}
		rlog.Info("GenerateQRCode", rlog.Any("id", id), rlog.Any("path:", path))

		handleQrImageCreation(shorturl, w)
	}
}

// @Summary	Get qrcode by query
// @Schemes
// @Description	get qrcode by query
// @Tags			qr
// @Accept			application/json
// @Produce		application/json
// @Produce		image/png
// @Param			q	query		string	true	"Query"
// @Success		200	{file }		imge/png
// @Failure		403	{string}	Forbidden
// @Failure		401	{string}	Unauthorized
// @Failure		500	{string}	Failure	message
// @Router			/qr/ [get]
func GenerateQRCodeFromUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("q")

		if url == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte(`{"error": "Missing query parameter 'q'"}`)); err != nil {
				rlog.Error("Failed to write error response", err)
				return
			}
			return
		}

		handleQrImageCreation(url, w)
	}
}

// handleQrImageCreation generates a QR code image from the provided URL input
// and writes it to the HTTP response writer with error handling
func handleQrImageCreation(input string, w http.ResponseWriter) {
	rlog.Infof("handleQrImageCreation - input: %s", input)

	// Validate input is a valid URL
	u, err := url.ParseRequestURI(input)
	if err != nil {
		rlog.Error("Failed to parse url", err, rlog.Any("url", input))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(fmt.Sprintf(`{"error": "Invalid URL format: %s"}`, err.Error()))); err != nil {
			rlog.Error("Failed to write error response", err)
		}
		return
	}

	// Set error correction level
	correctionLevel := qrcode.ErrorCorrectionHighest

	// Generate QR code with explicit error handling
	qrc, err := qrcode.NewWith(input, qrcode.WithErrorCorrectionLevel(correctionLevel))
	if err != nil {
		rlog.Error("Failed to generate QR code", err, rlog.Any("url", input))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error": "Failed to generate QR code"}`)); err != nil {
			rlog.Error("Failed to write error response", err)
		}
		return
	}

	// Create buffer for the image
	buf := bytes.NewBuffer(nil)
	wr := QrWriteCloser{Writer: buf}

	opts := []standard.ImageOption{
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
	}

	// Add the NHN logo for NHN domains
	if strings.HasSuffix(u.Host, "nhn.no") {
		// Verify logo file exists before trying to use it
		if _, err := os.Stat(media.ImageFile); err == nil {
			opts = append(opts, standard.WithLogoSizeMultiplier(2))
			opts = append(opts, standard.WithLogoImageFilePNG(media.ImageFile))
		} else {
			rlog.Warn("NHN logo file not found, generating QR code without logo", rlog.Any("logoPath", media.ImageFile))
		}
	}

	// Create a writer with the appropriate options
	writer := standard.NewWithWriter(&wr, opts...)
	defer writer.Close()

	// Save the QR code to the writer with explicit error handling
	if err = qrc.Save(writer); err != nil {
		rlog.Error("Failed to save QR code image", err, rlog.Any("input", input))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(fmt.Sprintf(`{"error": "Failed to generate QR code image: %s"}`, err.Error()))); err != nil {
			rlog.Error("Failed to write error response", err)
			return
		}
		return
	}

	// Set appropriate content type and write the image data
	w.Header().Set("Content-Type", "image/png")
	if _, err := w.Write(buf.Bytes()); err != nil {
		rlog.Error("Failed to write QR code image to response", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
