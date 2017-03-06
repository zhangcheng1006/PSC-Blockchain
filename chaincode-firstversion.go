package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

var KIDNERTABLENAME = "KidnerTable"
var ACTIF = "actif"
var UNACTIF = "unactif"
var logger = shim.NewLogger("kidner")
var KIDNERCOLOMNBYTES = 5

type KidnerObject struct {
	DonneurID    string
	DonneurHash  string
	ReceveurID   string
	ReceveurHash string
	State        string
}

type KidnerChaincode struct {
}

func InvokeFunction(fname string) func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("InvokeFunction() : calling method -")
	InvokeFunc := map[string]func(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error){
		"MiseAJour":     MiseAJour,
		"TrouveMatch":   TrouveMatch,
		"CreateNewUser": CreateNewUser,
	}
	return InvokeFunc[fname]
}

func (t *KidnerChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("Init(): calling method -")
	var err error
	logger.Debug("Init")
	err = InitKidnerTable(stub)
	if err != nil {
		return nil, fmt.Errorf("Init(): InitLedger of KidnerTable  Failed ")
	}
	logger.Debug("Init() Initialization Complete  : ", args)
	return []byte("Init(): Initialization Complete"), nil
}

func InitKidnerTable(stub shim.ChaincodeStubInterface) error {
	logger.Debug("InitKidnerTable() : calling method - ")
	var columnDefsKidnerTable []*shim.ColumnDefinition
	err := stub.DeleteTable(KIDNERTABLENAME)
	if err != nil {
		return fmt.Errorf("Init(): DeleteTable of KidnerTable Failed.")
	}
	logger.Debug("table: ", KIDNERTABLENAME, " deleted.")
	column1 := shim.ColumnDefinition{Name: "DonneurID", Type: shim.ColumnDefinition_STRING, Key: true}
	column2 := shim.ColumnDefinition{Name: "DonneurHash", Type: shim.ColumnDefinition_STRING, Key: true}
	column3 := shim.ColumnDefinition{Name: "ReceveurID", Type: shim.ColumnDefinition_STRING, Key: true}
	column4 := shim.ColumnDefinition{Name: "ReceveurHash", Type: shim.ColumnDefinition_STRING, Key: true}
	column5 := shim.ColumnDefinition{Name: "State", Type: shim.ColumnDefinition_STRING, Key: true}
	column6 := shim.ColumnDefinition{Name: "Kidner", Type: shim.ColumnDefinition_STRING, Key: false}
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column1)
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column2)
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column3)
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column4)
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column5)
	columnDefsKidnerTable = append(columnDefsKidnerTable, &column6)
	err = stub.CreateTable(KIDNERTABLENAME, columnDefsKidnerTable)
	if err == nil {
		logger.Debug("table: ", KIDNERTABLENAME, " created.")
	} else {
		logger.Error("table: ", KIDNERTABLENAME, " not created.")
	}
	return err
}

func main() {
	logger.SetLevel(shim.LogDebug)
	logLevel, _ := shim.LogLevel(os.Getenv("SHIM_LOGGING_LEVEL"))
	shim.SetLoggingLevel(logLevel)
	logger.Debug("main() : ******************************************************************************************")
	logger.Debug("main() : calling method -")

	err := shim.Start(new(KidnerChaincode))
	if err != nil {
		logger.Error("Error starting chaincode: %s", err)
	}
}

func (t *KidnerChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("Invoke() : calling method -")
	logger.Debug("Invoke Args supplied : ", args)
	InvokeRequest := InvokeFunction(function)
	if InvokeRequest != nil {
		buff, err := InvokeRequest(stub, function, args)
		logger.Debug("invoke response: buff: ", fmt.Sprintf("%s", buff), "err: ", fmt.Sprintf("%s", err))
		return buff, err
	} else {
		logger.Error("Invoke() Invalid function call : ", function)
		return nil, errors.New("Invoke() : Invalid function call : " + function)
	}
}

//a changer
func CreateNewUser(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("\n\nCreateNewUser() : calling method -")
	logger.Debug("Args: DonneurID:", args[0], "DonneurHash:", args[1], "ReceveurID:", args[2], "ReceveurHash:", args[3])
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

func CreateKidnerObject(stub shim.ChaincodeStubInterface, args []string) (KidnerObject, error) {
	logger.Debug("CreateKidnerObject() : calling method -")
	var aKidner KidnerObject
	if len(args) != 4 {
		logger.Error("CreateKidnerObject(): Incorrect number of arguments. Expecting 4 ")
		return aKidner, errors.New("CreateKidnerObject() : Incorrect number of arguments. Expecting 4 ")
	}
	// var state string = ACTIF
	//var consentID = Generate_uuid()
	//var consentID = stub.GetTxID()
	var DonneurID = args[0]
	var DonneurHash = args[1]
	var ReceveurID = args[2]
	var ReceveurHash = args[3]

	aKidner = KidnerObject{DonneurID, DonneurHash, ReceveurID, ReceveurHash, ACTIF}
	logger.Debug("CreateKidnerObject() : Kidner Object : ", aKidner)
	return aKidner, nil
}

func InsertRowKidnerTable(stub shim.ChaincodeStubInterface, kidner KidnerObject, buff []byte) error {
	logger.Debug("InsertRowKidnerTable() : calling method -")
	row := BuildRowKidnerTable(kidner, buff)
	ok, err := stub.InsertRow(KIDNERTABLENAME, row)
	if err != nil {
		return fmt.Errorf("insertTableRow operation failed. %s", err)
	}
	if !ok {
		return errors.New("insertTableRow operation failed. Row with given key already exists")
	}
	return nil
}

func BuildRowKidnerTable(kidner KidnerObject, buff []byte) shim.Row {
	logger.Debug("BuildRowKidnerTable() : calling method -")
	var columns []*shim.Column
	col1 := shim.Column{Value: &shim.Column_String_{String_: kidner.DonneurID}}
	columns = append(columns, &col1)
	col2 := shim.Column{Value: &shim.Column_String_{String_: kidner.DonneurHash}}
	columns = append(columns, &col2)
	col3 := shim.Column{Value: &shim.Column_String_{String_: kidner.ReceveurID}}
	columns = append(columns, &col3)
	col4 := shim.Column{Value: &shim.Column_String_{String_: kidner.ReceveurHash}}
	columns = append(columns, &col4)
	state := shim.Column{Value: &shim.Column_String_{String_: kidner.State}}
	columns = append(columns, &state)
	js := shim.Column{Value: &shim.Column_Bytes{Bytes: []byte(buff)}}
	columns = append(columns, &js)
	row := shim.Row{columns}
	logger.Debug("rowString= ", fmt.Sprintf("%s", row))
	return row
}

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

//a changer
func (t *KidnerChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("\n\nQuery() : calling method -")
	var ID string // Entities
	var err error

	if len(args) != 1 {
		logger.Error("Query(): Incorrect number of arguments. Expecting ID of the person")
		return nil, errors.New("Query(): Incorrect number of arguments. Expecting ID of the person")
	}
	ID = args[0]
	aKidner, err := GetKidnerObject(stub, ID)
	if err != nil {
		logger.Error("Query() : Failed to Query Object ")
		jsonResp := "{\"Error\":\"Failed to get Object Data for " + ID + "\"}"
		return nil, errors.New(jsonResp)
	}
	if aKidner.DonneurID == "" {
		logger.Error("Query() : Incorrect Query Object ")
		jsonResp := "{\"Error\":\"Incorrect information about the key for " + ID + "\"}"
		return nil, errors.New(jsonResp)
	}
	buff, err := KidnertoJSON(aKidner)
	if err != nil {
		logger.Error("Query() : KidnertoJSON failed. ")
		jsonResp := "{\"Error\":\"KidnertoJSON failed about the key for " + ID + "\"}"
		return nil, errors.New(jsonResp)
	}
	logger.Debug("Query() : Response : Successfull -")
	return buff, nil
}

func GetKidnerObject(stub shim.ChaincodeStubInterface, args string) (KidnerObject, error) {
	initial := KidnerObject{}
	row, err := GetRowFromKidnerTable(stub, args)
	if err != nil {
		return initial, fmt.Errorf("GetKidnerObject operation failed : %s", err)
	}
	kidner, err := JSONtoKidner(row)
	if err != nil {
		logger.Error("GetKidnerObject() Failed : Ummarshall error")
		return initial, fmt.Errorf("GetKidnerObject() operation failed. %s", err)
	}
	return kidner, nil
}

func GetRowFromKidnerTable(stub shim.ChaincodeStubInterface, args string) ([]byte, error) {
	logger.Debug("GetRowKidnerTable() : calling method -")
	var columns []shim.Column
	col := shim.Column{Value: &shim.Column_String_{String_: args}}
	columns = append(columns, col)
	logger.Debug("getRow: args=", args)
	row, err := stub.GetRow(KIDNERTABLENAME, columns)
	logger.Debug("rowString: ", fmt.Sprintf("%s", row))
	if err != nil {
		logger.Error("Error retrieving data record for Keys = ", args)
		return nil, err
	}
	logger.Debug("Length or number of rows retrieved ", len(row.Columns))

	if len(row.Columns) == 0 {
		jsonResp := "{\"Error\":\"Failed retrieving data " + args + ". \"}"
		logger.Error("Error retrieving data record for Keys = ", args, "Error : ", jsonResp)
		return nil, errors.New(jsonResp)
	}
	Avalbytes := row.Columns[KIDNERCOLOMNBYTES].GetBytes()
	return Avalbytes, nil
}

//a changer
func MiseAJour(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	var ID string
	logger.Debug("\n\nMiseAJour() : calling method -")
	logger.Debug("Args: ID:", args[0])
	ID = args[0]
	if len(args) != 3 {
		logger.Error("MiseAJour(): Incorrect number of arguments. Expecting 3")
		return nil, errors.New("MiseAJour(): Incorrect number of arguments. Expecting 3")
	}
	aKidner, err := GetKidnerObject(stub, ID)
	var bKidner KidnerObject
	bKidner = aKidner
	if args[1] == "DonneurHash" {
		bKidner = KidnerObject{aKidner.DonneurID, args[2], aKidner.ReceveurID, aKidner.ReceveurHash, aKidner.State}
	}
	if args[1] == "ReceveurHash" {
		bKidner = KidnerObject{aKidner.DonneurID, aKidner.DonneurHash, aKidner.ReceveurID, args[2], aKidner.State}
	}
	if args[1] == "State" {
		if args[2] != ACTIF && args[2] != UNACTIF {
			logger.Error("MiseAJour(): Incorrect argument of State. Expecting actif or unactif")
			return nil, errors.New("MiseAJour(): Incorrect argument of State. Expecting actif or unactif")
		}
		bKidner = KidnerObject{aKidner.DonneurID, aKidner.DonneurHash, aKidner.ReceveurID, aKidner.ReceveurHash, args[2]}
	}
	json, err := KidnertoJSON(bKidner)
	if err != nil {
		logger.Error("MiseAJour() : KidnertoJSON failed. ")
		jsonResp := "{\"Error\":\"KidnertoJSON failed about the new key for " + ID + "\"}"
		return nil, errors.New(jsonResp)
	}
	row := BuildRowKidnerTable(bKidner, json)
	ok, err := stub.ReplaceRow(KIDNERTABLENAME, row)
	if err != nil {
		return nil, fmt.Errorf("updateKidnerTable operation failed. %s", err)
	}
	if !ok {
		return nil, errors.New("updateKidnerTable operation failed. Row already exists")
	}
	return []byte("MiseAJour Successfully!"), nil
}

//a changer
func TrouveMatch(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Debug("\n\nTrouveMatch() : calling method -")
	if len(args) != 1 {
		logger.Error("TrouveMatch(): Incorrect number of arguments. Expecting ReceveurID ")
		return nil, errors.New("TrouveMatch(): Incorrect number of arguments. Expecting ReceveurID ")
	}
	logger.Debug("Args: ReceveurID:", args[0])
	var ReceveurID = args[0]
	aKidner, err := GetKidnerObject(stub, ReceveurID)
	newargs := []string{aKidner.DonneurHash, aKidner.ReceveurHash, ACTIF}
	kidners, err := GetKidnerObjects(stub, newargs)
	if err != nil {
		return nil, err
	}
	jsonRows, _ := json.Marshal(kidners)
	return jsonRows, nil
}

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

func GetRowsFromKidnerTable(stub shim.ChaincodeStubInterface, args []string) ([]shim.Row, error) {
	logger.Debug("GetRowsFromKidnerTable() : calling method -")
	var columns []shim.Column
	for i := 0; i < len(args); i++ {
		colNext := shim.Column{Value: &shim.Column_String_{String_: args[i]}}
		columns = append(columns, colNext)
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
