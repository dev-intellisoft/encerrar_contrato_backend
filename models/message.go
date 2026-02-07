package models

import "encoding/json"

// type SocketPatient struct {
// 	Name          string   `json:"name"`
// 	CPF           string   `json:"cpf"`
// 	DOB           string   `json:"birthDate"`
// 	Possibilities []string `json:"possibilities"`
// }

type SocketMessage struct {
	// Patient SocketPatient   `json:"patient"`
	Data  json.RawMessage `json:"data"`
	From  string          `json:"from"`
	Event string          `json:"event"`
	To    string          `json:"to"`
}
