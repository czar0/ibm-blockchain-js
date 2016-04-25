package main

import (
	"errors"
	"fmt"
	"strconv"
	"github.com/openblockchain/obc-peer/openchain/chaincode/shim"
	"encoding/json"
    "encoding/hex"
	"crypto/x509"
	"reflect"
	"encoding/asn1"
	"encoding/pem"
	"net/http"
	"net/url"
	"io/ioutil"
	"time"
	"strings"
	"crypto/md5"
)

//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	SimpleChaincode - A blank struct for use with Shim (An IBM Blockchain included go file used for get/put state
//					  and other IBM Blockchain functions)
//==============================================================================================================================
type SimpleChaincode struct {
}

type ECertResponse struct {
	OK string `json:"OK"`
}

type User struct {

	UserId			string		`json:"userId"`				//Used to register with CA
	Salt			string		`json:"salt"`
	Hash			string		`json:"hash"`
	FirstName		string		`json:"firstName"`
	LastName		string		`json:"lastName"`
	Clients			[]string	`json:"clients"`			//Array of client IDs who they are responsible for, if user is client then this will contain only their client ID
	Address         string      `json:"address"`            //Used for creditcomplex and relationship managers
	PhoneNumber     string      `json:"phoneNumber"`        //Used for creditcomplex and relationship managers
	EmailAddress    string      `json:"emailAddress"`       //Used for creditcomplex and relationship managers

}

type CaseFile struct {

	CaseFileId		string 		`json:"caseFileId"`
	CreationDate	string 		`json:"creationDate"`
	LastUpdate		string 		`json:"lastUpdate"`
	Status			string		`json:"status"`
	Documents		[]string	`json:"documents"`
	Progress        int         `json:"progress"`

}


type LoanFile struct {

	LoanFileId			string			`json:"loanFileId"`
	PropertyDocument	string		`json:"propertyDocument"`

}

type Client struct {

	KvkNumber				string		`json:"kvkNumber"`				//Used as Client ID
	CompanyName				string 		`json:"companyName"`
	RelationshipManager		string		`json:"relationshipManager"`
	CreditComplexManager	string		`json:"creditComplexManager"`
	LoanFiles				[]string	`json:"loanFiles"`				//references to all loan files for this client
	CaseFile 				string		`json:"caseFile"`				//reference to current case file for client
	ClosedCaseFiles			[]string	`json:"closedCaseFiles"`		//references to all previous case files for client
	Address         		string 		`json:"address"`
	MainAccountNumber		string		`json:"mainAccountNumber"`
	BusinessLine			string		`json:"businessLine"`
	Region					string		`json:"region"`
	Country					string		`json:"country"`
	Contact 				string		`json:"contact"`
	PhoneNumber				string		`json:"phoneNumber"`
	EmailAddress			string		`json:"emailAddress"`
	Bc                      string      `json:"bc"`
	Ucr                     int         `json:"ucr"`

}

type PropertyDocument struct {

	PropertyId					string		`json:"propertyId"`
	ObjectType					string		`json:"objectType"`
	Address						string		`json:"address"`
	Number						string		`json:"number"`
	City						string		`json:"city"`
	Postcode					string		`json:"postcode"`
	AssetType					string		`json:"assetType"`
	VVO							int			`json:"vvo"`
	HI							int			`json:"hi"`
	HW							int			`json:"hw"`
	AppraisedMarketValue		int			`json:"appraisedMarketValue"`
	CorrectedMarketValue		int			`json:"correctedMarketValue"`
	BarKK						float32		`json:"barKK"`
	Appraiser					string		`json:"appraiser"`
	TaxationDate				string		`json:"taxationDate"`
	ExpirationDate				string		`json:"expirationDate"`
	TenantsList                 []string    `json:"tenantsList"`    //references to all tenants for this property
	AppraisalStatus				string		`json:"appraisalStatus"`

}

type Tenant struct {

	TenantId            string      `json:"tenantId"`
	ObjectName          string      `json:"objectName"`
	VheDescription      string      `json:"vheDescription"`
	Tenant              string      `json:"tenant"`
	TenantType          string      `json:"tenantType"`
	AssetType           string      `json:"assetType"`
	Vvo                 int         `json:"vvo"`
	Hw                  int         `json:"hw"`
	TotalHw             int         `json:"totalHw"`
	ExpirationDate      string      `json:"expirationDate"`
	RestantLooptijd     int         `json:"restantLooptijd"`
	Status              string      `json:"status"`
	LastValidated       string      `json:"lastValidated"`

}

type Appraisal struct {

	AppraisalId				string  	`json:"appraisalId"`
	PropertyId  			string 		`json:"propertyId"`
	Appraiser  				string 		`json:"appraiser"`
	Requestor  				string 		`json:"requestor"`
	RequestDate  			string 		`json:"requestDate"`
	Client  				string 		`json:"client"`
	Status  				string 		`json:"status"`
	QuoteDate  				string 		`json:"quoteDate"`
	QuotePrice  			float64 	`json:"quotePrice"`
	ClientSignature			string 		`json:"clientSignature"`
    ClientQuoteSignature    string      `json:"clientQuoteSignature"`
	RequestorSignature 		string 		`json:"requestorSignature"`
	AppraiserSignature  	string 		`json:"appraiserSignature"`
    PropertyAddress         string      `json:"propertyAddress"`
    PropertyDescription     string      `json:"propertyDescription"`
    BuildingPlan            string      `json:"buildingPlan"`
    ExternalMeasures        string      `json:"externalMeasures"`
    UsableMeasures          string      `json:"usableMeasures"`
    SharedSpaceMeasures     string      `json:"sharedSpaceMeasures"`
    RentalContracts         string      `json:"rentalContracts"`
    LandQuality             string      `json:"landQuality"`
    BuldingQuality          string      `json:"buildingQuality"`
    PhotosInside            string      `json:"photosInside"`
    PhotosOutside           string      `json:"photosOutside"`
    EnvironmentReport       string      `json:"environmentReport"`
    Valuation               float64     `json:"valuation"`

}

type RentalContract struct {

	ContractId          string      `json:"contractId"`
	ObjectName          string      `json:"objectName"`
	VheDescription      string      `json:"vheDescription"`
	Tenant              string      `json:"tenant"`
	TenantType          string      `json:"tenantType"`
	AssetType           string      `json:"assetType"`
	Vvo                 int         `json:"vvo"`
	Hw                  int         `json:"hw"`
	TotalHw             int         `json:"totalHw"`
	ExpirationDate      string      `json:"expirationDate"`
	RestantLooptijd     int         `json:"restantLooptijd"`
	Status              string      `json:"status"`
	LastValidated       string      `json:"lastValidated"`

}

type EnvironmentReport struct {

	ReportId                string      `json:"reportId"`
	AppraisalId             string      `json:"appraisalId"`
	LocationDescription     string      `json:"locationDescription"`
	Marketability           string      `json:"marketability"`

}


type Signature struct {

	UserId  			string 		`json:"userId"`
	DocumentId			string  	`json:"documentId"`
	DocumentHash  		string 		`json:"documentHash"`
	SignDate  			string 		`json:"signDate"`

}

var AppraisalValidStatus = map[string]bool {
    "REQUESTED": true,
    "QUOTE PENDING": true,
    "QUOTE REJECTED": true,
    "IN PROGRESS": true,
    "CONCEPT PENDING": true,
    "CONCEPT APPROVED": true,
    "CONCEPT REJECTED": true,
    "SIGNED": true,
    "FINALIZED": true,
}

var QuoteValidResponse = map[string]bool {
	"PENDING": true,
	"ACCEPTED": true,
	"REQUESTED": true,
}


//TODO:
//-- Feedback if the user is an employee or client
//-- when used with bluemix, add parameter to assign api url for CA

//temp solutions
var users = []string{"emma1","bob1"}                                //all the user IDs
var clients = []string{"12345678","87654321","12344321"}            //all the client IDs

var appraisersIndexStr = "_appraisers"
var appraisalsIndexStr = "_appraisals"
var signaturesIndexStr = "_signatures"

//==============================================================================================================================
//	Run - Called on chaincode invoke. Takes a function name passed and calls that function. Converts some
//		  initial arguments passed to other things for use in the called function e.g. name -> ecert
//==============================================================================================================================
func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function == "init" {
		return t.init(stub, args)
	} else if function == "add_user" {
		return t.add_user(stub,args)
	} else if function == "add_client" {
		return t.add_client(stub,args)
	} else if function == "add_caseFile" {
		return t.add_caseFile(stub,args)
	} else if function == "add_loanFile" {
		return t.add_loanFile(stub,args)
	} else if function == "add_propertyDocument" {
		return t.add_propertyDocument(stub,args)
	} else if function == "add_tenantDocument" {
		return t.add_tenantDocument(stub,args)
	} else if function == "add_appraisal" {
		return t.add_appraisal(stub,args)
	} else if function == "quote_appraisal" {
		return t.quote_appraisal(stub,args)
	} else if function == "set_casefile_progress" {
		return t.set_casefile_progress(stub,args)
	} else if function == "respond_to_quote" {
		return t.respond_to_quote(stub,args)
    } else if function == "respond_to_concept" {
		return t.respond_to_concept(stub,args)
    } else if function == "create_appraisal_concept" {
		return t.create_appraisal_concept(stub,args)
	}
    
	return nil, errors.New("Received unknown invoke function name")
}
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//
//  args[0] is the function name
//=================================================================================================================================
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	if args[0] == "get_case_file" {
		return t.get_case_file(stub, args)
	} else if args[0] == "get_assigned_clients" {
		return t.get_assigned_clients(stub, args)
	} else if args[0] == "get_all_clients" {
		return t.get_all_clients(stub, args)
	} else if args[0] == "get_user" {
		return t.get_user(stub, args[1])
	} else if args[0] == "get_client" {
		return t.get_client(stub, args)
	} else if args[0] == "authenticate" {
		return t.authenticate(stub, args)
	} else if args[0] == "get_assigned_appraisals" {
		return t.get_assigned_appraisals(stub, args)
	} else if args[0] == "get_appraisal" {
		return t.get_appraisal(stub, args)
	} else if args[0] == "get_all_appraisers" {
		return t.get_all_appraisers(stub, args)
	}

	return nil, errors.New("Received unknown query function name")
}
//=================================================================================================================================
//	 Main - main - Starts up the chaincode
//=================================================================================================================================

func main() {

	err := shim.Start(new(SimpleChaincode))

	if err != nil { fmt.Printf("Error starting SimpleChaincode: %s", err) }
}

//==============================================================================================================================
//	Init Function - Called when the user deploys the chaincode
//==============================================================================================================================

func (t *SimpleChaincode) init(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	return nil, nil
}

//==============================================================================================================================
//		Utility Functions
//==============================================================================================================================

func calculate_hash(args []string) string {
    var str = ""
    for _,v := range args {
        str += v
    }
    hasher := md5.New()
    hasher.Write([]byte(str))
    return hex.EncodeToString(hasher.Sum(nil))
}

//==============================================================================================================================
//		Certificate Authentication
//==============================================================================================================================

func (t *SimpleChaincode) get_ecert(stub *shim.ChaincodeStub, name string) ([]byte, error) {

	var cert ECertResponse

	response, err := http.Get("http://localhost:5000/registrar/"+name+"/ecert") // Calls out to the HyperLedger REST API to get the ecert of the user with that name

	if err != nil { return nil, errors.New("Could not get ecert") }

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)			// Read the response from the http callout into the variable contents

	if err != nil { return nil, errors.New("Could not read body") }

	err = json.Unmarshal(contents, &cert)

	if err != nil { return nil, errors.New("ECert not found for user: "+name) }

	return []byte(string(cert.OK)), nil
}

func (t *SimpleChaincode) get_cert_username(stub *shim.ChaincodeStub, encodedCert string) (string, error) {

	decodedCert, err := url.QueryUnescape(encodedCert);    		// make % etc normal //

	if err != nil { return "", errors.New("Could not decode certificate") }

	pem, _ := pem.Decode([]byte(decodedCert))           	// Make Plain text   //

	x509Cert, err := x509.ParseCertificate(pem.Bytes);

	if err != nil { return "", errors.New("Couldn't parse certificate")	}

	return x509Cert.Subject.CommonName, nil

}

func (t *SimpleChaincode) check_role(stub *shim.ChaincodeStub, encodedCert string) (int64, error) {
	ECertSubjectRole := asn1.ObjectIdentifier{2, 1, 3, 4, 5, 6, 7}

	decodedCert, err := url.QueryUnescape(encodedCert);    		// make % etc normal //

	if err != nil { return -1, errors.New("Could not decode certificate") }

	pem, _ := pem.Decode([]byte(decodedCert))           	// Make Plain text   //

	x509Cert, err := x509.ParseCertificate(pem.Bytes);		// Extract Certificate from argument //

	if err != nil { return -1, errors.New("Couldn't parse certificate")	}

	var role int64
	for _, ext := range x509Cert.Extensions {				// Get Role out of Certificate and return it //
		if reflect.DeepEqual(ext.Id, ECertSubjectRole) {
			role, err = strconv.ParseInt(string(ext.Value), 10, len(ext.Value)*8)

			if err != nil { return -1, errors.New("Failed parsing role: " + err.Error())	}
			break
		}
	}

	return role, nil
}

//==============================================================================================================================
//		Invoke Functions
//==============================================================================================================================

func (t *SimpleChaincode) add_user (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//TODO:
	//-- Add validation
	//-- Add index to userIdxStr

	//Args
	//			0				1
	//		  index		user JSON object (as string)

	var u User

	var err = json.Unmarshal([]byte(args[1]), &u)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }

	eCert,_ := t.get_ecert(stub, u.UserId)

	certRole,_ := t.check_role(stub, string(eCert))

	// check if the user is an appraiser
	if certRole == 3 {
		// get the appraisers index
		appraisersAsBytes, err := stub.GetState(appraisersIndexStr)
		if err != nil { return nil, errors.New("Failed to get Appraisers Index") }

		var appraisersIndex []string
		json.Unmarshal(appraisersAsBytes, &appraisersIndex)

		// append the new insurance to the index
		appraisersIndex = append(appraisersIndex, u.UserId)
		fmt.Println("Added appraiser with id" + u.UserId + " to the appraisers index")
		jsonAsBytes, _ := json.Marshal(appraisersIndex)
		err = stub.PutState(appraisersIndexStr, jsonAsBytes)

		if err != nil { return nil, errors.New("Failed to append new insurance to insurance index") }

		fmt.Println("- Added appraiser with id " + u.UserId)
	}

	err = stub.PutState(args[0], []byte(args[1]))

	if err != nil { return nil, errors.New("Error putting user data into ledger") }

	return nil, nil
}

func (t *SimpleChaincode) add_client (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//TODO:
	//-- Add validation
	//-- Add index to clientIdxStr

	//Args
	//			0				1
	//		  index		client JSON object (as string)

	err := stub.PutState(args[0], []byte(args[1]))

	if err != nil { return nil, errors.New("Error putting client data into ledger") }

	return nil, nil
}

func (t *SimpleChaincode) add_caseFile (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//TODO:
	//-- Add validation
	//-- Add index to caseFileIdxStr

	//Args
	//			0				1
	//		  index		caseFile JSON object (as string)

	err := stub.PutState(args[0], []byte(args[1]))

	if err != nil { return nil, errors.New("Error putting caseFile data into ledger") }

	return nil, nil
}

func (t *SimpleChaincode) add_loanFile (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//TODO:
	//-- Add validation

	//Args
	//			0					1
	//		  index		loanFile JSON object (as string)

	err := stub.PutState(args[0], []byte(args[1]))

	if err != nil { return nil, errors.New("Error putting loanFile data into ledger") }

	return nil, nil
}

func (t *SimpleChaincode) add_propertyDocument (stub *shim.ChaincodeStub, args []string) ([]byte, error) {


	//Args
	//			0				1
	//		  index		propertyDocument JSON object (as string)

	/*
	var p PropertyDocument

	//Check that the data passed for the property document has the correct data structure

	err := json.Unmarshal([]byte(args[1]), &p)

	if err != nil { return nil, errors.New("Property document not in correct JSON structure") }

	//Check if the property document already exists. If it does, return an error to prevent overwriting an existing document

	_, err = stub.GetState(args[0])

	if err != nil { return nil, errors.New("Property Document already exists") }
	*/
	//Store the data

	err := stub.PutState(args[0], []byte(args[1]))

	if err != nil { return nil, errors.New("Error putting Property Document data into ledger") }

	return nil, nil
}

func (t *SimpleChaincode) add_tenantDocument (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//Args
	//			0				1
	//		  index		tenantDocument JSON object (as string)

	//Store the data
	err := stub.PutState(args[0], []byte(args[1]))

	if err != nil { return nil, errors.New("Error putting Tenant Document data into ledger") }

	return nil, nil
}

func (t *SimpleChaincode) add_appraisal (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//Args
	//		       0
	//		   appraisal JSON object (as string)

	var a Appraisal

	// Unmarshal the appraisal
	err := json.Unmarshal([]byte(args[0]), &a)
	if err != nil { return nil, errors.New("Could not convert appraisal to JSON") }

	// Check if the requestor is a user that is allowed to create a new appraisal
	eCert,_ := t.get_ecert(stub, a.Requestor)
	certRole,_ := t.check_role(stub, string(eCert))
	if certRole != 1 { return nil, errors.New("Permission denied - requestor not allowed to add an appraisal") }

	// Get the appraisals index
	appraisalsAsBytes, err := stub.GetState(appraisalsIndexStr)
	if err != nil { return nil, errors.New("Failed to get Appraisals Index") }

	// Unmarshal the appraisal index
	var appraisalsIndex []string
	json.Unmarshal(appraisalsAsBytes, &appraisalsIndex)

	// Create new id for the appraisal
	var newAppraisalId string
	newAppraisalId = "al" + strconv.Itoa(len(appraisalsIndex) + 1)

	// Store new appraisal id in the appraisal
	a.AppraisalId = newAppraisalId

	// append the new appraisal to the index
	appraisalsIndex = append(appraisalsIndex, a.AppraisalId)
	jsonAsBytes, _ := json.Marshal(appraisalsIndex)
	err = stub.PutState(appraisalsIndexStr, jsonAsBytes)
	if err != nil { return nil, errors.New("Error storing new appraisalIndex into ledger") }

	// Get the property for this appraisal
	propertyAsBytes, err := stub.GetState(a.PropertyId)
	if err != nil { return nil, errors.New("Failed to get Property") }

	// Unmarshal the property
	var pd PropertyDocument
	json.Unmarshal(propertyAsBytes, &pd)

	// Update the appraisal status
	pd.AppraisalStatus = "Requested"

	//Convert property JSON object to string
	updatedPropertyDocumentBytes, err := json.Marshal(pd)
	if err != nil { return nil, errors.New("Error marshalling property document data")}

	// Storing the updated property
	err = stub.PutState(a.PropertyId, []byte(updatedPropertyDocumentBytes))
	if err!= nil { return nil, errors.New("Error putting updated property data into the ledger") }

	// create a new array of bites with new appraisal
	newAppraisalBytes, err := json.Marshal(a)
	if err != nil { return nil, errors.New("Error getting marshalling data")}

	// Storing the appraisal
	err = stub.PutState(a.AppraisalId, []byte(newAppraisalBytes))
	if err!= nil { return nil, errors.New("Error putting Appraisal data into ledger") }

	return nil, nil
}

func (t *SimpleChaincode) quote_appraisal (stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	//Args
	//			0				1				 2
	//		  userId	    appraisalId 	 	quote

	eCert,_ := t.get_ecert(stub, args[0])

	certRole,_ := t.check_role(stub, string(eCert))

	// check if the user is an appraiser
	if certRole != 3 { return nil, errors.New("Permission denied - not allowed to quote an appraisal") }


	var u User
	var a Appraisal


	// get user information
	userBytes, err := stub.GetState(args[0])
	if err != nil { return nil, errors.New("Could not retrieve user information") }

	err = json.Unmarshal(userBytes, &u)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }

	// get the appraisal information
	appraisalBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve appraisal information") }

	err = json.Unmarshal(appraisalBytes, &a)
	if err != nil { return nil, errors.New("Could not convert appraisal information to JSON") }

	// check whether the appraiser has permission to access the appraisal
	if u.UserId != a.Appraiser { return nil, errors.New("Permission denied - not allowed to quote this appraisal") }

	// add price to the appraisal quote
	price, err := strconv.ParseFloat(args[2],64)
	if err != nil { return nil, errors.New("Could not parse string price to float") }
	a.QuotePrice = price

	// Update the status of the appraisal
	a.Status = "Quote Pending"

	// Add quote date to the appraisal
	timeStamp := time.Now().UnixNano() / 1000000
	a.QuoteDate = fmt.Sprint(timeStamp)

	// create a new array of bites with updated quote
	newAppraisalBytes, err := json.Marshal(a)
	if err != nil { return nil, errors.New("Error getting marshalling data")}

	errCc := stub.PutState(args[1], []byte(newAppraisalBytes))
	if errCc != nil { return nil, errors.New("Error putting new quote data for the appraisal into ledger") }

	return nil, nil

}

func (t *SimpleChaincode) respond_to_quote (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//Args
	//			0				1				2
	//		  userId	  appraisalId 	 	response

	var a Appraisal
	var c User
	var response = args[2]

	// get the appraisal information
	appraisalBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve appraisal information") }

	err = json.Unmarshal(appraisalBytes, &a)
	if err != nil { return nil, errors.New("Could not convert appraisal information to JSON") }

	// get the client information
	clientBytes, err := stub.GetState(args[0])
	if err != nil { return nil, errors.New("Could not retrieve client information") }

	err = json.Unmarshal(clientBytes, &c)
	if err != nil { return nil, errors.New("Could not convert client information to JSON") }

	// check whether the client has permission to access the appraisal
	if c.Clients[0] != a.Client { return nil, errors.New("Permission denied - not allowed to respond this appraisal") }
	fmt.Println("Permissions accepted")

	// check whether status is quote pending
	if a.Status != "Quote Pending" { return nil, errors.New("No quote to respond") }
	fmt.Println("quote is pending")

	// check if response is a valid status
	if(!AppraisalValidStatus[strings.ToUpper(response)]) { return nil, errors.New("Appraisal status not recognized") }
	fmt.Println("Appraisal status recognized")

	// updating the status
	a.Status = response;

    // sign the appraisal quote
    var hash = calculate_hash([]string{a.AppraisalId,
                                            a.PropertyId,
                                            a.Appraiser,
                                            a.Requestor,
                                            a.RequestDate,
                                            a.Client,
                                            a.QuoteDate,
                                            strconv.FormatFloat(a.QuotePrice, 'f', 2, 64),
                                            a.PropertyAddress,
                                            a.PropertyDescription,
                                            a.BuildingPlan,
                                            a.ExternalMeasures,
                                            a.UsableMeasures,
                                            a.SharedSpaceMeasures,
                                            a.RentalContracts,
                                            a.LandQuality,
                                            a.BuldingQuality,
                                            a.PhotosInside,
                                            a.PhotosOutside,
                                            a.EnvironmentReport,
                                            strconv.FormatFloat(a.Valuation, 'f', 2, 64)})

    signatureId, err := sign_document(stub, []string{c.Clients[0],a.AppraisalId,hash})
    if err != nil { return nil, errors.New("Error signing the document")}
    fmt.Println("Appraisal signed")
    
    a.ClientQuoteSignature = string(signatureId)
    
    
    // create a new array of bytes with updated quote
	newAppraisalBytes, err := json.Marshal(a)
	if err != nil { return nil, errors.New("Error getting marshalling data")}
    fmt.Println("Appraisal marshalled")

    errCc := stub.PutState(args[1], []byte(newAppraisalBytes))
	if errCc != nil { return nil, errors.New("Error putting response to quote data for the appraisal into ledger") }
    fmt.Println("New appraisal stored into ledger")

    
    
    // Next part it to update the status of the property document from Requested to In Progress
	propertyAsBytes, err := stub.GetState(a.PropertyId)
	if err != nil { return nil, errors.New("Failed to get Property") }

	// Unmarshal the property
	var pd PropertyDocument
	json.Unmarshal(propertyAsBytes, &pd)

	// Update the appraisal status
	pd.AppraisalStatus = "In progress"

	//Convert property JSON object to string
	updatedPropertyDocumentBytes, err := json.Marshal(pd)
	if err != nil { return nil, errors.New("Error marshalling property document data")}

	// Storing the updated property
	err = stub.PutState(a.PropertyId, []byte(updatedPropertyDocumentBytes))
	if err!= nil { return nil, errors.New("Error putting updated property data into the ledger") }
    
    return []byte(response), nil
}

func (t *SimpleChaincode) respond_to_concept (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	//Args
	//			0				1				2
	//		  userId	  appraisalId 	 	response

    eCert,_ := t.get_ecert(stub, args[0])

	certRole,_ := t.check_role(stub, string(eCert))

	var a Appraisal
    var c User
    var response = args[2]

	// get the appraisal information
	appraisalBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve appraisal information") }

	err = json.Unmarshal(appraisalBytes, &a)
	if err != nil { return nil, errors.New("Could not convert appraisal information to JSON") }

    // get the client information
	userBytes, err := stub.GetState(args[0])
	if err != nil { return nil, errors.New("Could not retrieve user information") }
    
    err = json.Unmarshal(userBytes, &c)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }
    
	// check whether the client has permission to access the appraisal
    if certRole == 2 {
	   if c.Clients[0] != a.Client { return nil, errors.New("Permission denied - not allowed to respond this appraisal") }
       fmt.Println("Permissions accepted")
    }
    
    // compare hashes

    var s Signature
    // get appraiser signature for the specific document (appraisal)
    signatureBytes, err := stub.GetState(a.AppraiserSignature)
	if err != nil { return nil, errors.New("Could not retrieve signature information") }
    err = json.Unmarshal(signatureBytes, &s)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }
    fmt.Println("Signature info retrieved")

    var appraiserSignature = s.DocumentHash;
    
    // sign the appraisal quote
    var hash = calculate_hash([]string{a.AppraisalId,
                                            a.PropertyId,
                                            a.Appraiser,
                                            a.Requestor,
                                            a.RequestDate,
                                            a.Client,
                                            a.QuoteDate,
                                            strconv.FormatFloat(a.QuotePrice, 'f', 2, 64),
                                            a.PropertyAddress,
                                            a.PropertyDescription,
                                            a.BuildingPlan,
                                            a.ExternalMeasures,
                                            a.UsableMeasures,
                                            a.SharedSpaceMeasures,
                                            a.RentalContracts,
                                            a.LandQuality,
                                            a.BuldingQuality,
                                            a.PhotosInside,
                                            a.PhotosOutside,
                                            a.EnvironmentReport,
                                            strconv.FormatFloat(a.Valuation, 'f', 2, 64)})

    if hash != appraiserSignature { return nil, errors.New("Calculated hash and appraiser hash mismatch") }
    fmt.Println("hash and signature matched")

    signatureId, err := sign_document(stub, []string{c.Clients[0],a.AppraisalId,hash})
    if err != nil { return nil, errors.New("Error signing the document")}
    fmt.Println("Appraisal signed")
    
    if certRole == 1 {
        a.RequestorSignature = string(signatureId)
    } else if certRole == 2 {
        a.ClientSignature = string(signatureId)
    }

    // check whether status is quote pending
	if a.Status != "Concept Pending" { return nil, errors.New("No concept to respond") }
    fmt.Println("concept is pending")

    // Change the status to Signed if the appraisal is signed by everyone else
    if a.RequestorSignature != "" && a.ClientSignature != "" {
        fmt.Println("Appraisal signed by requestor and client, can now be changed to status Signed")

        // check if response is a valid status
        if(!AppraisalValidStatus[strings.ToUpper(response)]) { return nil, errors.New("Appraisal status not recognized") }
        fmt.Println("Appraisal status recognized")

        // updating the status
        a.Status = response;
        
        // Next part it to update the status of the property document from Requested to In Progress
        propertyAsBytes, err := stub.GetState(a.PropertyId)
        if err != nil { return nil, errors.New("Failed to get Property") }

        // Unmarshal the property
        var pd PropertyDocument
        json.Unmarshal(propertyAsBytes, &pd)

        // Update the appraisal status
        pd.AppraisalStatus = "Signed"

        //Convert property JSON object to string
        updatedPropertyDocumentBytes, err := json.Marshal(pd)
        if err != nil { return nil, errors.New("Error marshalling property document data")}

        // Storing the updated property
        err = stub.PutState(a.PropertyId, []byte(updatedPropertyDocumentBytes))
        if err!= nil { return nil, errors.New("Error putting updated property data into the ledger") }
        
        
    }
    
    // create a new array of bytes with updated quote
	newAppraisalBytes, err := json.Marshal(a)
	if err != nil { return nil, errors.New("Error getting marshalling data")}
	fmt.Println("Appraisal marshalled")

	errCc := stub.PutState(args[1], []byte(newAppraisalBytes))
	if errCc != nil { return nil, errors.New("Error putting response to quote data for the appraisal into ledger") }
	fmt.Println("New appraisal stored into ledger")

	return []byte(response), nil
}

func sign_document(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
    //Args
    //			0				1			 2
	//		 userId        documentId    documentHash

	var s Signature

	s.UserId = args[0]
	s.DocumentId = args[1]
	s.DocumentHash = args[2]
	timeStamp := time.Now().UnixNano() / 1000000
	s.SignDate = fmt.Sprint(timeStamp)


    // Get the signatures index
    signaturesAsBytes, err := stub.GetState(signaturesIndexStr)
    if err != nil { return nil, errors.New("Failed to get Signatures Index") }
    fmt.Println("Signature index retrieved")

    // Unmarshal the signatures index
    var signaturesIndex []string
    json.Unmarshal(signaturesAsBytes, &signaturesIndex)
    fmt.Println("Signature index unmarshalled")

    // Create new id for the signature
    var newSignatureId string
    newSignatureId = "sg" + strconv.Itoa(len(signaturesIndex) + 1)

    // append the new signature to the index
    signaturesIndex = append(signaturesIndex, newSignatureId)
    jsonAsBytes, _ := json.Marshal(signaturesIndex)
    err = stub.PutState(signaturesIndexStr, jsonAsBytes)
    if err != nil { return nil, errors.New("Error storing new signaturesIndex into ledger") }
    fmt.Println("New signature index into ledger")

    // marshall new signature object
    signatureBytes, err := json.Marshal(s)
    if err != nil { return nil, errors.New("Error getting marshalling data")}
    fmt.Println("New signature marshalled")

    // key:  sg#
    errCc := stub.PutState(newSignatureId, []byte(signatureBytes))
	if errCc != nil { return nil, errors.New("Error putting signature for document into ledger") }
    fmt.Println("Signature added into ledger")

    return []byte(newSignatureId), nil

}

func (t *SimpleChaincode) create_appraisal_concept (stub *shim.ChaincodeStub, args []string) ([]byte, error) {

    //Args
    //			0				1			         2..14
	//		 userId        appraisalId            all the new values
    fmt.Println("## CREATE CONCEPT ##")
    for _,v := range args {
        fmt.Printf("arg: %s, ",v)
    }
    eCert,_ := t.get_ecert(stub, args[0])

	certRole,_ := t.check_role(stub, string(eCert))

	// check if the user is an appraiser
	if certRole != 3 { return nil, errors.New("Permission denied - not allowed to quote an appraisal") }
    fmt.Println("Appraiser recognized")


	var u User
	var a Appraisal


	// get user information
	userBytes, err := stub.GetState(args[0])
	if err != nil { return nil, errors.New("Could not retrieve user information") }

	err = json.Unmarshal(userBytes, &u)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }
    fmt.Println("User info retrieved")

	// get the appraisal information
	appraisalBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve appraisal information") }
    
    fmt.Println("Appraisal retrieved from ledger")
    err = json.Unmarshal(appraisalBytes, &a)

	if err != nil { return nil, errors.New("Could not convert appraisal information to JSON") }
    fmt.Println("Appraisal unmarshalled")

	// check whether the appraiser has permission to access the appraisal
	if u.UserId != a.Appraiser { return nil, errors.New("Permission denied - not allowed to quote this appraisal") }
    fmt.Println("Appraiser has permissions to read this appraisal")


    // adding new details to the appraisal
    a.PropertyAddress = args[2]
    a.PropertyDescription = args[3]
    // reference to the path of the image
    a.BuildingPlan = args[4]
    a.ExternalMeasures = args[5]
    a.UsableMeasures = args[6]
    a.SharedSpaceMeasures = args[7]

    // Rental contracts not implemented yet

    // split string containing all the object contract. e.g. "{ct1},{ct2},{ct3}"
//    var contractObjects = strings.Split(args[8],",")
//
//    for _,contract := range contractObjects {
//        var c RentalContract
//
//        err = json.Unmarshal([]byte(contract), &c)
//	    if err != nil { return nil, errors.New("Could not convert contract information to JSON") }
//
//        // key: AppraisalIdContractId
//        err := stub.PutState(a.AppraisalId+c.ContractId, []byte(contract))
//	    if err != nil { return nil, errors.New("Error putting contract data for the appraisal into ledger") }
//    }

    a.LandQuality = args[9]
    a.BuldingQuality = args[10]

    // reference to the path of the image
    a.PhotosInside = args[11]
    a.PhotosOutside = args[12]


    // Environment report not implemented yet

//    // add report object to blockchain
//    var e EnvironmentReport
//    err = json.Unmarshal([]byte(args[14]), &e)
//	if err != nil { return nil, errors.New("Could not convert report information to JSON") }
//
//    a.EnvironmentReport = e.ReportId
//
//    // key: AppraisalIdReportId
//    err2 := stub.PutState(a.AppraisalId+e.ReportId, []byte(args[14]))
//    if err2 != nil { return nil, errors.New("Error putting report data for the appraisal into ledger") }

    a.EnvironmentReport = args[13]

    a.Valuation,err = strconv.ParseFloat(args[14],64)
    if err != nil { return nil, errors.New("Could not parse string price to float") }
    fmt.Println("Valuation parsed")

    // sign the appraisal concept
    var hash = calculate_hash([]string{a.AppraisalId,
                                            a.PropertyId,
                                            a.Appraiser,
                                            a.Requestor,
                                            a.RequestDate,
                                            a.Client,
                                            a.QuoteDate,
                                            strconv.FormatFloat(a.QuotePrice, 'f', 2, 64),
                                            a.PropertyAddress,
                                            a.PropertyDescription,
                                            a.BuildingPlan,
                                            a.ExternalMeasures,
                                            a.UsableMeasures,
                                            a.SharedSpaceMeasures,
                                            a.RentalContracts,
                                            a.LandQuality,
                                            a.BuldingQuality,
                                            a.PhotosInside,
                                            a.PhotosOutside,
                                            a.EnvironmentReport,
                                            strconv.FormatFloat(a.Valuation, 'f', 2, 64)})

    signatureId, err := sign_document(stub, []string{a.Appraiser,a.AppraisalId,hash})
    if err != nil { return nil, errors.New("Error signing the document")}
    fmt.Println("Appraisal signed")

    a.AppraiserSignature = string(signatureId)

    // updating the status
    a.Status = "Concept Pending";
    
    // create a new array of bytes with updated appraisal information
	newAppraisalBytes, err := json.Marshal(a)
	if err != nil { return nil, errors.New("Error getting marshalling data")}
    fmt.Println("Appraisal marshalled")

	errCc := stub.PutState(args[1], []byte(newAppraisalBytes))
	if errCc != nil { return nil, errors.New("Error putting response to quote data for the appraisal into ledger") }
    fmt.Println("Appraisal concept added")

	return nil, nil
}

func (t *SimpleChaincode) set_casefile_progress (stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	//Args
	//			0				1				 2
	//		  progress	    caseFileId 	 	   userId

	eCert,_ := t.get_ecert(stub, args[2])

	certRole,_ := t.check_role(stub, string(eCert))

	// check if the user is an appraiser
	if certRole != 1 { return nil, errors.New("Permission denied - not allowed to quote an appraisal") }

	var c CaseFile

	caseFileBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve user information") }

	json.Unmarshal(caseFileBytes, &c)
	//	if err != nil { return nil, errors.New("Could not convert casefile information to JSON") }

	progress, err := strconv.Atoi(args[0]);
	c.Progress = progress;

	// create a new array of bites with updated quote
	newCaseFileBytes, err := json.Marshal(c)
	if err != nil { return nil, errors.New("Error marshalling data")}

	errCc := stub.PutState(args[1], []byte(newCaseFileBytes))
	if errCc != nil { return nil, errors.New("Error putting new casefile data") }

	fmt.Println("updating progress success")

	return nil, nil

}

//==============================================================================================================================
//		Query Functions
//==============================================================================================================================


func (t *SimpleChaincode) get_user(stub *shim.ChaincodeStub, userID string) ([]byte, error){

	bytes, err := stub.GetState(userID)

	if err != nil { return nil,errors.New("Could not retrieve information for this user")}

	return bytes, nil

}

func (t *SimpleChaincode) get_client(stub *shim.ChaincodeStub, args []string) ([]byte, error){

	// TODO Check authorization of user
	//		Check user is permitted to see client info


	//Args
	//			0				1			2
	//		function name	user ID		client ID

	var u User
	var c Client

	var i int
	var j int


	userBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve user information") }

	err = json.Unmarshal(userBytes, &u)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }

	eCert,_ := t.get_ecert(stub, args[1])

	certRole,_ := t.check_role(stub, string(eCert))

	if certRole != 1 {

		found := false

		for _, client := range u.Clients {
			if client == args[2] {
				found = true
			}
		}

		if !found { return nil, errors.New("Permission denied - not allowed to view client information") }

	}


	clientBytes, err := stub.GetState(args[2])
	if err != nil { return nil, errors.New("Could not retrieve information for this client")}

	err = json.Unmarshal(clientBytes, &c)
	if err != nil { return nil, errors.New("Could not convert client information to JSON") }

	//Get case file for this client
	caseFileBytes, err := stub.GetState(c.CaseFile)
	creditComplexManagerAsBytes, err := stub.GetState(c.CreditComplexManager)
	relationshipManagerAsBytes, err := stub.GetState(c.RelationshipManager)

	if err != nil { return nil, errors.New("Could not retrieve information for client's case file") }

	//Replace case file reference with case file information in client information
	c.CaseFile = string(caseFileBytes)
	c.CreditComplexManager = string(creditComplexManagerAsBytes)
	c.RelationshipManager = string(relationshipManagerAsBytes)

	for i=0;i<len(c.LoanFiles); i++ {

		var lf LoanFile

		//Get loan file from ledger
		loanFileBytes, err := stub.GetState(c.LoanFiles[i])
		if err != nil { return nil, errors.New("Could not retrieve loan file information") }

		err = json.Unmarshal(loanFileBytes, &lf)
		if err != nil { return nil, errors.New("Could not convert loan file information to JSON") }

		//Get property document for this loan file
		propertyDocumentBytes, err := stub.GetState(lf.PropertyDocument)
		if err != nil { return nil, errors.New("Could not retrieve property document information") }

		var pd PropertyDocument
		// Unmarshal the property document
		err = json.Unmarshal(propertyDocumentBytes, &pd)
		if err != nil { return nil, errors.New("Could not convert propery document information to JSON") }

		// Do for each tennant inside the property document tenantsList
		for j=0;j<len(pd.TenantsList); j++ {
			//Get tenant document from ledger
			tenantDocumentBytes, err := stub.GetState(pd.TenantsList[j])
			if err != nil { return nil, errors.New("Could not retrieve tenant document information") }

			//Replace tenant document ID with the tenant document information
			pd.TenantsList[j] = string(tenantDocumentBytes)
		}

		//Convert property document JSON object to string
		newPropertyDocument, err := json.Marshal(pd)

		//Replace property document ID with the property document information
		lf.PropertyDocument = string(newPropertyDocument)

		//Convert loan file JSON object to string
		newLoanFile, err := json.Marshal(lf)

		//Replace loan file ID in client information with loan file information
		c.LoanFiles[i] = string(newLoanFile)

	}

	newClientBytes, err := json.Marshal(c)

	return newClientBytes, nil

}

func (t *SimpleChaincode) get_all_clients(stub *shim.ChaincodeStub, args []string) ([]byte, error){


	//Gets all the open case files which are currently in the system, only employees can see all case files

	//Args
	//		0
	//		userID

	userID := args[1]

	eCert,_ := t.get_ecert(stub, userID)

	certRole,_ := t.check_role(stub, string(eCert))

	//Check if the user is an employee, if not return error message
	if certRole != 1 { return nil, errors.New("Permission denied - client can only view their case files")}

	var i int

	var result = "{\"result\":["

	for i=0;i<len(clients);i++ {

		var c Client

		//Get client information
		clientBytes, err := stub.GetState(clients[i])

		if err != nil { return nil, errors.New("Error getting state for " + clients[i])}

		//Convert client information from JSON
		json.Unmarshal(clientBytes, &c)

		//Get the case file for this client
		caseFileBytes, err := stub.GetState(c.CaseFile)
		creditComplexManagerAsBytes, err := stub.GetState(c.CreditComplexManager)
		relationshipManagerAsBytes, err := stub.GetState(c.RelationshipManager)

		c.CaseFile = string(caseFileBytes)
		c.CreditComplexManager = string(creditComplexManagerAsBytes)
		c.RelationshipManager = string(relationshipManagerAsBytes)

		//Convert client JSON object to string
		resClient, err := json.Marshal(c)

		if err != nil { return nil, errors.New("Error getting marshalling data")}

		result += string(resClient) + ", "

	}

	if len(result) > 11 {
		result = result[:len(result)-2]
	}

	result += "]}"

	return []byte(result), nil							//return data will be string in format {"result":[<client and caseFile information>,...]}
}

func (t *SimpleChaincode) get_assigned_clients(stub *shim.ChaincodeStub, args []string) ([]byte, error){


	//Get client and case file information which is assigned to them

	//Args
	//		0
	//		userID

	var u User

	var i int

	username := args[1]

	user,_ := t.get_user(stub, username)

	err := json.Unmarshal(user, &u)

	if err != nil {return nil, errors.New("Error converting user information to JSON")}

	var result = "{\"result\":["

	for i=0;i<len(u.Clients);i++ {

		var c Client

		//Get client information
		clientBytes, err := stub.GetState(u.Clients[i])

		if err != nil { return nil, errors.New("Error getting state for " + u.Clients[i])}

		//Convert client information to JSON
		json.Unmarshal(clientBytes, &c)

		//Get the case file for this client
		caseFileBytes, err := stub.GetState(c.CaseFile)
		creditComplexManagerAsBytes, err := stub.GetState(c.CreditComplexManager)
		relationshipManagerAsBytes, err := stub.GetState(c.RelationshipManager)

		c.CaseFile = string(caseFileBytes)
		c.CreditComplexManager = string(creditComplexManagerAsBytes)
		c.RelationshipManager = string(relationshipManagerAsBytes)

		//Convert client JSON object to string
		resClient, err := json.Marshal(c)

		if err != nil { return nil, errors.New("Error getting marshalling data")}

		result += string(resClient) + ", "

	}

	if len(result) > 11 {
		result = result[:len(result)-2]
	}

	result += "]}"

	return []byte(result), nil

}

func (t *SimpleChaincode) get_case_file(stub *shim.ChaincodeStub, args []string) ([]byte, error){

	//TODO:
	//-- need to add validation on user

	//Will get a specific case file

	//Args
	//			1
	//		caseFileId


	bytes, err := stub.GetState(args[1])

	if err != nil { return nil, errors.New("Error getting from ledger")}

	return bytes, nil

}

func (t *SimpleChaincode) get_assigned_appraisals(stub *shim.ChaincodeStub, args []string) ([]byte, error){

	//Args
	//			0				1
	//		function name	user ID

	var u User

	userBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve user information") }

	err = json.Unmarshal(userBytes, &u)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }

	eCert,_ := t.get_ecert(stub, args[1])
	certRole,_ := t.check_role(stub, string(eCert))

	// Get the appraisal index
	appraisalsAsBytes, err := stub.GetState(appraisalsIndexStr)
	if err != nil { return nil, errors.New("Failed to get Appraisals Index") }
	var appraisalsIndex []string
	json.Unmarshal(appraisalsAsBytes, &appraisalsIndex)

	var i int

	var result = "{\"result\":["

	for i=0;i<len(appraisalsIndex); i++ {

		var a Appraisal

		//Get appraisal information
		appraisalBytes, err := stub.GetState(appraisalsIndex[i])

		if err != nil { return nil, errors.New("Error getting state for " + appraisalsIndex[i])}

		//Convert appraisal information from JSON
		json.Unmarshal(appraisalBytes, &a)

		// Figure out which variable to use to check the appraisals on
		var appraisalParam string
		var userParam string

		if certRole == 1 {
			appraisalParam = a.Requestor
			userParam = args[1]
		} else if certRole == 2 {
			appraisalParam = a.Client
			userParam = u.Clients[0]
		} else if certRole == 3 {
			appraisalParam = a.Appraiser
			userParam = args[1]
		}

		// Check if current appraisal is connected to current suer
		if appraisalParam == userParam {

			//Get the requestor for this appraisal
			requestorAsBytes, err := stub.GetState(a.Requestor)
			a.Requestor = string(requestorAsBytes)

			//Get the appraiser for this appraisal
			appraiserAsBytes, err := stub.GetState(a.Appraiser)
			a.Appraiser = string(appraiserAsBytes)

			//Get the property for this appraisal
			propertyIdAsBytes, err := stub.GetState(a.PropertyId)
			a.PropertyId = string(propertyIdAsBytes)

			//Convert appraisal JSON object to string
			resAppraisel, err := json.Marshal(a)

			if err != nil { return nil, errors.New("Error getting marshalling data")}

			result += string(resAppraisel) + ", "
		}

	}

	if len(result) > 11 {
		result = result[:len(result)-2]
	}

	result += "]}"

	return []byte(result), nil	//return data will be string in format {"result":[appraisalObject,appraisalObject]}

}

func (t *SimpleChaincode) get_appraisal(stub *shim.ChaincodeStub, args []string) ([]byte, error){

	//Args
	//			0				1			     2
	//		function name	appraiser ID     appraisalId

	var u User
	var a Appraisal
	var pd PropertyDocument
	var i int

	userBytes, err := stub.GetState(args[1])
	if err != nil { return nil, errors.New("Could not retrieve user information") }

	err = json.Unmarshal(userBytes, &u)
	if err != nil { return nil, errors.New("Could not convert user information to JSON") }

	//    eCert,_ := t.get_ecert(stub, args[1])
	//
	//	certRole,_ := t.check_role(stub, string(eCert))

	//    if certRole != 3 {
	//		return nil, errors.New("Permission denied - not allowed to view appraisals")
	//	}

	// TODO check if the current user (can ben emma, client or appraiser) is allowed to view this appraisal

	//Get appraisal information
	appraisalBytes, err := stub.GetState(args[2])

	if err != nil { return nil, errors.New("Error getting state for " + args[2])}

	//Convert appraisal information from JSON
	json.Unmarshal(appraisalBytes, &a)

	// Check if appraiser is assigned to this appraisal
	if a.Appraiser != u.UserId {
		return nil, errors.New("Permission denied - current appraiser is not allowed to view this appraisal")
	}

	//Get the requestor for this appraisal
	requestorAsBytes, err := stub.GetState(a.Requestor)
	if err != nil { return nil, errors.New("Could not retrieve requestor document information") }
	a.Requestor = string(requestorAsBytes)

	//Get the client for this appraisal
	clientAsBytes, err := stub.GetState(a.Client)
	if err != nil { return nil, errors.New("Could not retrieve client document information") }
	a.Client = string(clientAsBytes)

	//Get the property for this appraisal
	propertyDocumentBytes, err := stub.GetState(a.PropertyId)
	if err != nil { return nil, errors.New("Could not retrieve property document information") }

	// Unmarshal the property document
	err = json.Unmarshal(propertyDocumentBytes, &pd)
	if err != nil { return nil, errors.New("Could not convert propery document information to JSON") }

	// Do for each tennant inside the property document tenantsList
	for i=0;i<len(pd.TenantsList); i++ {

		//Get tenant document from ledger
		tenantDocumentBytes, err := stub.GetState(pd.TenantsList[i])
		if err != nil { return nil, errors.New("Could not retrieve tenant document information") }

		//Replace tenant document ID with the tenant document information
		pd.TenantsList[i] = string(tenantDocumentBytes)

	}

	//Convert property document JSON object to string
	newPropertyDocument, err := json.Marshal(pd)
	if err != nil { return nil, errors.New("Error getting marshalling data property document") }

	a.PropertyId = string(newPropertyDocument)

	//Convert appraisal JSON object to string
	result, err := json.Marshal(a)

	if err != nil { return nil, errors.New("Error getting marshalling data")}

	return []byte(result), nil	//return data will be string in format {"result":[appraisalObject,appraisalObject]}

}

func (t *SimpleChaincode) get_all_appraisers(stub *shim.ChaincodeStub, args []string) ([]byte, error){


	// Get a list of all the appraisers

	//Args
	//		      0            1
	//		function name    userId

	eCert,_ := t.get_ecert(stub, args[1])

	certRole,_ := t.check_role(stub, string(eCert))

	//Check if the user is employee, if not return error message
	if certRole != 1 { return nil, errors.New("Permission denied - user does not have permissions to do this request")}

	appraisersAsBytes, err := stub.GetState(appraisersIndexStr)
	if err != nil { return nil, errors.New("Failed to get Appraisers Index") }
	var appraisersIndex []string
	json.Unmarshal(appraisersAsBytes, &appraisersIndex)

	var i int

	var result = "{\"result\":["

	for i=0;i<len(appraisersIndex);i++ {

		var u User

		user,_ := t.get_user(stub, appraisersIndex[i])
		err := json.Unmarshal(user, &u)
		if err != nil { return nil, errors.New("Error unmarshalling data")}

		//Convert client JSON object to string
		resUser, err := json.Marshal(u)

		if err != nil { return nil, errors.New("Error getting marshalling data")}

		result += string(resUser)

		if i+1 < len(appraisersIndex) { result += ", " }

	}

	result += "]}"

	return []byte(result), nil

}

func (t *SimpleChaincode) authenticate(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	// Args
	//	1		2
	//	userId	password

	var u User

	username := args[1]

	user, err := t.get_user(stub, username)

	if err != nil { return []byte(`{ "authenticated": false, "certRole": -1  }`), nil}

	_, err = http.Get("http://localhost:5000/registrar/"+username)

	if err != nil { return []byte(`{ "authenticated": false, "certRole": -1  }`), nil}

	eCert, err := t.get_ecert(stub, username)

	certRole,_ := t.check_role(stub, string(eCert))

	//Check if the user is an employee, if not return error message
	err = json.Unmarshal(user, &u)

	if err != nil { return []byte(`{ "authenticated": false, "certRole": -1}`), nil }

	certRoleAsString := strconv.FormatInt(certRole, 10)

	var str string

	// If user is client, pass
	if certRole == 2 {
		var c Client

		// get client based on u.Clients
		clientBytes, err := stub.GetState(u.Clients[0])
		if err != nil { return nil, errors.New("Could not retrieve information for this client")}

		err = json.Unmarshal(clientBytes, &c)
		if err != nil { return nil, errors.New("Could not convert client information to JSON") }

		str = `{ "authenticated": true, "certRole": ` + certRoleAsString + `,"user": ` + string(user) + `,"caseFileID":"` + c.CaseFile + `"}`
	} else {
		str = `{ "authenticated": true, "certRole": ` + certRoleAsString + `,"user": ` + string(user) + ` }`
	}

	// validate passwords
	return []byte(str), nil
}
