package main

import (
        "log"

        "github.com/eleven-am/enclave/internal/app"
)

func main() {
        if err := app.Run(); err != nil {
                log.Fatalf("failed to run enclave: %v", err)
        }
}
