package main

import (
    "encoding/json"
    "fmt"
)

type Process struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    // Test 1: nil slice
    var processes1 []Process
    b1, _ := json.Marshal(processes1)
    fmt.Printf("nil slice: %s\n", string(b1))
    
    // Test 2: empty slice with make
    processes2 := make([]Process, 0)
    b2, _ := json.Marshal(processes2)
    fmt.Printf("empty slice with make: %s\n", string(b2))
    
    // Test 3: empty slice literal
    processes3 := []Process{}
    b3, _ := json.Marshal(processes3)
    fmt.Printf("empty slice literal: %s\n", string(b3))
}
