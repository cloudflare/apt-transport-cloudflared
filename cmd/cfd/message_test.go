package main

import (
    "testing"
)

func TestNewMessage(t *testing.T) {
    m := NewMessage(404, "Description", nil)
    if m.StatusCode != 404 {
        t.Errorf("Bad Status Code")
    }
    if m.Description != "Description" {
        t.Errorf("Bad Description")
    }
    if len(m.Fields) > 0 {
        t.Errorf("Bad Field Count")
    }
}

