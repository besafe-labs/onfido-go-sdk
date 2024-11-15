# Onfido Go SDK

An unofficial Go SDK for interacting with the Onfido API. This SDK was created because Onfido does not provide an official Go SDK for developers.

## Installation

```bash
go get github.com/besafe-labs/onfido-go-sdk
```

## Usage

```go
import "github.com/besafe-labs/onfido-go-sdk"

// Initialize the client
client, err := onfido.NewClient("your-api-token")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Create an applicant
applicant, err := client.CreateApplicant(context.Background(), onfido.CreateApplicantPayload{
    FirstName: "John",
    LastName:  "Doe",
})
```

## Available Resources

### Applicants

- Create applicant
- Retrieve applicant
- List applicants
- Update applicant
- Delete applicant
- Restore deleted applicant

### Workflow Runs

- Create workflow run
- Retrieve workflow run
- List workflow runs
- Retrieve workflow run evidence summary file

## Features

- Automatic retries with configurable retry count and wait time
- Region-specific endpoints (EU, US, CA)
- Pagination support
- Comprehensive error handling
- Context support for cancellation and timeouts

## Configuration Options

```go
// Configure region
client, err := onfido.NewClient(token, onfido.WithRegion(onfido.API_REGION_US))

// Configure retries
client, err := onfido.NewClient(token, onfido.WithRetries(3, 5*time.Second))
```

## Error Handling

The SDK provides detailed error information through the `OnfidoError` struct:

```go
type OnfidoError struct {
    Type    string
    Message string
    Fields  map[string]any
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
