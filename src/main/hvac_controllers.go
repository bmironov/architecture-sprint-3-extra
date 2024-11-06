package main

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"
	"warm_home/hvac"

	"net/http"
)

func wrapHvacResponse(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf("{\"error\": %s}", strconv.Quote(err.Error()))))
}

func createHvac(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var data hvac.HVAC
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("HVAC.createHvac: can't parse request: %v", err)
		wrapHvacResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	hvac, err := hvac.CreateHvac(db, &data)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(hvac)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func updateHvac(w http.ResponseWriter, req *http.Request) {
	var data hvac.HVAC

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("HVAC.updateHvac: non-integer ID: %s", str)
		wrapHvacResponse(w, err, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("HVAC.createHvac: can't parse request: %v", err)
		wrapHvacResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	data.Id = id
	new_data, err := hvac.UpdateHvac(db, &data)
	if err != nil {
		if new_data != nil && new_data.Id == hvac.NOT_FOUND_ID {
			wrapHvacResponse(w, err, http.StatusNotFound)
		} else {
			wrapHvacResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	json, err := json.Marshal(new_data)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func deleteHvac(w http.ResponseWriter, req *http.Request) {
	fmt.Println("/hvac endpoint called")

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("HVAC.deleteHvac: non-integer ID: %s", str)
		wrapHvacResponse(w, err, http.StatusBadRequest)
		return
	}

	tmp, err := hvac.DeleteHvac(db, id)
	if err != nil {
		if tmp < 0 {
			wrapHvacResponse(w, err, http.StatusNotFound)
		} else {
			wrapHvacResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func findHvac(w http.ResponseWriter, req *http.Request) {
	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("HVAC.loadHvac: non-integer ID: %s", str)
		wrapHvacResponse(w, err, http.StatusBadRequest)
		return
	}

	data, err := hvac.FindHvac(db, id)
	if err != nil {
		if data != nil && data.Id == hvac.NOT_FOUND_ID {
			wrapHvacResponse(w, err, http.StatusNotFound)
		} else {
			wrapHvacResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	json, err := json.Marshal(data)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func addHvacTelemetry(w http.ResponseWriter, req *http.Request) {
	var data hvac.HVACtelemetry

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("HVAC.getHvacTelemetry: non-integer ID: %s", str)
		wrapHvacResponse(w, err, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("HVAC.addHvacTelemetry: can't parse request: %v", err)
		wrapHvacResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	data.HvacId = id
	data.CreatedAt = time.Now().UTC()
	tele, err := hvac.CreateHvacTelemetry(db, &data)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(tele)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func getHvacTelemetry(w http.ResponseWriter, req *http.Request) {
	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("HVAC.getHvacTelemetry: non-integer ID: %s", str)
		wrapHvacResponse(w, err, http.StatusBadRequest)
		return
	}

	data, err := hvac.GetHvacTelemetry(db, id)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(data)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func sendHvacCommand(w http.ResponseWriter, req *http.Request) {
	var data hvac.HVACstate

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("HVAC.sendHvacCommand: non-integer ID: %s", str)
		wrapHvacResponse(w, err, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("HVAC.sendHvacCommand: can't parse request: %v", err)
		wrapHvacResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	if !slices.Contains(hvac.ALLOWED_STATES, data.State) {
		err = fmt.Errorf("invalid state: %s", data.State)
		log.Print(err.Error())
		wrapHvacResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	system, err := hvac.FindHvac(db, id)
	if err != nil {
		if system != nil && system.Id == hvac.NOT_FOUND_ID {
			wrapHvacResponse(w, err, http.StatusNotFound)
		} else {
			wrapHvacResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	json, err := json.Marshal(data)
	if err != nil {
		wrapHvacResponse(w, err, http.StatusInternalServerError)
		return
	}
	fmt.Printf("Sending command to HVAC system #%d\n", id)
	fmt.Printf("Message: %s\n", json)

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
