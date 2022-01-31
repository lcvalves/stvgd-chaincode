/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

func main() {
	stvgdContract := new(StvgdContract)
	stvgdContract.Info.Version = "0.3.0"
	stvgdContract.Info.Description = "STVgoDigital Contract"
	stvgdContract.Info.License = new(metadata.LicenseMetadata)
	stvgdContract.Info.License.Name = "Apache-2.0"
	stvgdContract.Info.Contact = new(metadata.ContactMetadata)
	stvgdContract.Info.Contact.Name = "Luís Alves"
	stvgdContract.Info.Contact.Email = "luas@ipvc.pt"
	stvgdContract.Info.Contact.URL = "https://github.com/lcvalves/"

	chaincode, err := contractapi.NewChaincode(stvgdContract)
	chaincode.Info.Title = "STVgoDigital Chaincode"
	chaincode.Info.Version = "0.3.0"

	if err != nil {
		panic("Could not create chaincode from StvgdContract." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
