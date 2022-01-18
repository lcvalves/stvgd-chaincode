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
