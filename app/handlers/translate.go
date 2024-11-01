package handlers

import (
	"encoding/json"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/hlog"
	"github.com/translator/app/models"
)

var (
	translatorErrorAttempts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "translator_error_attempts",
			Help: "Counter of errors of a translator attempts",
		},
	)

	translatorSuccessAttempts = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "translator_success_attempts",
			Help: "Counter of success attempts of a translation",
		},
	)
)

func init() {
	prometheus.MustRegister(translatorErrorAttempts)
	prometheus.MustRegister(translatorSuccessAttempts)
}

// Healthcheck endpoint /Healthcheck
func Healthcheck(w http.ResponseWriter, r *http.Request) {
	mR := models.MyResponse{}
	mR.Msg = "Requested endpoint /healthcheck of translate API service"
	hlog.FromRequest(r).Info().Msg(mR.Msg)
	models.GenerateResponse(w, mR, http.StatusOK)
}

// Translate Translate
func Translate(w http.ResponseWriter, r *http.Request) {
	mR := models.MyResponse{Code: 1, Msg: "unexpected error"}

	input := models.InputTranslate{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&input)
	defer r.Body.Close()

	if err != nil {
		translatorErrorAttempts.Inc()
		mR.Msg = "error decoding input"
		hlog.FromRequest(r).Error().Err(err).Msg(mR.Msg)
		models.GenerateResponse(w, mR, http.StatusBadRequest)
		return
	}

	if !models.IsValidInputTranslate(input) {
		translatorErrorAttempts.Inc()
		mR.Msg = "invalid params"
		hlog.FromRequest(r).Error().Err(err).Msg(mR.Msg)
		models.GenerateResponse(w, mR, http.StatusBadRequest)
		return
	}

	translatedText, err := models.Translate(input.Text, input.From, input.To)
	if err != nil {
		translatorErrorAttempts.Inc()
		mR.Msg = "error in translation"
		hlog.FromRequest(r).Error().Err(err).Msg(mR.Msg)
		models.GenerateResponse(w, mR, http.StatusInternalServerError)
		return
	}

	translatorSuccessAttempts.Inc()

	output := make(map[string]interface{})
	output["text_translated"] = translatedText

	mR.Code = 0
	mR.Msg = "ok"
	mR.Data = output
	models.GenerateResponse(w, mR, http.StatusOK)
}
