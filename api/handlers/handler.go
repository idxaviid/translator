package handlers

import (
	"encoding/json"

	"net/http"

	gt "github.com/bas24/googletranslatefree"
	"github.com/rs/zerolog/hlog"
	"github.com/translator/api/models"
)

// Index endpoint /index
func Index(w http.ResponseWriter, r *http.Request) {
	mR := models.MyResponse{}
	mR.Msg = "Requested endpoint /index of translate API service"
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
		mR.Msg = "invalid params"
		hlog.FromRequest(r).Error().Err(err).Msg(mR.Msg)
		models.GenerateResponse(w, mR, http.StatusBadRequest)
		return
	}

	translatedText, err := gt.Translate(input.Text, input.From, input.To)
	if err != nil {
		mR.Msg = "unexpected error in translation"
		hlog.FromRequest(r).Error().Err(err).Msg(mR.Msg)
		models.GenerateResponse(w, mR, http.StatusInternalServerError)
		return
	}

	output := make(map[string]interface{})
	output["text_translated"] = translatedText

	mR.Code = 0
	mR.Msg = ""
	mR.Data = output
	models.GenerateResponse(w, mR, http.StatusOK)
}
