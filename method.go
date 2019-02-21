
package main

import (
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/sha512"
    "bufio"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    //"net/http/httptrace"
    "os"
    //"strconv"
    //"strings"
)

const (
    cfdVersion string = "0.1"
)

type CloudflaredMethod struct {
    Log      *log.Logger
    Client   *http.Client
    mwriter  *MessageWriter
    mreader  *MessageReader
}

// Create a new CloudflaredMethod
func NewCloudflaredMethod(output io.Writer, input *bufio.Reader, logFilename string) (*CloudflaredMethod, error) {
    var logger *log.Logger

    // The Client we use by default is the standard default client
    client := &http.Client{}

    // TODO: Create the logger
    logger = nil
    return &CloudflaredMethod{
        logger,
        client,
        NewMessageWriter(output),
        NewMessageReader(input),
    }, nil
}

// Run the method
func (c *CloudflaredMethod) Run() error {
    c.mwriter.Capabilities(cfdVersion, CapSendConfig | CapSingleInstance)
    mreader := NewMessageReader(bufio.NewReader(os.Stdin))

    // TODO: Just in case, keep a list of URLs that need to be dispatched, but haven't
    for {
        msg, err := mreader.ReadMessage()
        if err != nil {
            if err == io.EOF || err == io.ErrClosedPipe {
                return err
            }
            
            if !(err == io.ErrNoProgress || err == io.ErrShortBuffer) {
                return err
            }
        }

        switch msg.StatusCode {
        case 600: // Acquire URL
            go c.HandleAcquire(msg)
        case 601: // Configuration
            c.ParseConfig(msg)
        default:
            c.mwriter.GeneralFailure("Unhandled Message")
        }
    }

    return nil
}

// Handle the 600 Acquire URI message
// TODO: Figure out what an IMS-Hit indicates, and if that applies to this method
func (cfd *CloudflaredMethod) HandleAcquire(msg *Message) error {
    requestedURL := msg.Fields["URI"]
    filename := msg.Fields["Filename"]
    // TODO: Handle empty URI or Filename
    // This shouldn't happen, but it's best to be absurdly fault tolerant if possible

    uri, err := url.Parse(requestedURL)
    if err != nil {
        cfd.mwriter.FailedURI(requestedURL, "", fmt.Sprintf("URI Parse Failure: %s", err.Error()), false, false)
        return err
    }

    switch uri.Scheme {
    case "cfd+http":
        uri.Scheme = "http"
    case "cfd+https":
        uri.Scheme = "https"
    case "cfd":
        uri.Scheme = "https"
        cfd.mwriter.Warning("URI Scheme 'cfd' should not be used. Defaulting to cfd+https")
    default:
        cfd.mwriter.FailedURI(uri.String(), "", fmt.Sprintf("Invalid URI Scheme: %s", uri.Scheme), false, false)
    }

    // TODO: Get the token from Cloudflared
    // TODO: Let APT know we're getting the thing

    resp, err := cfd.Client.Get(uri.String())
    if err != nil {
        cfd.mwriter.FailedURI(uri.String(), "", err.Error(), false, false)
        return err
    }
    // Handle non-200 responses
    // TODO: Handle other 200 codes
    if resp.StatusCode != 200 {
        cfd.mwriter.FailedURI(uri.String(), "", resp.Status, false, false)
        return fmt.Errorf("GET for %s failed with %s", uri.String(), resp.Status)
    }

    // TODO: Write a Start URI message

    // Close the body at the end of the method
    defer resp.Body.Close()
    // We buffer up to 16kb at a time
    buffer := make([]byte, 1024 * 16)
    
    // We want to compute our different hashes, otherwise Apt will reject the package
    hashMD5 := md5.New()
    hashSHA1 := sha1.New()
    hashSHA256 := sha256.New()
    hashSHA512 := sha512.New()

    // And finally, we need to write to this file
    fp, err := os.Create(filename)
    if err != nil {
        cfd.mwriter.GeneralFailure(fmt.Sprintf("Unable to open file %s", filename))
        return err
    }

    for {
        n, err := resp.Body.Read(buffer)
        if n > 0 {
            // Get a slice to just what was read
            bslice := buffer[:n]
            // Update our hashes
            hashMD5.Write(bslice)
            hashSHA1.Write(bslice)
            hashSHA256.Write(bslice)
            hashSHA512.Write(bslice)
            // Write to the file
            fp.Write(bslice)
        }
        if err != nil {
            if err == io.EOF {
                break;
            }

            cfd.mwriter.GeneralFailure(fmt.Sprintf("Failure while reading response body: %s", err.Error()))
            return err
        }
    }

    cfd.mwriter.FinishURI(uri.String(), filename, "", "", false, false)

    return nil
}

func (cfd *CloudflaredMethod) ParseConfig(msg *Message) error {
    return nil
}
