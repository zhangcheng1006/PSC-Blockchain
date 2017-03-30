package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// Define the needed constants
var KIDNERTABLENAME = "KidnerTable"
var ACTIVE = "active"
var MATCH = "matched"
var NOTMATCH = "notmatched"
var INACTIVE = "inactive"
var logger = shim.NewLogger("kidner")
var KIDNERCOLOMNBYTES = 2

/*
Structure of KidnerObject
State: represent the state of couple, can be "active" or "inactive"
CoupleID: the identification of the couple, it is unique
DonorHash: the hashcode of donor's information which is used for the match
RecieverHash: the hashcode of reciever's information which is used for the match
Match: represent the state of couple, can be "matched" or "notmatched"
*/
type KidnerObject struct {
	State        string
	CoupleID     string
	DonorHash    string
	RecieverHash string
	Match        string
}

// Define the chaincode structure
type KidnerChaincode struct {
}

/*
InvokeFunction() which is consist of three sub-functions:
Update(): to update the informations of couple, which is the hashcode
FindMatch(): try to find a match of a couple
CreateNewUser(): to create a new user
*/
func InvokeFunction(fname string) func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("InvokeFunction() : calling method -")
	InvokeFunc := map[string]func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){
		"Update":        Update,
		"FindMatch":     FindMatch,
		"CreateNewUser": CreateNewUser,
	}
	return InvokeFunc[fname]
}

/*
Init() which is called only when we deploy the chaincode
we destroy the old database and create a new table as the database
*/
func (t *KidnerChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("Init(): calling method -")
	var err error
	logger.Debug("Init")
	err = InitKidnerTable(stub)
	if err != nil {
		return nil, fmt.Errorf("Init(): InitLedger of KidnerTable Failed.")
	}
	logger.Debug("Init() Initialization Complete: ", args)
	return []byte("Init(): Initialization Complete"), nil
}

/*
InitKidnerTable() which is used to initailise the table which has three columns:
column1: "State" as a key
column2: "CoupleID" as a key
column3: "Json" including all the other information in the format "json", which is not a key
*/
func InitKidnerTable(stub shim.ChaincodeStubInterface) error {
	logger.Debug("InitKidnerTable() : calling method - ")
	var columnDefsKidnerTable []*shim.ColumnDefinition
	err := stub.DeleteTable(KIDNERTABLENAME)
	if err != nil {
		return fmt.Errorf("Init(): DeleteTable of KidnerTable Failed.")
	}
	logger.Debug("table: ", KIDNERTABLENAME, " deleted.")
	column1 := shim.ColumnDefinition{Name: "State", Type: shim.ColumnDefinition_STRING, Key: true}
	column2 := shim.ColumnDefinition{Name: "CoupleID", Type: shim.ColumnDefinition_STRING, Key: true}
	column3 := shim.ColumnDefinition{Name: "Json", Type: shim.ColumnDefinition_BYTES, Key: false}
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column1)
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column2)
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column3)
	err = stub.CreateTable(KIDNERTABLENAME, columnDefsKidnerTable)
	if err == nil {
		logger.Debug("table: ", KIDNERTABLENAME, " created.")
	} else {
		logger.Error("table: ", KIDNERTABLENAME, " not created.")
	}
	return err
}

// Main function
func main() {
	logger.SetLevel(shim.LogDebug)
	logLevel, _ := shim.LogLevel(os.Getenv("SHIM_LOGGING_LEVEL"))
	shim.SetLoggingLevel(logLevel)
	logger.Debug("main() : ******************************************************************************************")
	logger.Debug("main() : calling method -")

	err := shim.Start(new(KidnerChaincode))
	if err != nil {
		logger.Error("Error while starting chaincode: %s", err)
	}
}

/*
function Invoke(): test which sub-function to use (Update(), FindMatch(), CreateNewUser())
*/
func (t *KidnerChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("Invoke() : calling method -")
	logger.Debug("Invoke Args supplied : ", args)
	InvokeRequest := InvokeFunction(function)
	if InvokeRequest != nil {
		buff, err := InvokeRequest(stub, function, args)
		logger.Debug("Invoke response: buff: ", fmt.Sprintf("%s", buff), "err: ", fmt.Sprintf("%s", err))
		return buff, err
	} else {
		logger.Error("Invoke() Invalid function call : ", function)
		return nil, errors.New("Invoke() : Invalid function call : " + function)
	}
}

/*
function CreateNewUser(): Create a new user by passing the parameters: CoupleID, DonorHash, RecieverHash
"State" is defined as "active" by default
"Match" is defined as "notmatched" by default
*/
func CreateNewUser(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("\n\nCreateNewUser() : calling method -")
	logger.Debug("Args: CoupleID:", args[0], "DonorHash:", args[1], "RecieverHash", args[2])
	KidnerUser, err := CreateKidnerObject(stub, args)
	if err != nil {
		logger.Error("CreateNewUser() : Failed Cannot create a new user : ", strings.Join(args, ", "))
		return nil, err
	}
	buff, err := KidnertoJSON(KidnerUser)
	if err != nil {
		logger.Error("CreateNewUser() : Failed Cannot create a new user : ", strings.Join(args, ", "))
		return nil, err
	}
	err = InsertRowKidnerTable(stub, KidnerUser, buff)
	if err != nil {
		logger.Error("CreateNewUser() : write error while inserting record ", err)
		return nil, err
	} else {
		return buff, err
	}
}

/*
function CreateKidnerObject() which create a kidner object and used for the function CreateNewUser()
*/
func CreateKidnerObject(stub shim.ChaincodeStubInterface, args []string) (KidnerObject, error) {
	logger.Debug("CreateKidnerObject() : calling method -")
	var aKidner KidnerObject
	if len(args) != 3 {
		logger.Error("CreateKidnerObject(): Incorrect number of arguments. Expecting 3 ")
		return aKidner, errors.New("CreateKidnerObject() : Incorrect number of arguments. Expecting 3 ")
	}
	var CoupleID = args[0]
	var DonorHash = args[1]
	var RecieverHash = args[2]

	aKidner = KidnerObject{ACTIVE, CoupleID, DonorHash, RecieverHash, NOTMATCH}
	logger.Debug("CreateKidnerObject() : Kidner Object : ", aKidner)
	return aKidner, nil
}

// function InsertRowKidnerTable() which insert a row to the table
func InsertRowKidnerTable(stub shim.ChaincodeStubInterface, kidner KidnerObject, buff []byte) error {
	logger.Debug("InsertRowKidnerTable() : calling method -")
	row := BuildRowKidnerTable(kidner, buff)
	ok, err := stub.InsertRow(KIDNERTABLENAME, row)
	if err != nil {
		return fmt.Errorf("InsertTableRow operation failed. %s", err)
	}
	if !ok {
		return errors.New("InsertTableRow operation failed. Row with given key already exists")
	}
	return nil
}

// function BuildRowKidnerTable() which build a row from a kidner object
func BuildRowKidnerTable(kidner KidnerObject, buff []byte) shim.Row {
	logger.Debug("BuildRowKidnerTable() : calling method -")
	var columns []*shim.Column
	state := shim.Column{Value: &shim.Column_String_{String_: kidner.State}}
	columns = append(columns, &state)
	col2 := shim.Column{Value: &shim.Column_String_{String_: kidner.CoupleID}}
	columns = append(columns, &col2)
	js := shim.Column{Value: &shim.Column_Bytes{Bytes: []byte(buff)}}
	columns = append(columns, &js)
	row := shim.Row{Columns: columns}
	logger.Debug("rowString= ", fmt.Sprintf("%s", row))
	return row
}

// function KidnertoJSON() which convert a kidner object to a json
func KidnertoJSON(kidner KidnerObject) ([]byte, error) {
	logger.Debug("KidnertoJSON() : calling method -")
	ajson, err := json.Marshal(kidner)
	if err != nil {
		logger.Error("KidnertoJSON error: ", err)
		return nil, err
	}
	logger.Debug("KidnertoJSON created: ", fmt.Sprintf("%s", ajson))
	return ajson, nil
}

// function JSONtoKidner() which convert a json to a kidner object
func JSONtoKidner(jsonkidner []byte) (KidnerObject, error) {
	logger.Debug("JSONtoKidner() : calling method -")
	logger.Debug("json: ", fmt.Sprintf("%s", jsonkidner))
	ur := KidnerObject{}
	err := json.Unmarshal(jsonkidner, &ur)
	if err != nil {
		logger.Error("JSONtoKidner error: ", err)
		return ur, err
	}
	logger.Debug("JSONtoKidner created: ", ur)
	return ur, err
}

/*
Function Query() which is used to request the information of a couple by passing the parameter: CoupleID
If the CoupleID is not correct or this couple is inactive, return error
*/
func (t *KidnerChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("\n\nQuery() : calling method -")
	var err error

	if len(args) != 1 {
		logger.Error("Query(): Incorrect number of arguments. Expecting coupleID")
		return nil, errors.New("Query(): Incorrect number of arguments. Expecting coupleID")
	}
	CoupleID := args[0]
	aKidner, err := GetKidnerObject(stub, args)
	if err != nil {
		logger.Error("Query() : Failed to Query Object ")
		indic := "Error: Failed to get Object Data for " + CoupleID + ". %s"
		return nil, fmt.Errorf(indic, err)
	}
	if aKidner.CoupleID == "" {
		logger.Error("Query() : Incorrect Query Object ")
		indic := "Error: Incorrect information about the key for " + CoupleID + ". %s"
		return nil, fmt.Errorf(indic, err)
	}
	buff, err := KidnertoJSON(aKidner)
	if err != nil {
		logger.Error("Query() : KidnertoJSON failed. ")
		indic := "Error: KidnertoJSON failed about the key for " + CoupleID + ". %s"
		return nil, fmt.Errorf(indic, err)
	}
	logger.Debug("Query() : Response : Successfull -")
	return buff, nil
}

// function GetKidnerObject() which get a kidner object from the table
func GetKidnerObject(stub shim.ChaincodeStubInterface, args []string) (KidnerObject, error) {
	initial := KidnerObject{}
	newArgs := []string{ACTIVE, args[0]}
	row, err := GetRowFromKidnerTable(stub, newArgs)
	if err != nil {
		return initial, fmt.Errorf("GetKidnerObject operation failed : CoupleID is not correct or the couple that you look for is not active. %s", err)
	}
	kidner, err := JSONtoKidner(row)
	if err != nil {
		logger.Error("GetKidnerObject Failed : Ummarshall error")
		return initial, fmt.Errorf("JSONtoKidner operation failed. %s", err)
	}
	return kidner, nil
}

// function GetRowFromKidnerTable() which get a row from the table
func GetRowFromKidnerTable(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	logger.Debug("GetRowKidnerTable() : calling method -")
	var columns []shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: args[0]}}
	columns = append(columns, col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: args[1]}}
	columns = append(columns, col2)
	logger.Debug("getRow: args=", args)
	row, err := stub.GetRow(KIDNERTABLENAME, columns)
	logger.Debug("rowString: ", fmt.Sprintf("%s", row))
	if err != nil {
		logger.Error("Error retrieving data record for Keys = ", args)
		return nil, err
	}
	logger.Debug("Length or number of rows retrieved ", len(row.Columns))

	if len(row.Columns) == 0 {
		jsonResp := "{\"Error\":\"Failed retrieving data " + args[2] + ". \"}"
		logger.Error("Error retrieving data record for Keys = ", args, "Error : ", jsonResp)
		return nil, errors.New(jsonResp)
	}
	Avalbytes := row.Columns[KIDNERCOLOMNBYTES].GetBytes()
	return Avalbytes, nil
}

/*
Function Update() which update the information of a couple by passing three parameters:
CoupleID, the information to changer, the value to put
*/
func Update(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var CoupleID string
	logger.Debug("\n\nUpdate() : calling method -")
	logger.Debug("Args: CoupleID:", args[0])
	CoupleID = args[0]
	if len(args) != 3 {
		logger.Error("Update(): Incorrect number of arguments. Expecting 3")
		return nil, errors.New("Update(): Incorrect number of arguments. Expecting 3")
	}
	newArgs := []string{CoupleID}
	aKidner, err := GetKidnerObject(stub, newArgs)
	var bKidner KidnerObject
	bKidner = aKidner
	if args[1] == "DonorHash" {
		bKidner = KidnerObject{aKidner.State, aKidner.CoupleID, args[2], aKidner.RecieverHash, aKidner.Match}
	}
	if args[1] == "RecieverHash" {
		bKidner = KidnerObject{aKidner.State, aKidner.CoupleID, aKidner.DonorHash, args[2], aKidner.Match}
	}
	if args[1] == "Match" {
		if args[2] != NOTMATCH {
			logger.Error("Update(): Incorrect argument of Match. Expecting notmatched")
			return nil, errors.New("Update(): Incorrect argument of Match. Expecting notmatched")
		}
		bKidner = KidnerObject{aKidner.State, aKidner.CoupleID, aKidner.DonorHash, aKidner.RecieverHash, args[2]}
	}
	if args[1] == "State" {
		if args[2] != INACTIVE {
			logger.Error("Update(): Incorrect argument of State. Expecting inactive")
			return nil, errors.New("Update(): Incorrect argument of State. Expecting inactive")
		}
		bKidner = KidnerObject{args[2], aKidner.CoupleID, aKidner.DonorHash, aKidner.RecieverHash, aKidner.Match}
	}
	json, err := KidnertoJSON(bKidner)
	if err != nil {
		logger.Error("Update() : KidnertoJSON failed. ")
		jsonResp := "{\"Error\":\"KidnertoJSON failed about the new key for " + CoupleID + "\"}"
		return nil, errors.New(jsonResp)
	}
	row := BuildRowKidnerTable(bKidner, json)
	if args[1] == "DonorHash" || args[1] == "RecieverHash" || args[1] == "Match" {
		ok, err := stub.ReplaceRow(KIDNERTABLENAME, row)
		if err != nil {
			return nil, fmt.Errorf("updateKidnerTable operation failed. %s", err)
		}
		if !ok {
			return nil, errors.New("updateKidnerTable operation failed. Row already exists")
		}
	}
	if args[1] == "State" && args[2] == INACTIVE {
		err := DeleteRowKidnerTable(stub, []string{ACTIVE, aKidner.CoupleID})
		if err != nil {
			return nil, err
		}
		ok, err := stub.InsertRow(KIDNERTABLENAME, row)
		if err != nil {
			return nil, fmt.Errorf("updateKidnerTable operation failed. %s", err)
		}
		if !ok {
			return nil, errors.New("updateKidnerTable operation failed. Row already exists")
		}
	}
	return []byte("Update Successfully!"), nil
}

// function DeleteRowKidnerTable() which delete a row from the table
func DeleteRowKidnerTable(stub shim.ChaincodeStubInterface, keys []string) error {
	logger.Debug("DeleteRowKidnerTable() : calling method -")
	var columns []shim.Column
	for i := 0; i < len(keys); i++ {
		col := shim.Column{Value: &shim.Column_String_{String_: keys[i]}}
		columns = append(columns, col)
	}
	err := stub.DeleteRow(KIDNERTABLENAME, columns)
	if err != nil {
		return fmt.Errorf("deleteRowKidnerTable operation failed. %s", err)
	}
	return nil
}

/*
Function FindMatch() which try to find a match for a couple by passing the parameter: CoupleID
*/
func FindMatch(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("\n\nFindMatch() : calling method -")
	if len(args) != 1 {
		logger.Error("FindMatch(): Incorrect number of arguments. Expecting CoupleID ")
		return nil, errors.New("FindMatch(): Incorrect number of arguments. Expecting CoupleID ")
	}
	logger.Debug("Args: CoupleID:", args[0])
	var CoupleID = args[0]
	newArgs := []string{CoupleID}
	aKidner, err := GetKidnerObject(stub, newArgs)
	if err != nil {
		return nil, err
	}
	if aKidner.State == INACTIVE {
		return nil, errors.New("Error: The couple that you try to find a match is not active!")
	}
	newargs := []string{ACTIVE}
	kidners, err := GetKidnerObjects(stub, newargs)
	if err != nil {
		return nil, err
	}
	var tlist []KidnerObject
	for i := 0; i < len(kidners); i++ {
		if aKidner.DonorHash == kidners[i].RecieverHash && aKidner.RecieverHash == kidners[i].DonorHash {
			kidners[i].Match = MATCH
			js, err := KidnertoJSON(kidners[i])
			row := BuildRowKidnerTable(kidners[i], js)
			ok, err := stub.ReplaceRow(KIDNERTABLENAME, row)
			if err != nil {
				return nil, fmt.Errorf("updateKidnerTable operation failed. %s", err)
			}
			if !ok {
				return nil, errors.New("updateKidnerTable operation failed. Row already exists")
			}
			tlist = append(tlist, kidners[i])
		}
	}
	if len(tlist) > 0 {
		aKidner.Match = MATCH
		js, err := KidnertoJSON(aKidner)
		row := BuildRowKidnerTable(aKidner, js)
		ok, err := stub.ReplaceRow(KIDNERTABLENAME, row)
		if err != nil {
			return nil, fmt.Errorf("updateKidnerTable operation failed. %s", err)
		}
		if !ok {
			return nil, errors.New("updateKidnerTable operation failed. Row already exists")
		}
	}
	jsonRows, _ := json.Marshal(tlist)
	return jsonRows, nil
}

//function GetKidnerObjects() which get several kidner objects from the table
func GetKidnerObjects(stub shim.ChaincodeStubInterface, args []string) ([]KidnerObject, error) {
	rows, err := GetRowsFromKidnerTable(stub, args)
	if err != nil {
		return nil, fmt.Errorf("GetKidners operation failed : %s", err)
	}
	tlist := make([]KidnerObject, len(rows))
	for i := 0; i < len(rows); i++ {
		ts := rows[i].Columns[KIDNERCOLOMNBYTES].GetBytes()
		kidner, err := JSONtoKidner(ts)
		if err != nil {
			logger.Error("GetKidners() Failed : Ummarshall error")
			return nil, fmt.Errorf("GetKidners() operation failed. %s", err)
		}
		tlist[i] = kidner
	}
	return tlist, nil
}

// function GetRowsFromKidnerTable() which get several rows from the table
func GetRowsFromKidnerTable(stub shim.ChaincodeStubInterface, args []string) ([]shim.Row, error) {
	logger.Debug("GetRowsFromKidnerTable() : calling method -")
	var columns []shim.Column
	for i := 0; i < len(args); i++ {
		col := shim.Column{Value: &shim.Column_String_{String_: args[i]}}
		columns = append(columns, col)
	}
	logger.Debug("getKidnerRows: args=", args)
	rowChannel, err := stub.GetRows(KIDNERTABLENAME, columns)
	if err != nil {
		return nil, fmt.Errorf("GetList operation failed. %s", err)
	}
	var rows []shim.Row
	for {
		select {
		case row, ok := <-rowChannel:
			if !ok {
				rowChannel = nil
			} else {
				rows = append(rows, row)
			}
		}
		if rowChannel == nil {
			break
		}
	}
	logger.Debug("Number of rows retrieved : ", len(rows))
	for i := 0; i < len(rows); i++ {
		logger.Debug("rowString: ", fmt.Sprintf("%s", rows[i]))
	}
	return rows, nil
}
