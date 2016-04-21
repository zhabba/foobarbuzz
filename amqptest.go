// amqptest.go
package main

import (
	"errors"
	"fmt"
	"math"

	"encoding/json"

	"github.com/streadway/amqp"
)

type IncomingMesage struct {
	Sum  float64 `json:"sum"`
	Days int64   `json:"days"`
}

type OutgoingMessage struct {
	Sum      float64 `json:"sum"`
	Days     int64   `json:"days"`
	Interest float64 `json:"interest"`
	TotalSum float64 `json:"totalSum"`
	Token    string  `json:"token"`
}

func main() {
	fmt.Println("Begin connection to AMQP ...")

	connection, err := amqp.Dial("amqp://user:password@host")
	if err != nil {
		panic(err)
	}

	channel, err := connection.Channel()
	if err != nil {
		panic(err)
	}

	deliveries, err := channel.Consume("interest-queue", "", false, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	for d := range deliveries {
		outMessage, err := CalculateInterest(d.Body)
		if err == nil {
			fmt.Printf("%s", "+")
			out, _ := json.Marshal(outMessage)
			channel.Publish("", "solved-interest-queue", false, false, amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "application/json",
				ContentEncoding: "UTF-8",
				Body:            out,
				DeliveryMode:    amqp.Transient,
				Priority:        0,
			},
			)
		} else {
			fmt.Printf("%s", "-")
		}
	}
}

func CalculateInterest(message []byte) (OutgoingMessage, error) {
	var incomingMesage IncomingMesage
	if err := json.Unmarshal(message, &incomingMesage); err != nil {
		panic(err)
	}
	sum := incomingMesage.Sum
	days := incomingMesage.Days
	if days > 0 && sum > 0 {
		var totalInterest float64 = 0.0
		var day int64
		for day = 1; day <= days; day++ {
			var dayInterest float64
			if day%15 == 0 {
				dayInterest = RoundPlus(((sum / 100) * 3), 2)
			} else if day%5 == 0 {
				dayInterest = RoundPlus(((sum / 100) * 2), 2)
			} else if day%3 == 0 {
				dayInterest = RoundPlus(((sum / 100) * 1), 2)
			} else {
				dayInterest = RoundPlus(((sum / 100) * 4), 2)
			}
			totalInterest += dayInterest
		}
		totalSum := sum + totalInterest
		return OutgoingMessage{
			sum,
			days,
			RoundPlus(totalInterest, 2),
			RoundPlus(totalSum, 2),
			"zhabba",
		}, nil
	}
	return OutgoingMessage{}, errors.New("Restricted values in input ...")
}

func Round(f float64) float64 {
	return math.Floor(f + .5)
}

func RoundPlus(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return Round(f*shift) / shift
}
