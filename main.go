package main

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	http.HandleFunc("/generate_plan", generatePlan)
	http.ListenAndServe(":8000", nil)
}

type input struct {
	LoanAmount  float64   `json:"loanAmount,string"`
	NominalRate float64   `json:"nominalRate,string"`
	Duration    int       `json:"duration"`
	StartDate   time.Time `json:"startDate,string"`
}

type output struct {
	BorrowerPaymentAmount         float64 `json:"borrowerPaymentAmount,string"`
	Date                          string  `json:"date"`
	InitialOutstandingPrincipal   float64 `json:"initialOutstandingPrincipal,string"`
	Interest                      float64 `json:"interest,string"`
	Principal                     float64 `json:"principal,string"`
	RemainingOutstandingPrincipal float64 `json:"remainingOutstandingPrincipal,string"`
}

func generatePlan(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var payload input
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Println("input json ummarshal error", err)
		}

		//calculate monthly payment
		monthlyPayment, err := calculateMonthlyPay(payload.NominalRate, payload.LoanAmount, payload.Duration)
		if err != nil {
			log.Println("get monthly payment error", err)
			http.Error(w, "input error", http.StatusBadRequest)
			return
		}

		outputSlice := make([]output, 0, payload.Duration)

		remainingOutstandingPrincipal := payload.LoanAmount
		date := payload.StartDate

		for i := 0; i < payload.Duration; i++ {
			initialOutstandingPrincipal := remainingOutstandingPrincipal
			interest := math.Round((payload.NominalRate/100*30*initialOutstandingPrincipal/360)*100) / 100
			annuity := monthlyPayment
			principal := math.Round((annuity-interest)*100) / 100
			//check for last month annuity
			if principal > initialOutstandingPrincipal {
				principal = initialOutstandingPrincipal
				annuity = principal + interest
			}

			remainingOutstandingPrincipal = math.Round((initialOutstandingPrincipal-principal)*100) / 100

			outputSlice = append(outputSlice, output{
				BorrowerPaymentAmount:         annuity,
				Date:                          date.Format("2006-01-02T15:04:05Z"),
				InitialOutstandingPrincipal:   initialOutstandingPrincipal,
				Interest:                      interest,
				Principal:                     principal,
				RemainingOutstandingPrincipal: remainingOutstandingPrincipal,
			})

			date = date.AddDate(0, 1, 0)
		}

		//respond json
		response, err := json.Marshal(&outputSlice)
		if err != nil {
			log.Println("json marshal error", err)
			http.Error(w, "server response error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	} else {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
	}
}

func calculateMonthlyPay(nominalRate float64, loanAmount float64, duration int) (float64, error) {
	if duration <= 0 || loanAmount <= 0 {
		return .0, errors.New("invalid input")
	}

	r := nominalRate / (12 * 100)
	numerator := r * loanAmount
	denominator := 1 - math.Pow(1+r, float64(-duration))
	return math.Round(numerator/denominator*100) / 100, nil
}
