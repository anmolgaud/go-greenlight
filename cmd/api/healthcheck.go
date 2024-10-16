package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status": "available",
		"systemInfo": map[string]string{
			"environment": app.config.env,
			"version": version,
		},
	}
	err := app.writeJson(w, http.StatusOK, data, nil)
	if err != nil {
		app.logger.Print(err);
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}