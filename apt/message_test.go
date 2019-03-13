package apt

import (
	"bufio"
	"io"
	"strings"
	"testing"
)

const (
	errOutputMismatch = "Expected output of %s does not match actual output\nExpected:\n\"%v\"\n\nRecieved:\n\"%v\"\n"
)

func expect(t *testing.T, scope, expected, actual string) {
	if expected != actual {
		t.Errorf(errOutputMismatch, scope, expected, actual)
	}
}

func checkMessage(t *testing.T, msg *Message, code uint64, desc string, fieldcount int) {
	if msg.StatusCode != code {
		t.Errorf("Bad Status Code; expected %d, got %d", code, msg.StatusCode)
	}

	if msg.Description != desc {
		t.Errorf("Bad Description; expected %s, got %s", desc, msg.Description)
	}

	if len(msg.Fields) > fieldcount {
		t.Errorf("Bad Field Count; expected %d, got %d", fieldcount, len(msg.Fields))
	}
}

func TestNewMessage(t *testing.T) {
	checkMessage(t, NewMessage(404, "desc"), 404, "desc", 0)
	checkMessage(t, NewMessage(303, "desc", Field{"a", "b"}), 303, "desc", 1)
	checkMessage(t, NewMessage(202, "desc", Field{"a", "b"}, Field{"c", "d"}),
		202, "desc", 2)
}

func TestReadMessage_NoDesc(t *testing.T) {
	input := "600\nFilename: /tmp/blah.deb\n\n"
	mreader := NewMessageReader(bufio.NewReader(strings.NewReader(input)))

	_, err := mreader.ReadMessage()
	if err == nil {
		t.Errorf("Expected error in message missing description")
	}

	input = "600   \nFilename: /tmp/blah.deb\n\n"
	mreader = NewMessageReader(bufio.NewReader(strings.NewReader(input)))
	_, err = mreader.ReadMessage()
	if err == nil {
		t.Errorf("Expected error in message missing description")
	}
}

func TestReadMessage_EOF(t *testing.T) {
	input := strings.NewReader("600 Acquire URI\nFilename: /tmp/blah.deb\n")
	mreader := NewMessageReader(bufio.NewReader(input))

	msg, err := mreader.ReadMessage()
	if err == nil {
		t.Errorf("Expected error on unexpected EOF, got nil")
	}

	if msg == nil {
		t.Errorf("Expected partial message to be returned, got nil")
	}
}

func TestReadMessage_BadField(t *testing.T) {
	input := strings.NewReader("600 Acquire URI\nFilename\nURI: https://httpbin.org/get\n\n")
	mreader := NewMessageReader(bufio.NewReader(input))

	msg, err := mreader.ReadMessage()
	if err == nil {
		t.Errorf("Expected error on bad field format, got nil")
	}

	if msg == nil {
		t.Errorf("Expected partial message to be returned, got nil")
	} else if len(msg.Fields) != 1 {
		t.Errorf("Expected partial message to contain fields post error, got none")
	}
}

func TestReadMessage_BadFieldEOF(t *testing.T) {
	input := strings.NewReader("600 Acquire URI\nFilename\nURI: https://httpbin.org/get\n")
	mreader := NewMessageReader(bufio.NewReader(input))

	msg, err := mreader.ReadMessage()
	if err == nil {
		t.Errorf("Expected error on bad field format, got nil")
	}

	if err == io.EOF {
		t.Errorf("Expected error group, got io.EOF")
	}

	if msg == nil {
		t.Errorf("Expected partial message to be returned, got nil")
	} else if len(msg.Fields) != 1 {
		t.Errorf("Expected partial message to contain fields post error, got none")
	}

	msg, err = mreader.ReadMessage()
	if err == nil {
		t.Errorf("Expected io.EOF, got nil")
	} else if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
	if msg != nil {
		t.Errorf("Expected io.EOF, but parsed message")
	}
}

func testParseHeaderInvalid(t *testing.T, headerstr, errmsg string) {
	_, err := ParseHeader(headerstr)
	if err == nil {
		t.Errorf(errmsg)
	}
}

func testParseHeaderValid(t *testing.T, headerstr string, code uint64, desc string) {
	msg, err := ParseHeader(headerstr)
	if err != nil {
		t.Errorf("Expected no error for valid header line '%s', got %v", headerstr, err)
	}
	if msg == nil {
		t.Errorf("Expected valid message from valid header line '%s', got nil", headerstr)
		return
	}

	if msg.StatusCode != code {
		t.Errorf("Got bad status code from valid header line '%s'; expected %d, got %d",
			headerstr, code, msg.StatusCode)
	}

	if msg.Description != desc {
		t.Errorf("Got bad description from valid header line '%s'; expected '%s', got '%s'",
			headerstr, desc, msg.Description)
	}
}

func TestParseHeader(t *testing.T) {
	// Valid: Standard
	testParseHeaderValid(t, "600 Acquire URI", 600, "Acquire URI")

	// Valid: Extra space
	testParseHeaderValid(t, "  500   Desc  ", 500, "Desc")

	// Error: Empty line
	testParseHeaderInvalid(t, "", "Expected error on empty line, got nil")

	// Error: No description
	testParseHeaderInvalid(t, "600", "Expected error on missing description, got nil")

	// Error: Empty description
	testParseHeaderInvalid(t, "600    ", "Expected error on empty description, got nil")

	// Error: Bad Status Code - non-integer
	testParseHeaderInvalid(t, "Hello World", "Expected error on non-integer status code, got nil")
}

func TestWriteMessage(t *testing.T) {
	expected := "600 Acquire URI\nFilename: /tmp/blah.deb\n\n"
	var out strings.Builder
	mwriter := NewMessageWriter(&out)
	mwriter.WriteMessage(NewMessage(600, "Acquire URI", Field{"Filename", "/tmp/blah.deb"}))

	if out.String() != expected {
		t.Errorf("Writer failed to write message correctly")
	}
}

func TestCapabilities(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Capabilities("1.0", CapSingleInstance|CapPipeline|CapSendConfig|
		CapLocalOnly|CapNeedsCleanup|CapRemovable|CapAuxRequests)

	expect(t, "MessageWriter.Capabilities()", "100 Capabilities\n"+
		"Version: 1.0\nSend-Config: true\nPipeline: true\n"+
		"Single-Instance: true\nLocal-Only: true\nNeeds-Cleanup: true\n"+
		"Removable: true\nAuxRequests: true\n\n", out.String())
}

func TestLog(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Log("Hello World")
	expect(t, "MessageWriter.Log()", "101 Log\nMessage: Hello World\n\n",
		out.String())
}

func TestStatus(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Status("Hello World")
	expect(t, "MessageWriter.Status()", "102 Status\nMessage: Hello World\n\n",
		out.String())
}

func TestRedirect(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Redirect("url_a", "url_b", "alt_urls", true)
	expect(t, "MessageWriter.Redirect()", "103 Redirect\nURI: url_a\n"+
		"New-URI: url_b\nUsedMirror: true\nAlt-URIs: alt_urls\n\n",
		out.String())
}

func TestWarning(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Warning("Warning")
	expect(t, "MessageWriter.Warning()", "104 Warning\nMessage: Warning\n\n",
		out.String())
}

func TestStartURI(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.StartURI("url", "resume", 1000, true)
	expect(t, "MessageWriter.StartURI()",
		"200 URI Start\nURI: url\nResume-Point: resume\nSize: 1000\n"+
			"UsedMirror: true\n\n", out.String())
}

func TestFinishURI(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.FinishURI("url", "/tmp/file", "resume", "true", true, true)
	expect(t, "MessageWriter.FinishURI()",
		"201 URI Done\nURI: url\nFilename: /tmp/file\nResume-Point: resume\n"+
			"IMS-Hit: true\nAlt-IMS-Hit: true\nUsedMirror: true\n\n",
		out.String())

	// Test the extras headers functionality
	out.Reset()
	mwriter.FinishURI("url", "/tmp/file", "", "", false, false, Field{"a", "b"}, Field{"c", "d"})
	expect(t, "MessageWriter.FinishURI()",
		"201 URI Done\nURI: url\nFilename: /tmp/file\na: b\nc: d\n\n",
		out.String())
}

func TestAuxRequest(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.AuxRequest("url", "auxurl", "short", "long", 1000, true)
	expect(t, "MessageWriter.AuxRequest()",
		"351 Aux Request\nURI: url\nAux-URI: auxurl\nMaximumSize: 1000\n"+
			"Aux-ShortDesc: short\nAux-Description: long\nUsedMirror: true\n\n",
		out.String())
}

func TestFailedURI(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	// URL is not empty, so we shouldn't get the message string
	// Transient-Failure is set, so we shouldn't get the fail reason
	mwriter.FailedURI("url", "message", "reason", true, true)
	expect(t, "MessageWriter.FailedURI()",
		"400 URI Failure\nURI: url\nTransient-Failure: true\n"+
			"UsedMirror: true\n\n", out.String())

	out.Reset()
	mwriter.FailedURI("url", "message", "reason", false, false)
	expect(t, "MessageWriter.FailedURI()",
		"400 URI Failure\nURI: url\nFailReason: reason\n\n", out.String())

	out.Reset()
	mwriter.FailedURI("", "message", "reason", false, false)
	expect(t, "MessageWriter.FailedURI()",
		"400 URI Failure\nMessage: message\n\n", out.String())
}

func TestGeneralFailure(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.GeneralFailure("fail")
	expect(t, "MessageWriter.GeneralFailure()",
		"401 General Failure\nMessage: fail\n\n", out.String())
}

func TestMediaChange(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.MediaChange("media", "drive")
	expect(t, "MessageWriter.MediaChange()",
		"403 Media Change\nMedia: media\nDrive: drive\n\n", out.String())
}
