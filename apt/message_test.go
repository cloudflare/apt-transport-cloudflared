package apt

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	t.Run("NewMessage", testNewMessage)
	t.Run("ReadMessage", testReadMessage)
}

func testNewMessage(t *testing.T) {
	t.Run("No Fields", func(t *testing.T) {
		msg := NewMessage(404, "desc")
		assert.Equal(t, msg.StatusCode, uint64(404))
		assert.Equal(t, msg.Description, "desc")
		assert.Equal(t, len(msg.Fields), 0)
	})
	t.Run("One Field", func(t *testing.T) {
		msg := NewMessage(303, "desc", Field{"a", "b"})
		assert.Equal(t, msg.StatusCode, uint64(303))
		assert.Equal(t, msg.Description, "desc")
		assert.Equal(t, len(msg.Fields), 1)
	})
	t.Run("Multiple Fields", func(t *testing.T) {
		msg := NewMessage(202, "202 desc", Field{"a", "b"}, Field{"c", "d"})
		assert.Equal(t, msg.StatusCode, uint64(202))
		assert.Equal(t, msg.Description, "202 desc")
		assert.Equal(t, len(msg.Fields), 2)
	})
}

func testReadMessage(t *testing.T) {
	t.Run("NoDesc", testReadMessageNoDesc)
	t.Run("EOF", testReadMessageEOF)
	t.Run("BadField", testReadMessageBadField)
	t.Run("BadFieldEOF", testReadMessageBadFieldEOF)
}

func testReadMessageNoDesc(t *testing.T) {
	t.Run("No spaces", func(t *testing.T) {
		input := "600\nFilename: /tmp/blah.deb\n\n"
		mreader := NewMessageReader(bufio.NewReader(strings.NewReader(input)))
		_, err := mreader.ReadMessage()
		assert.Error(t, err, "Expected error in message missing description")
	})
	t.Run("Spaces", func(t *testing.T) {
		input := "600   \nFilename: /tmp/blah.deb\n\n"
		mreader := NewMessageReader(bufio.NewReader(strings.NewReader(input)))
		_, err := mreader.ReadMessage()
		assert.Error(t, err, "Expected error in message missing description")
	})
}

func testReadMessageEOF(t *testing.T) {
	input := strings.NewReader("600 Acquire URI\nFilename: /tmp/blah.deb\n")
	mreader := NewMessageReader(bufio.NewReader(input))

	msg, err := mreader.ReadMessage()
	assert.Error(t, err, "Expected error on unexpected EOF, got nil")
	assert.NotNil(t, msg, "Expected partial message to be returned, got nil")
}

func testReadMessageBadField(t *testing.T) {
	input := strings.NewReader("600 Acquire URI\nFilename\nURI: https://httpbin.org/get\n\n")
	mreader := NewMessageReader(bufio.NewReader(input))

	msg, err := mreader.ReadMessage()
	assert.Error(t, err, "Expected error on bad field format, got nil")
	require.NotNil(t, msg, "Expected partial message to be returned, got nil")
	assert.Equal(t, len(msg.Fields), 1, "Expected partial message to contain fields post error, got none")
}

func testReadMessageBadFieldEOF(t *testing.T) {
	input := strings.NewReader("600 Acquire URI\nFilename\nURI: https://httpbin.org/get\n")
	mreader := NewMessageReader(bufio.NewReader(input))

	msg, err := mreader.ReadMessage()
	assert.Error(t, err, "Expected error on bad field format, got nil")
	assert.NotEqual(t, err, io.EOF, "Expected error group, got io.EOF")
	assert.NotNil(t, msg, "Expected partial message to be returned, got nil")
	if msg != nil {
		assert.Equal(t, len(msg.Fields), 1, "Expected partial message to contain fields post error")
	}

	msg, err = mreader.ReadMessage()
	assert.Equal(t, err, io.EOF)
	assert.Nil(t, msg, "Expected io.EOF, but parsed message")
}

func TestParseHeader(t *testing.T) {
	t.Run("Standard", func(t *testing.T) {
		msg, err := ParseHeader("600 Acquire URI")
		require.NoError(t, err)
		require.NotNil(t, msg)
		assert.Equal(t, msg.StatusCode, uint64(600))
		assert.Equal(t, msg.Description, "Acquire URI")
	})
	t.Run("Extra Space", func(t *testing.T) {
		msg, err := ParseHeader("  500  Desc ")
		require.NoError(t, err)
		require.NotNil(t, msg)
		assert.Equal(t, msg.StatusCode, uint64(500))
		assert.Equal(t, msg.Description, "Desc")
	})
	t.Run("Empty Line", func(t *testing.T) {
		msg, err := ParseHeader("")
		assert.Error(t, err)
		assert.Nil(t, msg)
	})
	t.Run("No Description", func(t *testing.T) {
		msg, err := ParseHeader("600")
		assert.Error(t, err)
		assert.Nil(t, msg)
	})
	t.Run("Empty Description", func(t *testing.T) {
		msg, err := ParseHeader("600    ")
		assert.Error(t, err)
		assert.Nil(t, msg)
	})
	t.Run("Bad Status Code", func(t *testing.T) {
		msg, err := ParseHeader("Hello World")
		assert.Error(t, err)
		assert.Nil(t, msg)
	})
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

	expected := "100 Capabilities\n" +
		"Version: 1.0\nSend-Config: true\nPipeline: true\n" +
		"Single-Instance: true\nLocal-Only: true\nNeeds-Cleanup: true\n" +
		"Removable: true\nAuxRequests: true\n\n"
	assert.Equal(t, expected, out.String())
}

func TestLog(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Log("Hello World")
	expected := "101 Log\nMessage: Hello World\n\n"
	assert.Equal(t, expected, out.String())
}

func TestLogf(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Logf("Hello %s", "World")
	expected := "101 Log\nMessage: Hello World\n\n"
	assert.Equal(t, expected, out.String())
}

func TestStatus(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Status("Hello World")
	expected := "102 Status\nMessage: Hello World\n\n"
	assert.Equal(t, expected, out.String())
}

func TestRedirect(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Redirect("url_a", "url_b", "alt_urls", true)
	expected := "103 Redirect\nURI: url_a\n" +
		"New-URI: url_b\nUsedMirror: true\nAlt-URIs: alt_urls\n\n"
	assert.Equal(t, expected, out.String())
}

func TestWarning(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.Warning("Warning")
	expected := "104 Warning\nMessage: Warning\n\n"
	assert.Equal(t, expected, out.String())
}

func TestStartURI(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.StartURI("url", "resume", 1000, true)
	expected := "200 URI Start\nURI: url\nResume-Point: resume\nSize: 1000\n" +
		"UsedMirror: true\n\n"
	assert.Equal(t, expected, out.String())
}

func TestFinishURI(t *testing.T) {
	t.Run("No Fields", func(t *testing.T) {
		var out strings.Builder
		mwriter := NewMessageWriter(&out)

		mwriter.FinishURI("url", "/tmp/file", "resume", "true", true, true)
		expected := "201 URI Done\nURI: url\nFilename: /tmp/file\n" +
			"Resume-Point: resume\nIMS-Hit: true\nAlt-IMS-Hit: true\n" +
			"UsedMirror: true\n\n"
		assert.Equal(t, expected, out.String())
	})

	t.Run("With Fields", func(t *testing.T) {
		var out strings.Builder
		mwriter := NewMessageWriter(&out)
		mwriter.FinishURI("url", "/tmp/file", "", "", false, false, Field{"a", "b"}, Field{"c", "d"})
		expected := "201 URI Done\nURI: url\nFilename: /tmp/file\na: b\nc: d\n\n"
		assert.Equal(t, expected, out.String())
	})
}

func TestAuxRequest(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.AuxRequest("url", "auxurl", "short", "long", 1000, true)
	expected := "351 Aux Request\nURI: url\nAux-URI: auxurl\nMaximumSize: 1000\n" +
		"Aux-ShortDesc: short\nAux-Description: long\nUsedMirror: true\n\n"
	assert.Equal(t, expected, out.String())
}

func TestFailedURI(t *testing.T) {
	t.Run("Transient Failure", func(t *testing.T) {
		var out strings.Builder
		mwriter := NewMessageWriter(&out)
		mwriter.FailedURI("url", "message", "reason", true, true)
		expected := "400 URI Failure\nURI: url\nTransient-Failure: true\n" +
			"UsedMirror: true\n\n"
		assert.Equal(t, expected, out.String())
	})
	t.Run("Regular Error", func(t *testing.T) {
		var out strings.Builder
		mwriter := NewMessageWriter(&out)
		mwriter.FailedURI("url", "message", "reason", false, false)
		expected := "400 URI Failure\nURI: url\nFailReason: reason\n\n"
		assert.Equal(t, expected, out.String())
	})
	t.Run("No URL", func(t *testing.T) {
		var out strings.Builder
		mwriter := NewMessageWriter(&out)
		mwriter.FailedURI("", "message", "reason", false, false)
		expected := "400 URI Failure\nMessage: message\n\n"
		assert.Equal(t, expected, out.String())
	})
}

func TestGeneralFailure(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.GeneralFailure("fail")
	expected := "401 General Failure\nMessage: fail\n\n"
	assert.Equal(t, expected, out.String())
}

func TestMediaChange(t *testing.T) {
	var out strings.Builder
	mwriter := NewMessageWriter(&out)

	mwriter.MediaChange("media", "drive")
	expected := "403 Media Change\nMedia: media\nDrive: drive\n\n"
	assert.Equal(t, expected, out.String())
}
