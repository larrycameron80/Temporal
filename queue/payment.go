package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type PinPaymentConfirmation struct {
	TxHash        string `json:"tx_hash"`
	EthAddress    string `json:"eth_address"`
	PaymentNumber string `json:"payment_number"`
	ContentHash   string `json:"content_hash"`
}

type PinPaymentSubmission struct {
	PrivateKey   []byte `json:"private_key"`
	Method       uint8  `json:"method"`
	Number       string `json:"number"`
	ChargeAmount string `json:"charge_amount"`
	// EthAddress string.... this is derived from the ethkey
	ContentHash string   `json:"content_hash"`
	H           [32]byte `json:"h"`
	V           uint8    `json:"v"`
	R           [32]byte `json:"r"`
	S           [32]byte `json:"s"`
	Hash        []byte   `json:"hash"`
	Sig         []byte   `json:"sig"`
	Prefixed    bool     `json:"prefixed"`
}

func ProcessPinPaymentConfirmation(msgs <-chan amqp.Delivery, db *gorm.DB, ipcPath, paymentContractAddress string) error {
	fmt.Println("dialing")
	client, err := ethclient.Dial(ipcPath)
	if err != nil {
		fmt.Println("error dialing", err)
		return err
	}
	fmt.Println("generating payment contract handler")
	contract, err := payments.NewPayments(common.HexToAddress(paymentContractAddress), client)
	if err != nil {
		fmt.Println("error generating payment contract", err)
		return err
	}
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
	for d := range msgs {
		fmt.Println("payment detected")
		ppc := &PinPaymentConfirmation{}
		err = json.Unmarshal(d.Body, ppc)
		if err != nil {
			//TODO handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(ppc.TxHash))
		if err != nil {
			//TODO handle
			fmt.Println(err)
			// could be temporary error, so lets not ack
			continue
		}
		if isPending {
			_, err := bind.WaitMined(context.Background(), client, tx)
			if err != nil {
				//TODO handle
				fmt.Println(err)
				// could be a temporary error so lets not ack
				continue
			}
		}
		numberBig, valid := new(big.Int).SetString(ppc.PaymentNumber, 10)
		if !valid {
			// the message was improperly formatted so its garbagio
			fmt.Println("unable to convert string to big int")
			d.Ack(false)
			continue
		}
		payment, err := contract.Payments(nil, common.HexToAddress(ppc.EthAddress), numberBig)
		if err != nil {
			// TODO handle
			fmt.Println(err)
			// could be a temporary issue, so lets not ack
			continue
		}
		fmt.Printf("Payment struct \n%+v\n", payment)
		// now lets verify that the payment was indeed processed
		if payment.State != uint8(1) {
			// this means the payment wasn't actually confirmed, could be transaction rejection, etc...
			// by getting to this step in the code, it means the transaction has been mined so we need to ack this failure
			fmt.Println("payment unable to be processed, likely due to transaction failure or other contract runtime issue")
			d.Ack(false)
			continue
		}
		// here we have confirmed payment went through, so we can upload the file to our system
		err = manager.Pin(ppc.ContentHash)
		if err != nil {
			// this could be temporary so we wont ack
			fmt.Println(err)
			continue
		}
	}
	return nil
}

func ProcessPinPaymentSubmissions(msgs <-chan amqp.Delivery, db *gorm.DB, ipcPath, paymentContractAddress string) error {
	client, err := ethclient.Dial(ipcPath)
	if err != nil {
		return err
	}
	contract, err := payments.NewPayments(common.HexToAddress(paymentContractAddress), client)
	if err != nil {
		return err
	}
	ppm := models.NewPinPaymentManager(db)
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
	for d := range msgs {
		fmt.Println("delivery detected")
		pps := PinPaymentSubmission{}
		err = json.Unmarshal(d.Body, &pps)
		if err != nil {
			fmt.Println("error unmarshaling", err)
			d.Ack(false)
			continue
		}
		k := keystore.Key{}
		err = k.UnmarshalJSON(pps.PrivateKey)
		if err != nil {
			fmt.Println("error unmarshaling private key", err)
			d.Ack(false)
			continue
		}
		auth := bind.NewKeyedTransactor(k.PrivateKey)
		h := pps.H
		v := pps.V
		r := pps.R
		s := pps.S
		method := pps.Method
		prefixed := pps.Prefixed
		num, valid := new(big.Int).SetString(pps.Number, 10)
		if !valid {
			fmt.Println("unable to convert payment number from string to big int")
			d.Ack(false)
			continue
		}
		amount, valid := new(big.Int).SetString(pps.ChargeAmount, 10)
		if !valid {
			fmt.Println("unable to convert charge amount from string to big int")
			d.Ack(false)
			continue
		}
		auth.GasLimit = 275000
		tx, err := contract.MakePayment(auth, h, v, r, s, num, method, amount, prefixed)
		if err != nil {
			// this could be a temporary error so we wont ack it
			fmt.Println("error making payment", err)
			continue
		}
		fmt.Println("successfully sent payment transaction, waiting for it to be mined")
		_, err = bind.WaitMined(context.Background(), client, tx)
		if err != nil {
			// this could be a temporary error, so we wont ack it
			fmt.Println("error waiting for tx to be mined", err)
			continue
		}
		paymentStruct, err := contract.Payments(nil, auth.From, num)
		if err != nil {
			//TODO: add error handling (msg client via email notifying failure)
			fmt.Println("error retrieving payments", err)
			d.Ack(false)
			continue
		}
		if paymentStruct.State != 1 {
			fmt.Println("error occured while processing payment and the upload will not be processed")
			d.Ack(false)
			continue
		}
		paymentFromDB, err := ppm.FindPaymentByNumberAndAddress(num.String(), auth.From.String())
		if err != nil {
			fmt.Println("erorr reading payment from database", err)
			d.Ack(false)
			continue
		}
		contentHash := paymentFromDB.ContentHash
		err = manager.Pin(contentHash)
		if err != nil {
			fmt.Println("error pinning to IPFS", err)
			d.Ack(false)
			continue
		}
		fmt.Println("Content pinned to IPFS")
		d.Ack(false)
	}
	return nil
}
