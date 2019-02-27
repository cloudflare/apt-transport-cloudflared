
package main

import (
    "context"
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
    "os/exec"
    "strconv"
    //"strings"
    //"time"
)

const (
    cfdVersion string = "0.1"
)

type CloudflaredMethod struct {
    Context  context.Context
    Log      *log.Logger
    Client   *http.Client
    mwriter  *MessageWriter
    mreader  *MessageReader
}

type HeaderEntry struct {
    Key   string
    Value string
}

// Create a new CloudflaredMethod
func NewCloudflaredMethod(output io.Writer, input *bufio.Reader, logFilename string) (*CloudflaredMethod, error) {
    var logger *log.Logger

    // The Client we use by default is the standard default client
    client := &http.Client{}

    // TODO: Only log when needed
    logger = nil
    return &CloudflaredMethod{
        Log: logger,
        Client: client,
        mwriter: NewMessageWriter(output),
        mreader: NewMessageReader(input),
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
                return nil
            }
            
            if !(err == io.ErrNoProgress || err == io.ErrShortBuffer) {
                return err
            }
        }

        switch msg.StatusCode {
        case 600: // Acquire URL
            c.HandleAcquire(msg)
        case 601: // Configuration
            c.ParseConfig(msg)
        default:
            c.mwriter.GeneralFailure("Unhandled Message")
        }
    }

    return nil
}

func (cfd *CloudflaredMethod) GetToken(ctx context.Context, uri *url.URL) ([]HeaderEntry, error) {
    // TODO: Support service tokens
    // Steps:
    //   1. Get the service token directory from the configuration message from Apt (default: ~/.cfd/servicetoken/)
    //   2. Check if the host name given is present in the service token directory
    //   3. Read the file and use that instead of using cloudflared
    // For now though, just login with cloudflared
    path := uri.Scheme + "://" + uri.Host
    cfd.mwriter.Log(fmt.Sprintf("Getting token for %s", path))

    login := exec.CommandContext(ctx, "cloudflared", "access", "login", path)
    // TODO: Display the URL that cloudflared outputs
    err := login.Run()
    if err != nil {
        return nil, err
    }

    cmd := exec.CommandContext(ctx, "cloudflared", "access", "token", "--app", path)
    token, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    cfd.mwriter.Log(fmt.Sprintf("Token fetched: %s", token))

    return []HeaderEntry{HeaderEntry{"cf-access-token", string(token)}}, nil
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
        cfd.mwriter.FailedURI(requestedURL, "", fmt.Sprintf("Invalid URI Scheme: %s", uri.Scheme), false, false)
    }

    // Fetch the token using cloudflared
    // The context is used to cancel the cloudflared commands if they hang too long. Default should be 20 seconds
    // TODO: Make this a header entry that supports both Bearer and Service tokens
    headers, err := cfd.GetToken(context.TODO(), uri)
    if err != nil {
        cfd.mwriter.FailedURI(requestedURL, "", err.Error(), false, false)
        return err
    }

    // TODO: Let APT know we're getting the thing

    // Build our request
    req, err := http.NewRequest("GET", uri.String(), nil)
    if err != nil {
        cfd.mwriter.FailedURI(requestedURL, "", err.Error(), false, false)
        return err
    }

    for _, h := range headers {
        req.Header.Set(h.Key, h.Value)
    }
    
    resp, err := cfd.Client.Do(req)
    if err != nil {
        cfd.mwriter.FailedURI(requestedURL, "", err.Error(), false, false)
        return err
    }
    // Handle non-200 responses
    // TODO: Handle other 200 codes
    if resp.StatusCode != 200 {
        cfd.mwriter.FailedURI(requestedURL, "", resp.Status, false, false)
        return fmt.Errorf("GET for %s failed with %s", uri.String(), resp.Status)
    }

    var size uint64
    // Check for header: Content-Length
    sizeHeader, ok := resp.Header["Content-Length"]
    if ok {
        // Base 10, 64 bits
        size, err = strconv.ParseUint(sizeHeader[0], 10, 64)
        if err != nil {
            log.Printf("Error parsing Content-Length: %s\n", err.Error())
            size = 0
        }
    } else {
        size = 0
    }
    cfd.mwriter.StartURI(requestedURL, "", size, false)

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

    strMD5 := string(hashMD5.Sum(nil))
    strSHA1 := string(hashSHA1.Sum(nil))
    strSHA256 := string(hashSHA256.Sum(nil))
    strSHA512 := string(hashSHA512.Sum(nil))

    cfd.mwriter.FinishURI(requestedURL, filename, "", "", false, false, []string{
        fmt.Sprintf("MD5-Hash: %x", strMD5),
        fmt.Sprintf("MD5Sum-Hash: %x", strMD5),
        fmt.Sprintf("SHA1-Hash: %x", strSHA1),
        fmt.Sprintf("SHA256-Hash: %x", strSHA256),
        fmt.Sprintf("SHA512-Hash: %x", strSHA512),
    })

    return nil
}

func (cfd *CloudflaredMethod) ParseConfig(msg *Message) error {
    return nil
}
