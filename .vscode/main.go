package main

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

type ErrorFrequency struct {
	Expression string `json:"expression"`
	Endpoint   string `json:"endpoint"`
	Frequency  int    `json:"frequency"`
	Type       string `json:"type"`
}

var errorFrequencyMap map[string]ErrorFrequency

func init() {
	// Initialize the map
	errorFrequencyMap = make(map[string]ErrorFrequency)
}

func incrementErrorFrequency(errMessage, endpoint, expression string) {
	// Check if the error message already exists in the map
	if freq, ok := errorFrequencyMap[errMessage]; ok {
		// If it exists, increment the frequency count
		freq.Frequency++
		// Update the map
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
	// Iterate over error frequency map and append ErrorFrequency objects
	for _, freq := range errorFrequencyMap {
		errorFrequencies = append(errorFrequencies, freq)
	}
	return errorFrequencies
}

var validOperators = []string{"plus", "minus", "multiplied", "divided"}

func extractNumberFromExpression(expression string) (int, error) {
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
					incrementErrorFrequency(errMsg, "/extractNumber", expression)
					return 0, fmt.Errorf("division by zero")
				}
			default:
				errMsg := "unsupported operator"
				incrementErrorFrequency(errMsg, "/extractNumber", expression)
				return 0, fmt.Errorf(errMsg)
			}
			// Update the operator for the next operation (if any)
			if i+1 < len(words) {
				nextOperator := words[i+1]
				// Check if the next operator is valid
				if _, err := strconv.Atoi(words[i+2]); err != nil {
					errMsg := "expressions with invalid syntax"
					incrementErrorFrequency(errMsg, "/extractNumber", expression)
					return 0, fmt.Errorf(errMsg)
				}
				operator = nextOperator
				// Skip the next word (operator) in the loop
				i++
			}
			// Set numberFound flag to true
			numberFound = true
		} else if numberFound {
			// If a number has been found previously, stop parsing further words
			break
		}
	}
	// If no number is found, return error
	if !numberFound {
		errMsg := "no number found in the expression"
		incrementErrorFrequency(errMsg, "/extractNumber", expression)
		return 0, fmt.Errorf(errMsg)
	}
	return result, nil
}
func errorHandler(w http.ResponseWriter, err error, endpoint, expression string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	// Increment error frequency count and update errorFrequencyMap
	incrementErrorFrequency(err.Error(), endpoint, expression)

	// Generate JSON response with error details
	errorResponse := ErrorFrequency{
		Expression: expression,
		Endpoint:   endpoint,
		Frequency:  errorFrequencyMap[err.Error()].Frequency,
		Type:       err.Error(),
	}

	// Marshal error response to JSON
	errorJSON, _ := json.Marshal(errorResponse)

	// Write JSON response
	_, _ = w.Write(errorJSON)
}

func getErrorFrequenciesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get error frequencies
	errorFrequencies := getErrorFrequencies()

	// Marshal error frequencies to JSON
	errorJSON, err := json.Marshal(errorFrequencies)
	if err != nil {
		errorHandler(w, err, "N/A", "N/A")
		return
	}

	// Write JSON response
	_, _ = w.Write(errorJSON)
}

// isValidOperator checks if the given word is a valid operator.
func isValidOperator(word string) bool {
	for _, op := range validOperators {
		if op == word {
			return true
		}
	}
	return false
}

// validateExpression validates the syntax of the given expression.
func validateExpression(expression string) error {
	// Split the expression into words
	words := strings.Fields(expression)
	// Flag to check if a number has been found
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
			// Set numberFound flag to true
			numberFound = true
		}
	}
	// If no number is found, return error
	if !numberFound {
		return fmt.Errorf("non-math question")
	}
	return nil
}

func handleRequest(w http.ResponseWriter, req *http.Request, validateOnly bool) {
	// Decode the JSON request body into RequestBody struct
	var requestBody RequestBody
	err := json.NewDecoder(req.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate or extract the expression based on validateOnly flag
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
		number, err := extractNumberFromExpression(requestBody.Expression)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response = strconv.Itoa(number)
	}

	// Create the response body
	responseBody := ResponseBody{
		Response: response,
	}

	// Encode the response body into JSON format
	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}

// extractHandler handles POST requests for extracting numbers from expressions.
func extractHandler(w http.ResponseWriter, req *http.Request) {
	handleRequest(w, req, false)
}

// validateHandler handles POST requests for validating expression syntax.
func validateHandler(w http.ResponseWriter, req *http.Request) {
	handleRequest(w, req, true)
}

func main() {
	// Register the handler for the POST endpoint
	http.HandleFunc("/extractNumber", extractHandler)
	http.HandleFunc("/validateExpression", validateHandler)
	http.HandleFunc("/errorFrequencies", getErrorFrequenciesHandler)
	// Start the HTTP server
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
