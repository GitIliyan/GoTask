package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type RequestBody struct {
	Expression string `json:"expression"`
}

type ResponseBody struct {
	Response string `json:"response"`
}

func IncrementErrorFrequency(errMessage, endpoint, expression string) {
	// Check if the error message already exists in the map
	if freq, ok := errorFrequencyMap[errMessage]; ok {
		// If it exists, increment the frequency count
		freq.Frequency++
		errorFrequencyMap[errMessage] = freq
	} else {
		// If it doesn't exist, add it to the map
		errorFrequencyMap[errMessage] = ErrorFrequency{
			Expression: expression,
			Endpoint:   endpoint,
			Frequency:  1,
			Type:       errMessage,
		}
	}
}

func getErrorFrequencies() []ErrorFrequency {
	var errorFrequencies []ErrorFrequency
	for _, freq := range errorFrequencyMap {
		errorFrequencies = append(errorFrequencies, freq)
	}
	return errorFrequencies
}

var validOperators = []string{"plus", "minus", "multiplied", "divided"}

func isValidOperator(word string) bool {
	for _, op := range validOperators {
		if op == word {
			return true
		}
	}
	return false
}
func ExtractNumberFromExpression(expression string) (int, error) {
	words := strings.Fields(expression)
	var result int
	operator := "plus"
	var numberFound bool
	for i := 0; i < len(words); i++ {
		word := words[i]
		// Attempt to convert the word to an integer
		num, err := strconv.Atoi(word)
		if err == nil {
			// If conversion is successful, perform the operation with the number
			switch operator {
			case "plus":
				result += num
			case "minus":
				result -= num
			case "multiplied":
				result *= num
			case "divided":
				if num != 0 {
					result /= num
				} else {
					errMsg := "division by zero"
					IncrementErrorFrequency(errMsg, "/extractNumber", expression)
					return 0, fmt.Errorf("division by zero")
				}
			default:
				errMsg := "unsupported operator"
				IncrementErrorFrequency(errMsg, "/extractNumber", expression)
				return 0, fmt.Errorf(errMsg)
			}
			// Update the operator for the next operation (if any)
			if i+1 < len(words) {
				nextOperator := words[i+1]
				// Check if the next operator is valid
				if _, err := strconv.Atoi(words[i+2]); err != nil {
					errMsg := "expressions with invalid syntax"
					IncrementErrorFrequency(errMsg, "/extractNumber", expression)
					return 0, fmt.Errorf(errMsg)
				}
				operator = nextOperator
				// Skip the next word (operator) in the loop
				i++
			}
			numberFound = true
		} else if numberFound {
			// If a number has been found previously, stop parsing further words
			break
		}
	}
	// If no number is found, return error
	if !numberFound {
		errMsg := "no number found in the expression"
		IncrementErrorFrequency(errMsg, "/extractNumber", expression)
		return 0, fmt.Errorf(errMsg)
	}
	return result, nil
}

func validateExpression(expression string) error {
	words := strings.Fields(expression)
	var numberFound bool
	for i := 0; i < len(words); i++ {
		word := words[i]
		// Attempt to convert the word to an integer
		_, err := strconv.Atoi(word)
		if err == nil {
			// If conversion is successful, check if the next word is a valid operator
			if i+1 < len(words) {
				nextWord := words[i+1]
				// Check if the next word is a valid operator
				if _, err := strconv.Atoi(nextWord); err != nil {
					// If the next word is not a number, check if it's a valid operator
					if !isValidOperator(nextWord) {
						return fmt.Errorf("unsupported operations")
					}
				} else {
					return fmt.Errorf("expressions with invalid syntax")
				}
			}
			numberFound = true
		}
	}
	if !numberFound {
		return fmt.Errorf("non-math question")
	}
	return nil
}

func handleRequest(w http.ResponseWriter, req *http.Request, validateOnly bool) {
	var requestBody RequestBody
	err := json.NewDecoder(req.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response string
	if validateOnly {
		fmt.Println("validate only")
		err = validateExpression(requestBody.Expression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response = "Expression is valid"
	} else {
		fmt.Println("not validate only")
		number, err := ExtractNumberFromExpression(requestBody.Expression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response = strconv.Itoa(number)
	}

	responseBody := ResponseBody{
		Response: response,
	}

	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}
