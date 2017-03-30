# Kidner
This is a project cooperated by a group of students from Ecole Polytechnique and a team of research from Orange.

## Introduction
* * *
This project is implemented with Golang, based on the plateform of HyperLedger/Fabric v0.6 and aims to facilitate the exchange of organs, kidney for example.

We assume a scenario as follows: 

A patient named **R1** (reciever 1) is suffering from a failure of kidney. One of his friend named **D1** (donor 1) is willing to donate a kidney to him. Unfortunately, their kidneys are not compatibly. At the same time, another pair of friends or famillies are in the same situation. We name them **R2** and **D2** respectively.

This project will provide a solution to this problem. All this kind of pairs (a donor and a reviever uncompatibly) whose information is recorded in our database will find another pair which is *cross-compatibly* with them. Here *cross-compatibly* means that **R1** is compatibly with **D2** and **R2** is compatibly with **D1**. So that we can establish a connection between these two pairs underline.

## Group Members
* * *
### Group Members form Ecole Polytechnique
- Cheng ZHANG
- Shiwen XIA
- Nikolay IONANOV
- Thibault La BATIDE-ALANORE
- Zhihao PENG

## Implementation
* * *
The implementation of this chaincode is consist of 3 principal functions: **Init()**, **Invoke()** and **Query()**.

- ### Init()
    This function, which is used to deploy a chaincode, creates an empty table in the database. This table has 3 columns: **State**, has 2 possible values *active* and *inactive*, which represents the current state of a user (a donor and a reciever). **CoupleID**, is the unique identity of a user in the database. **JSON**, is a string stocking informations of the user. Information stocked in the JSON string includes **DonorHash**, which is the hashcode of health information of the donor. Similarly, **RecieverHash** for the hashcode of the reciever. **Match**, which shows whether a user is matched with another user (which means the 2 donors and 2 recievers form a 'cross-match').\
    This table is initialised as empty.
- ### Query()
    This function is used to consult or check a user's state and information. In order to specify the user, the **CoupleID** must be passed as argument. Then, this function will return a string of the user's **State**, **DonorHash**, **RecieverHash** and state of **Match**.
- ### Invoke()
    This function is used to modify the contents in the table. More precisely, it creates new users and stock their information in the table as a registry row, modify the information of users already existe, and launches a research of compatibly users with a specific user. To achieve this, this function calls 3 sub-functions:
    1. **CreateNewUser()**: a triplet of strings needs to be passed as argument, in the form of '["CoupleID","DonorHash","RecieverHash"]'. Then this function will create a new user registry. Evidently, the **State** will be initialised as *active* and the **Match** will be initialized as *notmatched*.
    2. **Update()**: a triplet of strings needs to be passed as argument, in the form of '["CoupleID","ColumnName","NewValue"]'. Then this function will write the new value to the corresponding column of the user "CoupleID".
    3. **FindMatch()**: a single string needs to be passed as argument, in the form of '["CoupleID"]'. Then this function will run through the whole table to find all active users which are compatibly with the user "CoupleID", and modify all compatibly users as well as "CoupleID" itself as *matched*.


## Test
* * *
This project runs on the plateform of HyperLedger fabric v0.6. In order to test the performance of its functions, we need `vagrant` and `Oracle VirtualBox` to create virtual machines and simulate different nodes.
On the host machine, in the first terminal:
```
cd $GOPATH/src/github.com/hyperledger/fabric/devenv
vagrant up
```
For the first time, this command will download the image of virtual machines and configure it.
Then, under the same directory,
```
vagrant ssh
```
We will enter into the virtual machine. In the virtual machine, We need to make a new directory `kidner` under `GOPATH/src/github.com`,
```
cd $GOPATH/src/github.com
mkdir kidner
cd kidner
```
Then copy `kidner.go` to this directory and build it with `go build`. If no errors are reported, we can start to deploy the chaincode.

### Deployment
We will stay on the virtual machine.

Firstly, go the directory of `$GOPATH/src/github.com/hyperledger/fabric/build/bin` and start the node.
```
cd $GOPATH/src/github.com/hyperledger/fabric/build/bin
./peer node start
``` 
Keep this terminal open, because we will read running logs in it.

Secondly, use a second terminal and enter into the same virtual machine, go under the same directory and deploy the chaincode.
```
cd $GOPATH/src/github.com/hyperledger/fabric/build/bin
./peer chaincode deploy -p github.com/kidner -c '{"Function":"init","Args":[]}'
```
In the command, `-p` means `path`, `-c` means `constractor`. We call `init` to deploy the chaincode and pass no arguments. We will then receive the returned message in the second terminal:
```
Deploy chaincode: 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be
```
This is the hashcode of our chaincode.
In the first terminal, we will see the transaction ID and no error, which means the chaincode is deployed successfully.

### CreateNewUser
We will create somme fake users with fake information to test other functions.
```
./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"CreateNewUser","Args":["1","a","b"]}'
```
This command creates a user with "**CoupleID** = 1, **DonorHash** = a, **RecieverHash** = b, **State** = active, **Match** = notmatched".

We can check this user by running
```
./peer chaincode query -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"query","Args":["1"]}'
```
And we get
```
Query Result: {"State":"active","CoupleID":"1","DonorHash":"a","RecieverHash":"b","Match":"notmatched"}
```
We cannot create another user with the same identity
```
./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"CreateNewUser","Args":["1","a","b"]}'
```
We get in the first terminal
```
[chaincode] processStream -> ERRO 02d Got error: InsertTableRow operation failed. Row with given key already exists
```

In order to test other functions, we create several fake users with the same command:
```
./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"CreateNewUser","Args":["2","b","a"]}'

./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"CreateNewUser","Args":["3","b","a"]}'

./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"CreateNewUser","Args":["4","c","d"]}'

./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"CreateNewUser","Args":["5","b","a"]}'
```
Of course we can check them with the **query** function.

### Update
We will test the update function by modifying the **DonorHash** of user 5. We check its original information:
```
./peer chaincode query -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"query","Args":["5"]}'

Query Result: {"State":"active","CoupleID":"5","DonorHash":"b","RecieverHash":"a",Match":"notmatched"}
```
Then we call the **Update** function and recheck it:
```
./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"Update","Args":["5","DonorHash","b2"]}'

./peer chaincode query -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"query","Args":["5"]}'

Query Result: {"State":"active","CoupleID":"5","DonorHash":"b2","RecieverHash":"a",Match":"notmatched"}
```
The **DonorHash** is successfully modified. With the same method, we can modify other attributes.


### FindMatch
We created users 2, 3 and 4 intendedly so that 2 and 3 can form a cross-match with 1, while 4 cannot form one. We launche the search of matches for user 1:
```
./peer chaincode invoke -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"FindMatch","Args":["1"]}'
```
Then we recheck states of users from 1 to 4:
```
./peer chaincode query -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"query","Args":["1"]}'

Query Result: {"State":"active","CoupleID":"1","DonorHash":"a","RecieverHash":"b",Match":"matched"}

./peer chaincode query -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"query","Args":["2"]}'

Query Result: {"State":"active","CoupleID":"2","DonorHash":"b","RecieverHash":"a",Match":"matched"}

./peer chaincode query -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"query","Args":["3"]}'

Query Result: {"State":"active","CoupleID":"3","DonorHash":"b","RecieverHash":"a",Match":"matched"}

./peer chaincode query -n 1d630ba1934ed96507bb623472c242deac1d5401233e3284ca3b0f1e1b421bde8357de82492c06238375f5d71e215f8e6f395a7d020e094237e5ef84ff6080be -c '{"Function":"query","Args":["4"]}'

Query Result: {"State":"active","CoupleID":"4","DonorHash":"c","RecieverHash":"d",Match":"notmatched"}
```
Clearly, user 1, 2 and 3 are marked as *matched* while user 4 stays *notmatched*.

### TODO
* * *
1. There is still a bug need to be fixed: When we try to query a user not exists or inactive, we don't get the waited error messsage "CoupleID not exist or inactive", but a "Error handling chaincode support stream: stream error".
2. The **FindMatch** can just mark a user as matched rather than note which users are matched with this one. Maybe we need to create another table to stock all match pairs.




