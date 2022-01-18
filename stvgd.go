/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

// Lot stores information about the lots/batches in the supply chain
type Lot struct {
	DocType       string  `json:"docType"` //docType ("lot") is used to distinguish the various types of objects in state database
	ID            string  `json:"ID"`      //the field tags are needed to keep case from bouncing around
	LotType       string  `json:"lotType"`
	ProdActivity  string  `json:"prodActivity,omitempty" metadata:",optional"` // ProdActivity can be null / empty for lots that come from Production Activities outside of the system
	Amount        float32 `json:"amount"`
	Unit          string  `json:"unit"`
	ProdUnit      string  `json:"prodUnit"` // "owner"
	LotInternalID string  `json:"lotInternalID"`
}

/* // ProdActivity stores information about the production activities in the supply chain
type ProdActivity struct {
	DocType          string             `json:"docType"` //docType ("prodActivity") is used to distinguish the various types of objects in state database
	ID               string             `json:"ID"`      //the field tags are needed to keep case from bouncing around
	ActivityType     string             `json:"activityType"`
	ProdUnit         string             `json:"prodUnit"`
	InputLots        map[string]float32 `json:"inputLots"` //,omitempty" metadata:",optional"` // inputLots nullable for ProdActivities that create new lots as raw material not sourced from outside the system
	OutputLot        string             `json:"outputLot"`
	ActivityEndDate  string             `json:"activityEndDate"`
	CompanyLegalName string             `json:"companyLegalName"`
	Location         string             `json:"location"`
	EnvScore         float32            `json:"envScore"`
} */
