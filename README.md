# loan_plan
1. run the code `go run main.go`
2. make a post request to `http://localhost:8000/generate_plan` with payload `{
"loanAmount": "5000",
"nominalRate": "5.0",
"duration": 24,
"startDate": "2018-01-01T00:00:00Z"
}`
3. check the response with expected value