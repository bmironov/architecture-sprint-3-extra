package main

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"
	"warm_home/light"

	"net/http"
)

func wrapLightResponse(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	w.Write([]byte(fmt.Sprintf("{\"error\": %s}", strconv.Quote(err.Error()))))
}

func createLight(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var data light.LIGHT
	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("LIGHT.createLight: can't parse request: %v", err)
		wrapLightResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	light, err := light.CreateLight(db, &data)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(light)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func updateLight(w http.ResponseWriter, req *http.Request) {
	var data light.LIGHT

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("LIGHT.updateLight: non-integer ID: %s", str)
		wrapLightResponse(w, err, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("LIGHT.createLight: can't parse request: %v", err)
		wrapLightResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	data.Id = id
	new_data, err := light.UpdateLight(db, &data)
	if err != nil {
		if new_data != nil && new_data.Id == light.NOT_FOUND_ID {
			wrapLightResponse(w, err, http.StatusNotFound)
		} else {
			wrapLightResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	json, err := json.Marshal(new_data)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func deleteLight(w http.ResponseWriter, req *http.Request) {
	fmt.Println("/light endpoint called")

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("LIGHT.deleteLight: non-integer ID: %s", str)
		wrapLightResponse(w, err, http.StatusBadRequest)
		return
	}

	tmp, err := light.DeleteLight(db, id)
	if err != nil {
		if tmp < 0 {
			wrapLightResponse(w, err, http.StatusNotFound)
		} else {
			wrapLightResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func findLight(w http.ResponseWriter, req *http.Request) {
	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("LIGHT.loadLight: non-integer ID: %s", str)
		wrapLightResponse(w, err, http.StatusBadRequest)
		return
	}

	data, err := light.FindLight(db, id)
	if err != nil {
		if data != nil && data.Id == light.NOT_FOUND_ID {
			wrapLightResponse(w, err, http.StatusNotFound)
		} else {
			wrapLightResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	json, err := json.Marshal(data)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func addLightTelemetry(w http.ResponseWriter, req *http.Request) {
	var data light.LIGHTtelemetry

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("LIGHT.getLightTelemetry: non-integer ID: %s", str)
		wrapLightResponse(w, err, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("LIGHT.addLightTelemetry: can't parse request: %v", err)
		wrapLightResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	data.LightId = id
	data.CreatedAt = time.Now().UTC()
	tele, err := light.CreateLightTelemetry(db, &data)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(tele)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func getLightTelemetry(w http.ResponseWriter, req *http.Request) {
	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("LIGHT.getLightTelemetry: non-integer ID: %s", str)
		wrapLightResponse(w, err, http.StatusBadRequest)
		return
	}

	data, err := light.GetLightTelemetry(db, id)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	json, err := json.Marshal(data)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func sendLightCommand(w http.ResponseWriter, req *http.Request) {
	var data light.LIGHTstate

	str := req.PathValue("id")
	id, err := strconv.Atoi(str)
	if err != nil {
		log.Printf("LIGHT.sendLightCommand: non-integer ID: %s", str)
		wrapLightResponse(w, err, http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	err = decoder.Decode(&data)
	if err != nil {
		log.Printf("LIGHT.sendLightCommand: can't parse request: %v", err)
		wrapLightResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	if !slices.Contains(light.ALLOWED_STATES, data.State) {
		err = fmt.Errorf("invalid state: %s", data.State)
		log.Print(err.Error())
		wrapLightResponse(w, err, http.StatusUnprocessableEntity)
		return
	}

	system, err := light.FindLight(db, id)
	if err != nil {
		if system != nil && system.Id == light.NOT_FOUND_ID {
			wrapLightResponse(w, err, http.StatusNotFound)
		} else {
			wrapLightResponse(w, err, http.StatusInternalServerError)
		}
		return
	}

	json, err := json.Marshal(data)
	if err != nil {
		wrapLightResponse(w, err, http.StatusInternalServerError)
		return
	}
	fmt.Printf("Sending command to LIGHT system #%d\n", id)
	fmt.Printf("Message: %s\n", json)

	w.WriteHeader(http.StatusOK)
	w.Write(json)
}
