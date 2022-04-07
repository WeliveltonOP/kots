package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/replicatedhq/kots/pkg/helm"
	"github.com/replicatedhq/kots/pkg/logger"
)

type HelmLoginRequest struct {
	Hostname string `json:"hostname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) HelmLogin(w http.ResponseWriter, r *http.Request) {
	helmLoginRequest := HelmLoginRequest{}
	if err := json.NewDecoder(r.Body).Decode(&helmLoginRequest); err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ok, err := helm.CheckHelmRegistryCredentials(helmLoginRequest.Hostname, helmLoginRequest.Username, helmLoginRequest.Password)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !ok {
		data := map[string]interface{}{
			// TODO we're still using Go v1.12
			// Go v1.13 has Unwrap() and this would reduce to:
			// "error": errors.Unwrap(err),
			"error": "invalid credentials",
		}
		JSON(w, http.StatusBadRequest, data)
		return
	}

	JSON(w, http.StatusNoContent, "")
}
